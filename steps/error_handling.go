package steps

import (
	"fmt"

	"github.com/bitflow-stream/go-bitflow"
	"github.com/bitflow-stream/go-bitflow-pipeline"
	"github.com/bitflow-stream/go-bitflow-pipeline/script/reg"
	log "github.com/sirupsen/logrus"
)

func RegisterDropErrorsStep(b reg.ProcessorRegistry) {
	b.RegisterAnalysisParamsErr("drop_errors",
		func(p *pipeline.SamplePipeline, params map[string]string) error {
			var err error
			logDebug := reg.BoolParam(params, "log-debug", false, true, &err)
			logInfo := reg.BoolParam(params, "log-info", false, true, &err)
			logWarn := reg.BoolParam(params, "log-warn", false, true, &err)
			logError := reg.BoolParam(params, "log", !(logDebug || logInfo || logWarn), true, &err) // Enable by default if no other log level was selected
			if err == nil {
				p.Add(&DropErrorsProcessor{
					LogError:   logError,
					LogWarning: logWarn,
					LogInfo:    logInfo,
					LogDebug:   logDebug,
				})
			}
			return err
		},
		"All errors of subsequent processing steps are only logged and not forwarded to the steps before. By default, the errors are logged (can be disabled).", reg.OptionalParams("log", "log-debug", "log-info", "log-warn"))
}

type DropErrorsProcessor struct {
	bitflow.NoopProcessor
	LogError   bool
	LogWarning bool
	LogDebug   bool
	LogInfo    bool
}

func (p *DropErrorsProcessor) String() string {
	return fmt.Sprintf("Drop errors of subsequent steps (error: %v, warn: %v, info: %v, debug: %v)", p.LogError, p.LogWarning, p.LogInfo, p.LogDebug)
}

func (p *DropErrorsProcessor) Sample(sample *bitflow.Sample, header *bitflow.Header) error {
	err := p.NoopProcessor.Sample(sample, header)
	if err != nil {
		if p.LogError {
			log.Errorln("(Dropped error)", err)
		} else if p.LogWarning {
			log.Warnln("(Dropped error)", err)
		} else if p.LogInfo {
			log.Infoln("(Dropped error)", err)
		} else if p.LogDebug {
			log.Debugln("(Dropped error)", err)
		}
	}
	return nil
}
