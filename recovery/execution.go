package recovery

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/antongulenko/go-bitflow-pipeline/query"
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
	RandSeed        int

	Events func(node string, recovery string, success bool, duration time.Duration)

	random *rand.Rand
}

func NewMockExecution(params map[string]string) (*MockExecutionEngine, error) {
	var err error
	numRecoveries := query.IntParam(params, "num-mock-recoveries", 0, false, &err)
	recoveryErrorPercentage := query.FloatParam(params, "recovery-error-percentage", 0, false, &err)
	avgRecTime := query.DurationParam(params, "avg-recovery-time", 0, false, &err)
	randSeed := query.IntParam(params, "rand-seed", 1, true, &err)
	if err != nil {
		return nil, err
	}
	engine := &MockExecutionEngine{
		AvgRecoveryTime: avgRecTime,
		ErrorPercentage: recoveryErrorPercentage,
		RandSeed:        randSeed,
	}
	engine.SetNumRecoveries(numRecoveries)
	return engine, nil
}

func (e *MockExecutionEngine) SetNumRecoveries(num int) {
	e.Recoveries = make([]string, num)
	for i := range e.Recoveries {
		e.Recoveries[i] = fmt.Sprintf("recovery-%v", i)
	}
}

func (e *MockExecutionEngine) String() string {
	return fmt.Sprintf("Mock execution engine (num-recoveries: %v, error-percentage: %v, avg-recovery-time: %v, rand-seed: %v)",
		len(e.Recoveries), e.ErrorPercentage, e.AvgRecoveryTime, e.RandSeed)
}

func (e *MockExecutionEngine) PossibleRecoveries(node string) []string {
	return e.Recoveries
}

func (e *MockExecutionEngine) RunRecovery(node string, recovery string) (duration time.Duration, err error) {
	if e.random == nil {
		e.random = rand.New(rand.NewSource(int64(e.RandSeed)))
	}
	duration = time.Duration((e.random.NormFloat64() * float64(e.AvgRecoveryTime)) + float64(e.AvgRecoveryTime))
	if duration < 0 {
		duration = 0
	}
	roll := e.random.Float64()
	failed := roll < e.ErrorPercentage
	if failed {
		err = fmt.Errorf("Mock execution of recovery '%v' on node '%v' failed (%.2v < %.2v), duration %v",
			recovery, node, roll, e.ErrorPercentage, duration)
	}
	if callback := e.Events; callback != nil {
		callback(node, recovery, !failed, duration)
	}
	return
}
