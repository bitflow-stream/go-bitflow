package math

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/bitflow-stream/go-bitflow/bitflow"
	"github.com/bitflow-stream/go-bitflow/script/reg"
	"github.com/bitflow-stream/go-bitflow/steps"
	log "github.com/sirupsen/logrus"
	"gonum.org/v1/gonum/mat"
	"gonum.org/v1/gonum/stat"
)

const DefaultContainedVariance = 0.99

func SamplesToMatrix(samples []*bitflow.Sample) mat.Matrix {
	if len(samples) < 1 {
		return mat.NewDense(0, 0, nil)
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
	return mat.NewDense(len(samples), cols, values)
}

type PCAModel struct {
	Vectors            *mat.Dense
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
	pc.VarsTo(model.RawVariances)
	pc.VectorsTo(model.Vectors)

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

func (model *PCAModel) WriteModel(writer io.Writer) error {
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
	if err := gob.NewDecoder(file).Decode(model); err != nil {
		return err
	}
	return file.Close()
}

func (model *PCAModel) Project(numComponents int) *PCAProjection {
	vectors := model.Vectors.Slice(0, len(model.ContainedVariances), 0, numComponents)
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
	Vectors    mat.Matrix
	Components int
}

func (model *PCAProjection) Matrix(matrix mat.Matrix) *mat.Dense {
	var result mat.Dense
	result.Mul(matrix, model.Vectors)
	return &result
}

func (model *PCAProjection) Vector(vec []float64) []float64 {
	matrix := model.Matrix(mat.NewDense(1, len(vec), vec))
	return matrix.RawRowView(0)
}

func (model *PCAProjection) Sample(sample *bitflow.Sample) (result *bitflow.Sample) {
	values := model.Vector(steps.SampleToVector(sample))
	result = new(bitflow.Sample)
	steps.FillSample(result, values)
	result.CopyMetadataFrom(sample)
	return
}

func StorePCAModel(filename string) bitflow.BatchProcessingStep {
	var counter int
	group := bitflow.NewFileGroup(filename)

	return &bitflow.SimpleBatchProcessingStep{
		Description: fmt.Sprintf("Compute & store PCA model to %v", filename),
		Process: func(header *bitflow.Header, samples []*bitflow.Sample) (*bitflow.Header, []*bitflow.Sample, error) {
			var model PCAModel
			err := model.ComputeAndReport(samples)
			if err == nil {
				var file *os.File
				file, err = group.OpenNewFile(&counter)
				if err == nil {
					log.Println("Storing PCA model to", file.Name())
					err = model.WriteModel(file)
					if err == nil {
						err = file.Close()
					}
				}
			}
			return header, samples, err
		},
	}
}

func LoadBatchPCAModel(filename string, containedVariance float64) (bitflow.BatchProcessingStep, error) {
	var model PCAModel
	if err := model.Load(filename); err != nil {
		return nil, err
	}

	return &bitflow.SimpleBatchProcessingStep{
		Description: fmt.Sprintf("Project PCA (model loaded from %v)", filename),
		Process: func(header *bitflow.Header, samples []*bitflow.Sample) (*bitflow.Header, []*bitflow.Sample, error) {
			projection, header, err := model.ProjectHeader(containedVariance, header)
			if err != nil {
				return nil, nil, err
			}

			// Convert sample slice to matrix, do the projection, then fill the new values back into the same sample slice
			// Should minimize allocations, since the value slices have the same length before and after projection
			matrix := projection.Matrix(SamplesToMatrix(samples))
			steps.FillSamplesFromMatrix(samples, matrix)
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

	return &bitflow.SimpleProcessor{
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

func ComputeAndProjectPCA(containedVariance float64) bitflow.BatchProcessingStep {
	return &bitflow.SimpleBatchProcessingStep{
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
			steps.FillSamplesFromMatrix(samples, matrix)
			return header, samples, nil
		},
	}
}

func RegisterPCA(b reg.ProcessorRegistry) {
	b.RegisterBatchStep("pca",
		func(params map[string]interface{}) (bitflow.BatchProcessingStep, error) {
			return ComputeAndProjectPCA(params["var"].(float64)), nil
		},
		"Create a PCA model of a batch of samples and project all samples into a number of principal components with a total contained variance given by the parameter").
		Required("var", reg.Float())
}

func RegisterPCAStore(b reg.ProcessorRegistry) {
	b.RegisterBatchStep("pca_store",
		func(params map[string]interface{}) (bitflow.BatchProcessingStep, error) {
			return StorePCAModel(params["file"].(string)), nil
		},
		"Create a PCA model of a batch of samples and store it to the given file").
		Required("file", reg.String())
}

func RegisterPCALoad(b reg.ProcessorRegistry) {
	b.RegisterBatchStep("pca_load",
		func(params map[string]interface{}) (bitflow.BatchProcessingStep, error) {
			return LoadBatchPCAModel(params["file"].(string), params["var"].(float64))
		},
		"Load a PCA model from the given file and project all samples into a number of principal components with a total contained variance given by the parameter").
		Required("var", reg.Float()).
		Required("file", reg.String())
}

func RegisterPCALoadStream(b reg.ProcessorRegistry) {
	create := func(p *bitflow.SamplePipeline, params map[string]interface{}) error {
		step, err := LoadStreamingPCAModel(params["file"].(string), params["var"].(float64))
		if err == nil {
			p.Add(step)
		}
		return err
	}
	b.RegisterStep("pca_load_stream", create,
		"Like pca_load, but process every sample individually, instead of batching them up").
		Required("var", reg.Float()).
		Required("file", reg.String())
}
