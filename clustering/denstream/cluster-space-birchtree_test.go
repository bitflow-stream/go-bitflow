package denstream

import (
	"testing"

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

func (s *birchTestSuite) TestXXX() {
}
