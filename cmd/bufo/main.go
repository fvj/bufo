package main

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"github.com/akamensky/argparse"
	"github.com/davecgh/go-spew/spew"
	"github.com/fvj/bufo/internal/runners"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"sync"
)

const BUFO_DATABASE = "bufo.db"

//go:embed migrations/*.sql
var migrations embed.FS

func migrateDatabase(dbPath string) {
	os.Remove(dbPath)

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		panic(err)
	}

	files, err := migrations.ReadDir("migrations")
	if err != nil {
		panic(err)
	}

	for _, file := range files {
		migration, err := migrations.ReadFile("migrations/" + file.Name())
		if err != nil {
			panic(err)
		}

		fmt.Println("+ Running migration:", file.Name())
		_, err = db.Exec(string(migration))
		if err != nil {
			panic(err)
		}
	}

	defer db.Close()
}

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

	if _, err := os.Stat(BUFO_DATABASE); errors.Is(err, os.ErrNotExist) {
		migrateDatabase(BUFO_DATABASE)
	}

	// instantiate database
	db, err := sql.Open("sqlite3", BUFO_DATABASE)

	if err != nil {
		panic(err)
	}

	val, err := db.Exec("INSERT INTO run (name, bufo_definition, status) VALUES (?, ?, ?)",
		genericRunner.Name, data, "INIT")

	if err != nil {
		panic(err)
	}

	runId, err := val.LastInsertId()
	if err != nil {
		panic(err)
	}

	defer db.Close()
	metrics := make(chan runners.MetricsType)
	errors := make(chan error)
	var wg sync.WaitGroup

	wg.Add(2)

	go func() {
		defer wg.Done()
		for metric := range metrics {
			spew.Dump(metric)
			if _, err := db.Exec("INSERT INTO metric (run_id, name, value, timestamp) VALUES (?, ?, ?, ?)", runId, metric.GetType(), metric.GetValue(), metric.GetTimestamp()); err != nil {
				panic(err)
			}
		}
	}()

	go func() {
		defer wg.Done()
		for err := range errors {
			spew.Dump(err)
			if _, err := db.Exec("INSERT INTO error (run_id, name, message, timestamp) VALUES (?, ?, ?, ?)", runId, "", err.Error(), 0); err != nil {
				panic(err)
			}
		}
	}()

	if _, err := db.Exec("UPDATE run SET status = ? WHERE id = ?", "RUNNING", runId); err != nil {
		panic(err)
	}

	// run the specific runner
	if err := specificRunner.Run(metrics, errors); err != nil {
		panic(err)
	}

	wg.Wait()

	if _, err := db.Exec("UPDATE run SET status = ?, ended_at = CURRENT_TIMESTAMP WHERE id = ?", "DONE", runId); err != nil {
		panic(err)
	}
}
