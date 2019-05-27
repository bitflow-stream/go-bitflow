package steps

import (
	"fmt"

	"github.com/bitflow-stream/go-bitflow/bitflow"
	"github.com/bitflow-stream/go-bitflow/script/reg"
	log "github.com/sirupsen/logrus"
)

func RegisterDropErrorsStep(b reg.ProcessorRegistry) {
	b.RegisterStep("drop_errors",
		func(p *bitflow.SamplePipeline, params map[string]interface{}) error {
			logDebug := params["log-debug"].(bool)
			logInfo := params["log-info"].(bool)
			logWarn := params["log-warn"].(bool)
			logError := params["log"].(bool) || !(logDebug || logInfo || logWarn) // Enable by default if no other log level was selected

			p.Add(&DropErrorsProcessor{
				LogError:   logError,
				LogWarning: logWarn,
				LogInfo:    logInfo,
				LogDebug:   logDebug,
			})
			return nil
		},
		"All errors of subsequent processing steps are only logged and not forwarded to the steps before. By default, the errors are logged (can be disabled).").
		Optional("log", reg.Bool(), false).
		Optional("log-debug", reg.Bool(), false).
		Optional("log-info", reg.Bool(), false).
		Optional("log-warn", reg.Bool(), false)
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
