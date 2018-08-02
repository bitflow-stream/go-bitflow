package script

//TODO
import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type processorRegistryTestSuite struct {
	t *testing.T
	*require.Assertions
}

func TestProcessorRegistry(t *testing.T) {
	suite.Run(t, new(processorRegistryTestSuite))
}

func (suite *processorRegistryTestSuite) T() *testing.T {
	return suite.t
}

func (suite *processorRegistryTestSuite) SetT(t *testing.T) {
	suite.t = t
	suite.Assertions = require.New(t)
}

func (suite *processorRegistryTestSuite) TestGivenRegisteredStep_whenGetStep_returnRegisteredStep() {

}
