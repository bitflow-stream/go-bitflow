package steps

import (
	"fmt"

	bitflow "github.com/antongulenko/go-bitflow"
	pipeline "github.com/antongulenko/go-bitflow-pipeline"
	"github.com/antongulenko/go-bitflow-pipeline/query"
	"github.com/antongulenko/golib"
)

type BlockingProcessor struct {
	bitflow.NoopProcessor
	block *golib.BoolCondition
	key   string
}

func (p *BlockingProcessor) Sample(sample *bitflow.Sample, header *bitflow.Header) error {
	p.block.Wait()
	return p.NoopProcessor.Sample(sample, header)
}

func (p *BlockingProcessor) String() string {
	return fmt.Sprintf("block (key: %v)", p.key)
}

func (p *BlockingProcessor) Close() {
	p.Release()
	p.NoopProcessor.Close()
}

func (p *BlockingProcessor) Release() {
	p.block.Broadcast()
}

type BlockerList struct {
	Blockers []*BlockingProcessor
}

func (l *BlockerList) ReleaseAll() {
	for _, blocker := range l.Blockers {
		blocker.Release()
	}
}

func (l *BlockerList) Add(blocker *BlockingProcessor) {
	l.Blockers = append(l.Blockers, blocker)
}

type ReleasingProcessor struct {
	bitflow.NoopProcessor
	blockers *BlockerList
	key      string
}

func (p *ReleasingProcessor) Close() {
	p.blockers.ReleaseAll()
	p.NoopProcessor.Close()
}

func (p *ReleasingProcessor) String() string {
	return fmt.Sprintf("release all blocks with key %v", p.key)
}

type BlockManager struct {
	blockers map[string]*BlockerList
}

func NewBlockManager() *BlockManager {
	return &BlockManager{
		blockers: make(map[string]*BlockerList),
	}
}

func (m *BlockManager) GetList(key string) *BlockerList {
	list, ok := m.blockers[key]
	if !ok {
		list = new(BlockerList)
		m.blockers[key] = list
	}
	return list
}

func (m *BlockManager) NewBlocker(key string) *BlockingProcessor {
	blocker := &BlockingProcessor{
		block: golib.NewBoolCondition(),
		key:   key,
	}
	m.GetList(key).Add(blocker)
	return blocker
}

func (m *BlockManager) NewReleaser(key string) *ReleasingProcessor {
	return &ReleasingProcessor{
		blockers: m.GetList(key),
		key:      key,
	}
}

func (m *BlockManager) RegisterBlockingProcessor(b *query.PipelineBuilder) {
	b.RegisterAnalysisParamsErr("block", func(p *pipeline.SamplePipeline, params map[string]string) error {
		if err := AddDecoupleStep(p, params); err != nil {
			return err
		}
		p.Add(m.NewBlocker(params["key"]))
		return nil
	}, "Block further processing of the samples until a release() with the same key is closed. Creates a new goroutine, input buffer size must be specified.", []string{"key", "buf"})
}

func (m *BlockManager) RegisterReleasingProcessor(b *query.PipelineBuilder) {
	b.RegisterAnalysisParams("releaseOnClose", func(p *pipeline.SamplePipeline, params map[string]string) {
		p.Add(m.NewReleaser(params["key"]))
	}, "When this step is closed, release all instances of block() with the same key value", []string{"key"})
}
