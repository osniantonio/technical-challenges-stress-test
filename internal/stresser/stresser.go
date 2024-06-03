package stresser

import (
	"fmt"
	"net/http"
	"os"
	"sync"
	"text/tabwriter"
	"text/template"
	"time"
)

var (
	reportTempl = `Results
Concurrency level:          {{.ConcLevel}}
Time taken for tests:       {{.TimeTaken | printf "%.2f"}} seconds
Total of required requests: {{.Total}}
Total of done requests:     {{.Done}}
Errored requests:           {{.Errors}}
Successfully requests:      {{.Success}}
Unsuccessfully requests:

`
)

type Options struct {
	URL   string
	Total int
	Conc  int
}

type result struct {
	err  error
	code int
}

type report struct {
	ConcLevel int
	Total     int
	Done      int
	TimeTaken float64
	Errors    int
	Success   int
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
		reqChan:      make(chan struct{}, opts.Conc),
		resChan:      make(chan *result, opts.Total),
		done:         make(chan struct{}),
		statsCounter: make(map[int]int),
		errorCount:   0,
	}
}

func (s *Stresser) Exec() {
	start := time.Now()
	go s.start()
	go s.execReq()
	go s.saveRes()
	<-s.done
	end := time.Now()
	s.execTime = end.Sub(start)
}

func (s *Stresser) start() {
	s.wg.Add(s.opts.Total)
	for range s.opts.Total {
		s.reqChan <- struct{}{}
	}
	s.wg.Wait()
	close(s.reqChan)
}

func (s *Stresser) execReq() {
	for {
		<-s.reqChan
		go func() {
			defer s.wg.Done()
			res, err := http.Get(s.opts.URL)
			if err != nil {
				s.resChan <- &result{err: err}
				return
			}
			defer res.Body.Close()
			s.resChan <- &result{code: res.StatusCode}
		}()
	}
}

func (s *Stresser) saveRes() {
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
	close(s.resChan)
	s.done <- struct{}{}
}

func (s *Stresser) Report() {
	report := &report{
		ConcLevel: s.opts.Conc,
		Total:     s.opts.Total,
		Done:      s.opts.Total - s.errorCount,
		TimeTaken: s.execTime.Seconds(),
		Errors:    s.errorCount,
		Success:   s.countSuccess(),
	}
	templ := template.Must(template.New("report").Parse(reportTempl))
	templ.Execute(os.Stdout, report)
	s.printFailures()
}

func (s *Stresser) countSuccess() int {
	if value, ok := s.statsCounter[http.StatusOK]; ok {
		return value
	}
	return 0
}

func (s *Stresser) printFailures() {
	w := tabwriter.NewWriter(os.Stdout, 4, 4, 4, ' ', 0)
	fmt.Fprintln(w, "\tCode\tQuantity")
	for k, v := range s.statsCounter {
		if k != http.StatusOK {
			fmt.Fprintf(w, "\t%v\t%v\n", k, v)
		}
	}
	w.Flush()
}
