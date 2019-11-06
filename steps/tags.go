package steps

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/bitflow-stream/go-bitflow/bitflow"
	"github.com/bitflow-stream/go-bitflow/script/reg"
)

func RegisterTaggingProcessor(b reg.ProcessorRegistry) {
	b.RegisterStep("tags",
		func(p *bitflow.SamplePipeline, params map[string]interface{}) error {
			p.Add(NewTaggingProcessor(params["tags"].(map[string]string)))
			return nil
		},
		"Set the given tags on every sample").
		Required("tags", reg.Map(reg.String()))
}

func NewTaggingProcessor(tags map[string]string) bitflow.SampleProcessor {
	templates := make(map[string]bitflow.TagTemplate, len(tags))
	for key, value := range tags {
		templates[key] = bitflow.TagTemplate{
			Template:     value,
			MissingValue: "",
		}
	}

	return &bitflow.SimpleProcessor{
		Description: fmt.Sprintf("Set tags %v", tags),
		Process: func(sample *bitflow.Sample, header *bitflow.Header) (*bitflow.Sample, *bitflow.Header, error) {
			for key, template := range templates {
				value := template.Resolve(sample)
				sample.SetTag(key, value)
			}
			return sample, header, nil
		},
	}
}

func RegisterTagMapping(b reg.ProcessorRegistry) {
	b.RegisterStep("map-tag",
		func(p *bitflow.SamplePipeline, params map[string]interface{}) error {
			sourceTag := params["from"].(string)
			targetTag := params["to"].(string)
			if targetTag == "" {
				targetTag = sourceTag
			}

			mapping := params["mapping"].(map[string]string)
			mappingFile := params["mapping-file"].(string)
			if len(mapping) == 0 {
				if mappingFile == "" {
					return fmt.Errorf("Either 'mapping' or 'mapping-file' parameter must be defined")
				}
				jsonData, err := ioutil.ReadFile(mappingFile)
				if err != nil {
					return fmt.Errorf("Failed to read file '%v': %v", mappingFile, err)
				}
				mapping = make(map[string]string)
				err = json.Unmarshal(jsonData, &mapping)
				if err != nil {
					return fmt.Errorf("Failed to parse file '%v' as map[string]string: %v", mappingFile, err)
				}
			} else if mappingFile != "" {
				return fmt.Errorf("Cannot define both 'mapping' and 'mapping-file' parameters: %v, %v", mapping, mappingFile)
			}

			p.Add(NewTagMapper(sourceTag, targetTag, mapping))
			return nil
		},
		"Load a lookup table from the parameter or from a file (format: single JSON object with string keys and values). Translate the source tag of each sample through the lookup table and write the result to a target tag.").
		Required("from", reg.String()).
		Optional("mapping", reg.Map(reg.String()), map[string]string{}).
		Optional("mapping-file", reg.String(), "").
		Optional("to", reg.String(), "")
}

func NewTagMapper(sourceTag, targetTag string, mapping map[string]string) bitflow.SampleProcessor {
	tagDescription := "'" + sourceTag + "'"
	if sourceTag != targetTag {
		tagDescription += " to '" + targetTag + "'"
	}

	return &bitflow.SimpleProcessor{
		Description: fmt.Sprintf("Map tag %v based on %v map entries", tagDescription, len(mapping)),
		Process: func(sample *bitflow.Sample, header *bitflow.Header) (*bitflow.Sample, *bitflow.Header, error) {
			sourceValue := sample.Tag(sourceTag)
			targetValue, exists := mapping[sourceValue]
			if !exists {
				targetValue = sourceValue
			}
			sample.SetTag(targetTag, targetValue)
			return sample, header, nil
		},
	}
}
