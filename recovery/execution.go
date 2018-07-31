package recovery

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/antongulenko/go-bitflow-pipeline/query"
	log "github.com/sirupsen/logrus"
)

type ExecutionEngine interface {
	String() string
	PossibleRecoveries(node string) []string

	// Returns only after recovery completed. Error indicates the recovery has failed to execute, the time is the execution time.
	RunRecovery(node string, recovery string) (time.Duration, error)
}

type MockExecutionEngine struct {
	Recoveries      []string
	AvgRecoveryTime time.Duration
	ErrorPercentage float64 // Should in in [0..1]
}

func NewMockExecution(params map[string]string) (*MockExecutionEngine, error) {
	var err error
	numRecoveries := query.IntParam(params, "num-mock-recoveries", 0, false, &err)
	recoveryErrorPercentage := query.FloatParam(params, "recovery-error-percentage", 0, false, &err)
	avgRecTime := query.DurationParam(params, "avg-recovery-time", 0, false, &err)
	if err != nil {
		return nil, err
	}
	recoveries := make([]string, numRecoveries)
	for i := range recoveries {
		recoveries[i] = fmt.Sprintf("recovery-%v", i)
	}
	return &MockExecutionEngine{
		Recoveries:      recoveries,
		AvgRecoveryTime: avgRecTime,
		ErrorPercentage: recoveryErrorPercentage,
	}, nil
}

func (e *MockExecutionEngine) String() string {
	return fmt.Sprintf("Mock execution engine, recoveries: %v", e.Recoveries)
}

func (e *MockExecutionEngine) PossibleRecoveries(node string) []string {
	return e.Recoveries
}

func (e *MockExecutionEngine) RunRecovery(node string, recovery string) (duration time.Duration, err error) {
	duration = time.Duration((rand.NormFloat64() * float64(e.AvgRecoveryTime)) + float64(e.AvgRecoveryTime))
	if duration < 0 {
		duration = 0
	}
	roll := rand.Float64()
	failed := roll < e.ErrorPercentage
	log.Debugf("Mock execution of recovery '%v' on node '%v': success %v (%.2v >= %.2v), duration %v",
		recovery, node, !failed, roll, e.ErrorPercentage, duration)
	if failed {
		err = fmt.Errorf("Mock execution of recovery '%v' on node '%v' failed (%.2v < %.2v), duration %v",
			recovery, node, roll, e.ErrorPercentage, duration)
	}
	return
}
