package run

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mongodb-forks/drone-helm3/internal/env"
)

type CheckOutdated struct {
	*config
	chart        string
	chartVersion string
}

func NewCheckOutdated(cfg env.Config) *CheckOutdated {
	return &CheckOutdated{
		config:       newConfig(cfg),
		chart:        cfg.Chart,
		chartVersion: cfg.ChartVersion,
	}
}

func (c *CheckOutdated) Execute() error {
	latest, err := c.latestVersion(c.globalFlags())
	if err != nil {
		return err
	}

	if c.chartVersion != latest {
		fmt.Printf("Chart version %s is outdated, latest version is %s\n", c.chartVersion, latest)
	}

	return nil
}

func (c *CheckOutdated) Prepare() error {
	if !validChartString(c.chart) {
		return fmt.Errorf("invalid chart reference '%s', format must be repo/chartName", c.chart)
	}

	return nil
}

func (c *CheckOutdated) latestVersion(globalArgs []string) (string, error) {
	var latestVersion string
	args := globalArgs
	args = append(args, "search", "repo", c.chart, "-o", "json")
	searchCmd := command(helmBin, args...)

	data, err := searchCmd.Output()
	if err != nil {
		return latestVersion, err
	}

	output := []struct {
		Name        string
		Version     string
		AppVersion  string `json:"app_version"`
		Description string
	}{}

	err = json.Unmarshal(data, &output)
	if err != nil {
		return latestVersion, err
	}

	latestVersion = output[0].Version

	return latestVersion, nil
}

func validChartString(input string) bool {
	split := strings.SplitN(input, "/", 2)
	return len(split) == 2
}
