package main

import (
	"flag"
	"fmt"
	"os"
	"text/template"

	"github.com/osniantonio/technical-challenges-stress-test/internal/stresser"
)

const (
	headerTempl = `Stress Test tool Benchmarking {{.URL}} (be patient)...`
	usage       = `Usage: stresstest [options]
		Options:
		--url         URL to request.
		--requests    Total of requests to send to URL especified by --url. Default to 10.
		--concurrency Total of concurrent requests. Default to 1.
		--help        Show this help.`
)

func main() {
	help := flag.Bool("help", false, "show usage")
	url := flag.String("url", "", "URL to request")
	total := flag.Int("requests", 10, "total of requests")
	conc := flag.Int("concurrency", 1, "total of concurrent requests")
	flag.Parse()

	if *url == "" || *help {
		fmt.Println(usage)
		os.Exit(0)
	}

	templ := template.Must(template.New("header").Parse(headerTempl))
	templ.Execute(os.Stdout, struct{ URL string }{URL: *url})
	opts := &stresser.Options{
		URL:   *url,
		Total: *total,
		Conc:  *conc,
	}
	strs := stresser.NewStresser(opts)
	strs.Exec()
	strs.Report()
}
