package main

import (
	"fmt"
	"net"
	"time"

	. "github.com/antongulenko/data2go/analysis"
	"github.com/antongulenko/data2go/sample"
)

func doHandlePipeline_Prototype(p *sample.CmdSamplePipeline) {
	printer := newTagResultPrinter()
	printer.Hosts = map[string]map[string]string{
		"virtual": map[string]string{
			"192.168.4.177": "bono-1",
		},
		"physical": map[string]string{
			"192.168.4.131": "wally131",
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
			fmt.Printf("== %v hosts ==\n", layer)
			for ip, hostname := range p.Hosts[layer] {
				value, ok := summaryCopy[ip]
				if !ok {
					value = "?"
				} else {
					delete(summaryCopy, ip)
				}
				fmt.Println(hostname, "=", value)
			}
		}
		if len(summaryCopy) > 0 {
			fmt.Printf("== %v hosts ==\n", "unknown")
			for key, val := range summaryCopy {
				fmt.Println(key, "=", val)
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func (p *tagResultPrinter) String() string {
	return "tagResultPrinter"
}
