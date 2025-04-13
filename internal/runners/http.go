package runners

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type HTTPMetricsID string

// Supported HTTP metrics
const (
	StatusCode   HTTPMetricsID = "status_code"
	ResponseTime HTTPMetricsID = "response_time"
	ResponseSize HTTPMetricsID = "response_size"
)

type HTTPMetrics struct {
	// The name of the metric
	Name string `json:"name" yaml:"name"`
	// The id of the metric
	MetricId HTTPMetricsID `json:"metricId" yaml:"metricId"`
	// The value of the metric
	Value string `json:"value" yaml:"value"`
	// The timestamp of the metric
	Timestamp uint64 `json:"timestamp" yaml:"timestamp"`
}

func (m HTTPMetrics) _bufo_metric() {}
func (m HTTPMetrics) GetName() string {
	return m.Name
}
func (m HTTPMetrics) GetType() string {
	return string(m.MetricId)
}
func (m HTTPMetrics) GetValue() string {
	return m.Value
}
func (m HTTPMetrics) GetTimestamp() uint64 {
	return m.Timestamp
}

type HTTPMethod string

const (
	GET     HTTPMethod = "GET"
	POST    HTTPMethod = "POST"
	PUT     HTTPMethod = "PUT"
	DELETE  HTTPMethod = "DELETE"
	PATCH   HTTPMethod = "PATCH"
	HEAD    HTTPMethod = "HEAD"
	OPTIONS HTTPMethod = "OPTIONS"
	CONNECT HTTPMethod = "CONNECT"
	TRACE   HTTPMethod = "TRACE"
)

type HTTPTaskDefinition struct {
	// The request URL
	URL string `json:"url" yaml:"url"`
	// The request method
	Method HTTPMethod `json:"method" yaml:"method"`
	// The request headers
	Headers map[string]string `json:"headers,omitempty" yaml:"headers,omitempty"`
	// The request body
	Body string `json:"body,omitempty" yaml:"body,omitempty"`
	// The request timeout
	Timeout int `json:"timeout,omitempty" yaml:"timeout,omitempty"`
}

type HTTPTaskType string

const (
	Request HTTPTaskType = "request"
)

type TaskSetRunMode string

const (
	Sequential TaskSetRunMode = "sequential"
)

type CaptureDefinition struct {
	// The name of the capture
	Name string `json:"name" yaml:"name"`
	// The type of the capture
	Type string `json:"type" yaml:"type"`
	// The value of the capture
	Value string `json:"value" yaml:"value"`
	// The regex to use for the capture
	Regex string `json:"regex" yaml:"regex"`
	// The storage name of the captured value
	StorageName string `json:"store" yaml:"store"`
}

type AssertionDefinitions map[HTTPMetricsID]string

type HTTPTask struct {
	// The name of the task
	Name string `json:"name" yaml:"name"`
	// The type of the task
	Type HTTPTaskType `json:"type" yaml:"type"`
	// The asssertions to run
	Assertions AssertionDefinitions `json:"assert" yaml:"assert"`
	// The task definition
	TaskDefinition HTTPTaskDefinition `json:"definition" yaml:"definition"`
	// The metrics to collect
	Metrics []HTTPMetricsID `json:"collect" yaml:"collect"`
	// The data to capture
	DataCapture []CaptureDefinition `json:"capture" yaml:"capture"`
}

type HTTPTaskSet struct {
	// The name of the task set
	Name string `json:"name" yaml:"name"`
	// The mode to run the tasks in
	Mode TaskSetRunMode `json:"mode" yaml:"mode"`
	// The maximum number of runs
	MaxRuns uint `json:"max_runs" yaml:"max_runs"`
	// The tasks to run
	Tasks []HTTPTask `json:"tasks" yaml:"tasks"`
	// The metrics to collect
	Metrics []HTTPMetricsID `json:"metrics" yaml:"metrics"`
}

type HTTPRunner struct {
	Runner `yaml:",inline"`
	// The metrics to collect
	Metrics []HTTPMetricsID `json:"metrics" yaml:"metrics"`
	// The task set to run
	TaskSet HTTPTaskSet `json:"task_set" yaml:"task_set"`
}

type AssertionError struct {
	// The name of the task
	TaskName string `json:"task_name" yaml:"task_name"`
	// The name of the assertion
	ID HTTPMetricsID `json:"id" yaml:"id"`
	// The expected value
	Expected string `json:"expected" yaml:"expected"`
	// The actual value
	Actual string `json:"actual" yaml:"actual"`
	// The error message
	Message string `json:"message" yaml:"message"`
}

func (e *AssertionError) Error() string {
	return fmt.Sprintf("Assertion error in task %s: %s - expected %s, got %s: %s", e.TaskName, e.ID, e.Expected, e.Actual, e.Message)
}

// NewHTTPRunner creates a new HTTPRunner with the given request
func NewHTTPRunner(raw []byte) *HTTPRunner {
	httpRunner := HTTPRunner{}
	if err := yaml.Unmarshal(raw, &httpRunner); err != nil {
		return nil
	}

	return &httpRunner
}

func (r *HTTPRunner) Run(metricsChannel chan MetricsType, errorChannel chan error) error {
	var wg sync.WaitGroup

	wg.Add(r.Parallel)

	for i := range r.Parallel {
		go func(i int) {
			defer wg.Done()
			for _, task := range r.TaskSet.Tasks {
				client := &http.Client{}
				var body io.Reader

				if task.TaskDefinition.Body != "" {
					body = strings.NewReader(task.TaskDefinition.Body)
				}

				req, err := http.NewRequest(string(task.TaskDefinition.Method), task.TaskDefinition.URL, body)
				if err != nil {
					errorChannel <- err
					return
				}

				for key, value := range task.TaskDefinition.Headers {
					req.Header.Set(key, value)
				}

				sendTimestamp := time.Now()

				resp, err := client.Do(req)

				if err != nil {
					errorChannel <- err
					return
				}

				defer resp.Body.Close()

				if len(task.Assertions) > 0 {
					for metricId, value := range task.Assertions {
						switch metricId {
						case StatusCode:
							if strconv.Itoa(resp.StatusCode) != value {
								errorChannel <- &AssertionError{
									TaskName: task.Name,
									ID:       metricId,
									Expected: value,
									Actual:   strconv.Itoa(resp.StatusCode),
									Message:  "status code mismatch",
								}
							}
						}
					}
				}

				for i, metric := range task.Metrics {
					switch metric {
					case StatusCode:
						metricsChannel <- HTTPMetrics{Name: fmt.Sprintf("%s-%s-%d", r.Name, task.Name, i), MetricId: StatusCode, Value: strconv.Itoa(resp.StatusCode), Timestamp: uint64(time.Now().Unix())}
					case ResponseTime:
						metricsChannel <- HTTPMetrics{Name: fmt.Sprintf("%s-%s-%d", r.Name, task.Name, i), MetricId: ResponseTime, Value: strconv.Itoa(int(time.Since(sendTimestamp).Milliseconds())), Timestamp: uint64(time.Now().Unix())}
					case ResponseSize:
						metricsChannel <- HTTPMetrics{Name: fmt.Sprintf("%s-%s-%d", r.Name, task.Name, i), MetricId: ResponseSize, Value: strconv.Itoa(int(resp.ContentLength)), Timestamp: uint64(time.Now().Unix())}
					}
				}
			}
		}(i)
	}

	wg.Wait()
	close(metricsChannel)
	close(errorChannel)

	return nil
}
