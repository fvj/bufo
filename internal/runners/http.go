package runners

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"net/http"
	"strings"
	"sync"
	"time"
)

type HTTPRequest struct {
	// The URL to send the request to
	URL string `json:"url" yaml:"url"`
	// The HTTP method to use (GET, POST, PUT, DELETE, etc.)
	Method string `json:"method" yaml:"method"`
	// The headers to include in the request
	Headers map[string]string `json:"headers" yaml:"headers"`
	// The body of the request
	Body string `json:"body" yaml:"body"`
	// The timeout for the request in seconds
	Timeout int `json:"timeout" yaml:"timeout"`
}

type HTTPMetrics struct {
	// The id of the metric
	ID string `json:"id" yaml:"id"`
	// The status code of the response
	StatusCode int `json:"status_code" yaml:"status_code"`
	// The response time in milliseconds
	ResponseTime int `json:"response_time" yaml:"response_time"`
	// The size of the response in bytes
	ResponseSize int `json:"response_size" yaml:"response_size"`
}

type HTTPRunner struct {
	Runner `yaml:",inline"`
	// The request to send
	Request HTTPRequest `json:"request" yaml:"request"`
}

// NewHTTPRunner creates a new HTTPRunner with the given request
func NewHTTPRunner(raw []byte) *HTTPRunner {
	httpRunner := &HTTPRunner{}
	if err := yaml.Unmarshal(raw, &httpRunner); err != nil {
		return nil
	}

	return httpRunner
}

func (r *HTTPRunner) Run() error {
	metricsChannel := make(chan HTTPMetrics)
	errorChannel := make(chan error)

	var wg sync.WaitGroup
	wg.Add(r.Parallel)

	for i := range r.Parallel {
		go func() {
			defer wg.Done()
			// convert the body into a reader
			bodyReader := strings.NewReader(r.Request.Body)
			// create a timestamp to calculate response time
			timestamp := time.Now()
			req, err := http.NewRequest(r.Request.Method, r.Request.URL, bodyReader)
			if err != nil {
				errorChannel <- err
				return
			}
			for key, value := range r.Request.Headers {
				req.Header.Set(key, value)
			}
			client := &http.Client{}
			res, err := client.Do(req)
			if err != nil {
				errorChannel <- err
				return
			}
			defer res.Body.Close()
			metricsChannel <- HTTPMetrics{
				ID:           fmt.Sprintf("%s-%d", r.Name, i),
				StatusCode:   res.StatusCode,
				ResponseTime: int(time.Since(timestamp).Milliseconds()),
				ResponseSize: int(res.ContentLength),
			}
		}()
	}

	go func() {
		for response := range metricsChannel {
			fmt.Printf("metric: %+v\n", response)
		}
	}()

	go func() {
		for err := range errorChannel {
			fmt.Printf("error: %+v\n", err)
		}
	}()

	wg.Wait()
	close(metricsChannel)
	close(errorChannel)

	return nil
}
