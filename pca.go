package pipeline

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"

	log "github.com/Sirupsen/logrus"
	"github.com/antongulenko/go-bitflow"
	"github.com/gonum/matrix/mat64"
	"github.com/gonum/stat"
)

const DefaultContainedVariance = 0.99

func SamplesToMatrix(samples []*bitflow.Sample) mat64.Matrix {
	if len(samples) < 1 {
		return mat64.NewDense(0, 0, nil)
	}
	cols := len(samples[0].Values)
	values := make([]float64, len(samples)*cols)
	index := 0
	for _, sample := range samples {
		for _, val := range sample.Values {
			values[index] = float64(val)
			index++
		}
	}
	return mat64.NewDense(len(samples), cols, values)
}

func FillSampleFromMatrix(s *bitflow.Sample, row int, mat *mat64.Dense) {
	FillSample(s, mat.RawRowView(row))
}

func FillSamplesFromMatrix(s []*bitflow.Sample, mat *mat64.Dense) {
	for i, sample := range s {
		FillSampleFromMatrix(sample, i, mat)
	}
}

type PCAModel struct {
	Vectors            *mat64.Dense
	RawVariances       []float64
	ContainedVariances []float64
}

func (model *PCAModel) ComputeModel(samples []*bitflow.Sample) error {
	matrix := SamplesToMatrix(samples)
	pc := new(stat.PC)
	ok := pc.PrincipalComponents(matrix, nil)
	if !ok {
		return errors.New("PCA model could not be computed")
	}
	model.Vectors, model.RawVariances = pc.Vectors(nil), pc.Vars(nil)

	model.ContainedVariances = make([]float64, len(model.RawVariances))
	var sum float64
	for _, variance := range model.RawVariances {
		sum += variance
	}
	for i, variance := range model.RawVariances {
		model.ContainedVariances[i] = variance / sum
	}
	return nil
}

func (model *PCAModel) ComputeAndReport(samples []*bitflow.Sample) error {
	log.Println("Computing PCA model")
	if err := model.ComputeModel(samples); err != nil {
		outErr := fmt.Errorf("Error computing PCA model: %v", err)
		log.Errorln(outErr)
		return outErr
	}
	log.Println(model.Report(DefaultContainedVariance))
	return nil
}

func (model *PCAModel) ComponentsContainingVariance(variance float64) (count int, sum float64) {
	for _, contained := range model.ContainedVariances {
		sum += contained
		count++
		if sum > variance {
			break
		}
	}
	return
}

func (model *PCAModel) String() string {
	return model.Report(DefaultContainedVariance)
}

func (model *PCAModel) Report(reportVariance float64) string {
	totalComponents := len(model.ContainedVariances)
	if model.Vectors == nil || totalComponents == 0 {
		return "PCA model (empty)"
	}
	var buf bytes.Buffer
	num, variance := model.ComponentsContainingVariance(reportVariance)
	fmt.Fprintf(&buf, "PCA model (%v total components, %v components contain %.4f variance): ", totalComponents, num, variance)
	fmt.Fprintf(&buf, "%.4f", model.ContainedVariances[:num])
	return buf.String()
}

func (model *PCAModel) WriteTo(writer io.Writer) error {
	err := gob.NewEncoder(writer).Encode(model)
	if err != nil {
		err = fmt.Errorf("Failed to marshal *PCAModel to binary gob: %v", err)
	}
	return err
}

func (model *PCAModel) Load(filename string) (err error) {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close() // Drop error
	return gob.NewDecoder(file).Decode(model)
}

func (model *PCAModel) Project(numComponents int) *PCAProjection {
	vectors := model.Vectors.View(0, 0, len(model.ContainedVariances), numComponents)
	return &PCAProjection{
		Model:      model,
		Vectors:    vectors,
		Components: numComponents,
	}
}

func (model *PCAModel) ProjectHeader(variance float64, header *bitflow.Header) (*PCAProjection, *bitflow.Header, error) {
	if len(header.Fields) != len(model.ContainedVariances) {
		return nil, nil, fmt.Errorf("Cannot compute PCA projection: PCA model contains %v columns, but samples have %v", len(model.ContainedVariances), len(header.Fields))
	}

	if variance <= 0 {
		variance = DefaultContainedVariance
	}
	comp, variance := model.ComponentsContainingVariance(variance)
	log.Printf("Projecting data into %v components (variance %.4f)...", comp, variance)
	projection := model.Project(comp)

	outFields := make([]string, comp)
	for i := 0; i < comp; i++ {
		outFields[i] = "component" + strconv.Itoa(i)
	}
	return projection, header.Clone(outFields), nil
}

