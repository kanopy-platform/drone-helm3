package run

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"

	"github.com/pelotech/drone-helm3/internal/env"
)

type ConvertTestSuite struct {
	suite.Suite
	ctrl            *gomock.Controller
	mockCmd         *Mockcmd
	originalCommand func(string, ...string) cmd
}

func (suite *ConvertTestSuite) BeforeTest(_, _ string) {
	suite.ctrl = gomock.NewController(suite.T())
	suite.mockCmd = NewMockcmd(suite.ctrl)

	suite.originalCommand = command
	command = func(path string, args ...string) cmd { return suite.mockCmd }
}

func (suite *ConvertTestSuite) AfterTest(_, _ string) {
	command = suite.originalCommand
}

func TestConvertTestSuite(t *testing.T) {
	suite.Run(t, new(ConvertTestSuite))
}

func (suite *ConvertTestSuite) TestNewConvert() {
	cfg := env.Config{
		DryRun:        true,
		Release:       "myapp",
		Context:       "helm",
		DeleteV2Releases: true,
		TillerLabel: "OWNER=TILLER",
		ReleaseVersionsMax: 10,
	}

	c := NewConvert(cfg, "default", "/root/.kube/config")
	suite.Equal(cfg.Release, c.release)
	suite.Equal(cfg.ReleaseVersionsMax, c.releaseVersionsMax)
	suite.Equal(cfg.DeleteV2Releases, c.deleteV2Releases)
	suite.Equal("default", c.tillerNS)
	suite.Equal("/root/.kube/config", c.kubeConfig)
	suite.Equal(cfg.Context, c.kubeContext)
	suite.Equal(true, c.dryRun)
	suite.NotNil(c.config)
}
