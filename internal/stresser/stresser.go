package stresser

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"sync"
	"text/template"
	"time"
)

var (
	reportTempl = `
Results:
Concurrency level:                       {{.ConcLevel}}
Total time spent executing:              {{.TimeTaken | printf "%.2f"}} seconds
Total number of requests made:           {{.Total}}
Number of requests with HTTP status 200: {{.Success}}
Errored requests:                        {{.Errors}}
Distribution of other HTTP status codes (such as 404, 500, etc.):
{{range $code, $count := .StatusCodes }}
    {{$code}}: {{$count}}
{{end}}
`
)

type Options struct {
	URL      string
	Total    int
	Conc     int
	Insecure bool
}

type result struct {
	err  error
	code int
}

type report struct {
	ConcLevel   int
	Total       int
	Success     int
	Errors      int
	TimeTaken   float64
	StatusCodes map[int]int
}

type Stresser struct {
	opts         *Options
	wg           *sync.WaitGroup
	reqChan      chan struct{}
	resChan      chan *result
	done         chan struct{}
	statsCounter map[int]int
	errorCount   int
	execTime     time.Duration
}

func NewStresser(opts *Options) *Stresser {
	return &Stresser{
		opts:         opts,
		wg:           &sync.WaitGroup{},
		reqChan:      make(chan struct{}, opts.Total),
		resChan:      make(chan *result, opts.Total),
		done:         make(chan struct{}),
		statsCounter: make(map[int]int),
		errorCount:   0,
	}
}

func (s *Stresser) Execute() {
	start := time.Now()
	s.wg.Add(s.opts.Total)
	go s.beforeRun()
	go s.runRequests()
	go s.getResponse()
	<-s.done
	end := time.Now()
	s.execTime = end.Sub(start)
}

func (s *Stresser) beforeRun() {
	fmt.Println("Preparing requests...")
	s.reqChan <- struct{}{}
	for range s.opts.Total - 1 {
		s.reqChan <- struct{}{}
	}
	close(s.reqChan)
	fmt.Println("Requests prepared.")
}

func (s *Stresser) runRequests() {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	fmt.Println("Sending requests...")
	for {
		_, more := <-s.reqChan
		if !more {
			break
		}
		go func() {
			defer s.wg.Done()
			res, err := http.Get(s.opts.URL)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				s.resChan <- &result{err: err}
				return
			}
			defer res.Body.Close()
			s.resChan <- &result{code: res.StatusCode}
		}()
	}
	fmt.Println("All requests sent.")
}

func (s *Stresser) getResponse() {
	fmt.Println("Receiving responses...")
	for range s.opts.Total {
		res := <-s.resChan
		if res.err != nil {
			s.errorCount++
			continue
		}
		if _, ok := s.statsCounter[res.code]; !ok {
			s.statsCounter[res.code] = 0
		}
		s.statsCounter[res.code]++
	}
	fmt.Println("All responses received.")
	close(s.resChan)
	s.done <- struct{}{}
}

func (s *Stresser) ToReport() {
	success := 0
	otherStatus := make(map[int]int)

	for code, count := range s.statsCounter {
		if code == http.StatusOK {
			success = count
		} else {
			otherStatus[code] = count
		}
	}

	report := &report{
		ConcLevel:   s.opts.Conc,
		Total:       s.opts.Total,
		Success:     success,
		Errors:      s.errorCount,
		TimeTaken:   s.execTime.Seconds(),
		StatusCodes: otherStatus,
	}

	templ := template.Must(template.New("report").Parse(reportTempl))
	templ.Execute(os.Stdout, report)
}

func (s *Stresser) Start() {
	fmt.Println("Starting stress test...")
	go s.getResponse()
	s.Execute()
	<-s.done
	fmt.Println("Stress test completed.")
}
