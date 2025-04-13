package main

import (
	"fmt"
	"github.com/akamensky/argparse"
	"github.com/davecgh/go-spew/spew"
	"github.com/fvj/bufo/internal/runners"
	"os"
)

func main() {
	parser := argparse.NewParser("bufo", "bufo -- load-test web services")
	run := parser.NewCommand("run", "Run a load test")
	filepath := run.String("f", "file", &argparse.Options{
		Required: true,
		Help:     "Path to the YAML file",
	})
	// parse the command line arguments
	err := parser.Parse(os.Args)
	if err != nil {
		fmt.Println(parser.Usage(err))
		os.Exit(1)
	}
	// read filepath into byte array
	data, err := os.ReadFile(*filepath)
	if err != nil {
		panic(err)
	}

	// create a new general runner
	genericRunner := runners.NewRunner(data)
	if genericRunner == nil {
		panic("failed to create runner")
	}

	// create a new runner based on the type
	var specificRunner runners.RunnerType
	switch genericRunner.Type {
	case "http":
		specificRunner = runners.NewHTTPRunner(data)
	default:
		panic("unknown runner type")
	}

	metrics := make(chan runners.MetricsType)
	errors := make(chan error)

	go func() {
		for metric := range metrics {
			spew.Dump(metric)
		}
	}()

	go func() {
		for err := range errors {
			spew.Dump(err)
		}
	}()

	// run the specific runner
	if err := specificRunner.Run(metrics, errors); err != nil {
		panic(err)
	}
}
