package runners

import (
	"gopkg.in/yaml.v3"
)

type MetricsType interface {
	GetName() string
	GetType() string
	GetValue() string
	GetTimestamp() uint64
	_bufo_metric()
}

type RunnerType interface {
	Run(chan MetricsType, chan error) error
}

type Runner struct {
	// The name of the runner
	Name string `json:"name" yaml:"name"`
	// The type of the runner (e.g., "http", "shell", etc.)
	Type string `json:"type" yaml:"type"`
	// The amount of Runners to run in parallel
	Parallel int `json:"parallel" yaml:"parallel"`
	// The rate in which to spin up the Runners
	Rate int `json:"rate" yaml:"rate"`
	// The duration of the test in seconds
	Duration int `json:"duration" yaml:"duration"`
}

// NewRunner creates a new Runner with the given name and type
func NewRunner(raw []byte) *Runner {
	runner := Runner{}
	if err := yaml.Unmarshal(raw, &runner); err != nil {
		return nil
	}

	return &runner
}
