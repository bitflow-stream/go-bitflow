package denstream

import (
	"testing"
	"time"

	"github.com/bitflow-stream/go-bitflow-pipeline/clustering"
	// log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
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

func (s *birchTestSuite) TestInsertRoot() {

	ass := assert.New(s.t)
	clusterSpace := NewBirchTreeClusterSpace(2)

	point := []float64{-0.41527107552819653, -0.14652729027863676}
	point1 := []float64{-0.41527107552819653, -0.14652729027863676}

	clusterSpace.NewCluster(point, time.Now())
	ass.Equal(clusterSpace.root.Coreset.W(), 1.0)

	clusterSpace.NewCluster(point1, time.Now())
	ass.Equal(clusterSpace.root.Coreset.W(), 2.0)

	ass.Equal(clusterSpace.root.numChildren, 2)

}

func (s *birchTestSuite) TestSplitandInsert() {

	ass := assert.New(s.t)
	clusterSpace := NewBirchTreeClusterSpace(2)
	var added = [][]float64{[]float64{-0.41527107552819653, -0.14652729027863676},
		[]float64{-0.376650025446918, -0.1465249275808625},
		[]float64{-0.6067124276495246, -0.14815727290903336},
		[]float64{-0.6129442509054625, -0.12516007473218244}}

	for _, point := range added {
		clusterSpace.NewCluster(point, time.Now())
	}

	ass.Equal(clusterSpace.root.numChildren, 2)
	ass.Equal(clusterSpace.root.Coreset.W(), 4.0)

}

func (s *birchTestSuite) TestNearestCluster() {

	ass := assert.New(s.t)
	clusterSpace := NewBirchTreeClusterSpace(2)
	var added = [][]float64{[]float64{-0.41527107552819653, -0.14652729027863676},
		[]float64{-0.376650025446918, -0.1465249275808625},
		[]float64{-0.6067124276495246, -0.14815727290903336},
		[]float64{-0.6129442509054625, -0.12516007473218244}}

	testpoint := []float64{-0.6035847296436938, 0.10782419498799414}
	for _, point := range added {
		clusterSpace.NewCluster(point, time.Now())
	}

	ass.Equal(clusterSpace.root.numChildren, 2)
	ass.Equal(clusterSpace.root.Coreset.W(), 4.0)

	nearestCluster := clusterSpace.NearestCluster(testpoint)

	testnearestCluster := testpoint
	dist := 0.0
	for _, cluster := range added {
		d := clustering.EuclideanDistance(cluster, testpoint)
		if dist == 0.0 || d < dist {
			testnearestCluster = cluster
		}

	}

	ass.Equal(testnearestCluster, nearestCluster.Center())

}

func (s *birchTestSuite) TestUpdateCluster() {

	ass := assert.New(s.t)
	clusterSpace := NewBirchTreeClusterSpace(2)
	var added = [][]float64{[]float64{-0.41527107552819653, -0.14652729027863676},
		[]float64{-0.376650025446918, -0.1465249275808625},
		[]float64{-0.6067124276495246, -0.14815727290903336},
		[]float64{-0.6129442509054625, -0.12516007473218244}}

	testpoint := []float64{-0.6035847296436938, 0.10782419498799414}
	for _, point := range added {
		clusterSpace.NewCluster(point, time.Now())
	}

	ass.Equal(clusterSpace.root.numChildren, 2)
	ass.Equal(clusterSpace.root.Coreset.W(), 4.0)

	nearestCluster := clusterSpace.NearestCluster(testpoint)

	testnearestCluster := testpoint
	dist := 0.0
	for _, cluster := range added {
		d := clustering.EuclideanDistance(cluster, testpoint)
		if dist == 0.0 || d < dist {
			testnearestCluster = cluster
		}

	}

	ass.Equal(testnearestCluster, nearestCluster.Center())

	clusterSpace.UpdateCluster(nearestCluster, func() bool {
		nearestCluster.Merge(testpoint)
		return true
	})

}