type PCAProjection struct {
	Model      *PCAModel
	Vectors    mat64.Matrix
	Components int
}

func (model *PCAProjection) Matrix(matrix mat64.Matrix) *mat64.Dense {
	var result mat64.Dense
	result.Mul(matrix, model.Vectors)
	return &result
}

func (model *PCAProjection) Vector(vec []float64) []float64 {
	matrix := model.Matrix(mat64.NewDense(1, len(vec), vec))
	return matrix.RawRowView(0)
}

func (model *PCAProjection) Sample(sample *bitflow.Sample) (result *bitflow.Sample) {
	values := model.Vector(SampleToVector(sample))
	FillSample(result, values)
	result.CopyMetadataFrom(sample)
	return
}

func StorePCAModel(filename string) BatchProcessingStep {
	var counter int
	group := bitflow.NewFileGroup(filename)

	return &SimpleBatchProcessingStep{
		Description: fmt.Sprintf("Compute & store PCA model to %v", filename),
		Process: func(header *bitflow.Header, samples []*bitflow.Sample) (*bitflow.Header, []*bitflow.Sample, error) {
			var model PCAModel
			err := model.ComputeAndReport(samples)
			if err == nil {
				var file *os.File
				file, err = group.OpenNewFile(&counter)
				if err == nil {
					defer file.Close() // Drop error
					log.Println("Storing PCA model to", file.Name())
					err = model.WriteTo(file)
				}
			}
			return header, samples, err
		},
	}
}

func LoadBatchPCAModel(filename string, containedVariance float64) (BatchProcessingStep, error) {
	var model PCAModel
	if err := model.Load(filename); err != nil {
		return nil, err
	}

	return &SimpleBatchProcessingStep{
		Description: fmt.Sprintf("Project PCA (model loaded from %v)", filename),
		Process: func(header *bitflow.Header, samples []*bitflow.Sample) (*bitflow.Header, []*bitflow.Sample, error) {
			projection, header, err := model.ProjectHeader(containedVariance, header)
			if err != nil {
				return nil, nil, err
			}

			// Convert sample slice to matrix, do the projection, then fill the new values back into the same sample slice
			// Should minimize allocations, since the value slices have the same length before and after projection
			matrix := projection.Matrix(SamplesToMatrix(samples))
			FillSamplesFromMatrix(samples, matrix)
			return header, samples, nil
		},
	}, nil
}

func LoadStreamingPCAModel(filename string, containedVariance float64) (bitflow.SampleProcessor, error) {
	var (
		model      PCAModel
		checker    bitflow.HeaderChecker
		outHeader  *bitflow.Header
		projection *PCAProjection
	)
	if err := model.Load(filename); err != nil {
		return nil, err
	}

	return &SimpleProcessor{
		Description: fmt.Sprintf("Streaming-project PCA (model loaded from %v)", filename),
		Process: func(sample *bitflow.Sample, header *bitflow.Header) (*bitflow.Sample, *bitflow.Header, error) {
			var err error
			if checker.HeaderChanged(header) {
				projection, outHeader, err = model.ProjectHeader(containedVariance, header)
				if err != nil {
					return nil, nil, err
				}
			}
			if outHeader != nil {
				sample = projection.Sample(sample)
			}
			return sample, outHeader, nil
		},
	}, nil
}

func ComputeAndProjectPCA(containedVariance float64) BatchProcessingStep {
	return &SimpleBatchProcessingStep{
		Description: fmt.Sprintf("Compute & project PCA (%v variance)", containedVariance),
		Process: func(header *bitflow.Header, samples []*bitflow.Sample) (*bitflow.Header, []*bitflow.Sample, error) {
			var model PCAModel
			if err := model.ComputeAndReport(samples); err != nil {
				return nil, nil, err
			}
			projection, header, err := model.ProjectHeader(containedVariance, header)
			if err != nil {
				return nil, nil, err
			}

			// Convert sample slice to matrix, do the projection, then fill the new values back into the same sample slice
			// Should minimize allocations, since the value slices have the same length before and after projection
			matrix := projection.Matrix(SamplesToMatrix(samples))
			FillSamplesFromMatrix(samples, matrix)
			return header, samples, nil
		},
	}
}
