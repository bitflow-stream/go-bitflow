package metrics

import (
	"fmt"

	"github.com/rgbkrk/libvirt-go"
)

const (
	// virDomainMemoryStatStruct.Tag
	VIR_DOMAIN_MEMORY_STAT_SWAP_OUT       = 1
	VIR_DOMAIN_MEMORY_STAT_AVAILABLE      = 5 // Max usable memory
	VIR_DOMAIN_MEMORY_STAT_ACTUAL_BALLOON = 6 // Used memory?
	VIR_DOMAIN_MEMORY_STAT_RSS            = 7 // Occuppied by VM process
	MAX_NUM_MEMORY_STATS                  = 8
)

/*
	// Metrics
	v.GetBlockInfo()
	v.InterfaceStats()
	v.BlockStatsFlags()
	v.GetCPUStats()

	// ?
	v.GetInterfaceParameters()
	v.GetInfo()
	v.GetXMLDesc()

	// State
	v.GetState()
	v.IsActive()

	// Info
	v.GetID()
	v.GetMetadata()
	v.GetName()
	v.GetUUID()
	v.GetUUIDString()
	v.GetAutostart()
	v.GetVcpus()
*/

func RegisterLibvirtCollectors(connectUris ...string) {
	for _, uri := range connectUris {
		RegisterCollector(&LibvirtCollector{ConnectUri: uri})
	}
}

func LibvirtSsh(host string, keyfile string) string {
	if keyfile != "" {
		keyfile = "&keyfile=" + keyfile
	}
	return fmt.Sprintf("qemu+ssh://%s/system?no_verify=1%s", host, keyfile)
}

func LibvirtLocal() string {
	return "qemu:///system"
}

type LibvirtCollector struct {
	AbstractCollector
	ConnectUri string
	conn       libvirt.VirConnection
}

func (col *LibvirtCollector) Init() error {
	col.Reset(col)
	col.readers = map[string]MetricReader{
	// TODO add metric readers
	}
	return nil
}

func (col *LibvirtCollector) Update() (err error) {
	if err = col.update(); err != nil {
		col.UpdateMetrics()
	}
	return
}

func (col *LibvirtCollector) update() (err error) {
	col.conn, err = libvirt.NewVirConnection(col.ConnectUri)
	return
}

func (col *LibvirtCollector) readNumVMs() Value {
	a, _ := col.conn.ListAllDomains(0)
	v := a[0]

	v.MemoryStats(MAX_NUM_MEMORY_STATS, 0)

	return Value(0)
}
