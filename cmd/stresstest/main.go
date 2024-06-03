package main

import (
	"flag"
	"fmt"
	"os"
	"text/template"

	"github.com/osniantonio/technical-challenges-stress-test/internal/stresser"
)

const (
	header = `Stress Test tool Benchmarking {{.URL}}
		running...`
	usageTip = `Parameter Entry via CLI
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
	insecure := flag.Bool("insecure", true, "disable certificate verification")
	flag.Parse()

	if *url == "" || *help {
		fmt.Println(usageTip)
		os.Exit(0)
	}

	templ := template.Must(template.New("header").Parse(header))
	templ.Execute(os.Stdout, struct{ URL string }{URL: *url})
	opts := &stresser.Options{
		URL:      *url,
		Total:    *total,
		Conc:     *conc,
		Insecure: *insecure,
	}
	strs := stresser.NewStresser(opts)
	strs.Execute()
	strs.ToReport()
}
