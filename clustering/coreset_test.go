package clustering

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type microclustersTestSuite struct {
	t *testing.T
	*require.Assertions
}

func TestMicroClusters(t *testing.T) {
	suite.Run(t, new(microclustersTestSuite))
}

func (suite *microclustersTestSuite) T() *testing.T {
	return suite.t
}

func (suite *microclustersTestSuite) SetT(t *testing.T) {
	suite.t = t
	suite.Assertions = require.New(t)
}

func (s *microclustersTestSuite) dotest(addSet [][]float64, removeSet [][]float64) {
	s.t.Run(fmt.Sprintf("Added %v, Removed %v", len(addSet), len(removeSet)), func(t *testing.T) {
		ass := assert.New(t)

		x := NewCoreset(len(added[0]))
		for _, add := range addSet {
			x.silentMerge(add)
		}
		for _, rem := range removeSet {
			x.silentSubtractPoint(rem)
		}
		ass.NotPanics(func() {
			x.computeRadius()
		})
	})
}

func (s *microclustersTestSuite) TestManyExampleValues() {

	// This works
	removed = removed[:157]
	for i := 0; i < len(added)+len(removed); i++ {
		if i < len(added) {
			s.dotest(added[:i], nil)
		} else {
			s.dotest(added, removed[:i-len(added)])
		}
	}

	// This works
	x := append(added, added...)
	x = append(x, removed...)
	x = append(added, added...)
	x = append(x, removed...)
	x = append(x, removed...)
	s.dotest(x, x)

	// This fails
	s.dotest(added, removed[:157]) // Starts to fail at 157
}
