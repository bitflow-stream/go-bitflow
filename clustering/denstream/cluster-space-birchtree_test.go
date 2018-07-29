package denstream

import (
	// log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

type birchTestSuite struct {
	t *testing.T
	*require.Assertions
}

func TestBirchClusterSpace(t *testing.T) {
	suite.Run(t, new(birchTestSuite))
}

func (suite *birchTestSuite) T() *testing.T {
	return suite.t
}

func (suite *birchTestSuite) SetT(t *testing.T) {
	suite.t = t
	suite.Assertions = require.New(t)
}

func (s *birchTestSuite) TestXXX() {

	ass := assert.New(t)

	point := []float64{-0.41527107552819653, -0.14652729027863676}
	clusterSpace := NewBirchTreeClusterSpace(2)

	clusterSpace.NewCluster(point, time.Now())
	ass.Equal(clusterSpace.root.Coreset.W(), 1)
}
