package main

import (
	"fmt"
	"net"
	"sort"
	"time"

	. "github.com/antongulenko/data2go/analysis"
	"github.com/antongulenko/data2go/sample"
)

func doHandlePipeline_Prototype(p *sample.CmdSamplePipeline) {
	printer := newTagResultPrinter()
	printer.Hosts = map[string]map[string]string{
		"virtual": map[string]string{
			"192.168.4.177": "bono-1",
			"192.168.4.178": "sprout-1",
			"192.168.4.179": "homestead-1",
			"192.168.4.180": "homer-1",
			"192.168.4.181": "ellis-1",
			"192.168.4.182": "ralf-1",
		},
		"physical": map[string]string{
			"130.149.249.141": "wally131",
			"130.149.249.144": "wally134",
			"130.149.249.145": "wally135",
			"130.149.249.146": "wally136",
			"130.149.249.147": "wally137",
			"130.149.249.148": "wally138",
			"130.149.249.149": "wally139",
			"130.149.249.151": "wally141",
			"130.149.249.152": "wally142",
			"130.149.249.155": "wally145",
			"130.149.249.156": "wally146",
			"130.149.249.157": "wally147",
			"130.149.249.158": "wally148",
			"130.149.249.159": "wally149",
		},
	}
	printer.Layers = []string{"virtual", "physical"}
	p.Add(printer)
}

// =========================== Print results of distributed analysis ===========================

type tagResultPrinter struct {
	AbstractProcessor
	Hosts  map[string]map[string]string // Layer -> IP -> Reverse DNS hostname
	Layers []string

	summary map[string]string
}

func newTagResultPrinter() *tagResultPrinter {
	p := &tagResultPrinter{
		summary: make(map[string]string),
		Hosts:   make(map[string]map[string]string),
	}
	go p.printResults()
	return p
}

func (p *tagResultPrinter) Sample(sample sample.Sample, header sample.Header) error {
	if err := p.CheckSink(); err != nil {
		return err
	}
	if err := sample.Check(header); err != nil {
		return err
	}
	p.handleSample(&sample)
	return p.OutgoingSink.Sample(sample, header)
}

func (p *tagResultPrinter) handleSample(sample *sample.Sample) {
	cls := sample.Tag(ClassTag)
	if cls == "" {
		cls = "(none)"
	}
	if src := sample.Tag(SourceTag); src != "" {
		if ip, _, err := net.SplitHostPort(src); err == nil {
			src = ip
		}
		p.summary[src] = cls
	}
}

func (p *tagResultPrinter) printResults() {
	for {
		summaryCopy := make(map[string]string)
		for key, val := range p.summary {
			summaryCopy[key] = val
		}

		fmt.Println("================================")
		for _, layer := range p.Layers {
			outputs := make([]string, 0, len(p.Hosts[layer]))
			for ip, hostname := range p.Hosts[layer] {
				value, ok := summaryCopy[ip]
				if !ok {
					value = "?"
				} else {
					delete(summaryCopy, ip)
				}
				outputs = append(outputs, fmt.Sprintf("%s = %s", hostname, value))
			}
			p.printLines(outputs, layer)
		}
		if len(summaryCopy) > 0 {
			outputs := make([]string, 0, len(summaryCopy))
			for key, val := range summaryCopy {
				outputs = append(outputs, fmt.Sprintf("%s = %s", key, val))
			}
			p.printLines(outputs, "unknown")
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func (p *tagResultPrinter) printLines(lines []string, layer string) {
	sort.Strings(lines)
	fmt.Printf("== %v hosts ==\n", layer)
	for _, out := range lines {
		fmt.Println(out)
	}
}

func (p *tagResultPrinter) String() string {
	return "tagResultPrinter"
}
