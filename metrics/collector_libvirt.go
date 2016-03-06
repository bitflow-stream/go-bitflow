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

func RegisterLibvirtCollector(connectUri string) {
	RegisterCollector(&LibvirtCollector{ConnectUri: connectUri})
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

// ==================== General collector ====================
type LibvirtCollector struct {
	AbstractCollector
	ConnectUri string
	conn       libvirt.VirConnection
	domains    map[string]libvirt.VirDomain
	vmReaders  []*vmMetricsReader
}

func (col *LibvirtCollector) Init() error {
	col.Reset(col)
	col.domains = make(map[string]libvirt.VirDomain)
	if err := col.update(false); err != nil {
		return err
	}
	col.readers = make(map[string]MetricReader)
	col.vmReaders = make([]*vmMetricsReader, 0, len(col.domains))
	for name, _ := range col.domains {
		vmReader := &vmMetricsReader{
			col:  col,
			name: name,
		}
		col.vmReaders = append(col.vmReaders, vmReader)
		col.readers["libvirt/vm/"+name+"/swap"] = vmReader.readSwap
		col.readers["libvirt/vm/"+name+"/available"] = vmReader.readAvailable
		col.readers["libvirt/vm/"+name+"/balloon"] = vmReader.readBalloon
		col.readers["libvirt/vm/"+name+"/rss"] = vmReader.readRss
	}
	return nil
}

func (col *LibvirtCollector) Update() (err error) {
	if err = col.update(true); err != nil {
		if err = col.updateVms(); err != nil {
			col.UpdateMetrics()
		}
	}
	return
}

func (col *LibvirtCollector) update(checkChange bool) error {
	conn, err := libvirt.NewVirConnection(col.ConnectUri)
	if err != nil {
		return err
	}
	col.conn = conn
	domains, err := conn.ListAllDomains(0) // No flags: return all domains
	if err != nil {
		return err
	}
	if checkChange && len(col.domains) != len(domains) {
		return MetricsChanged
	}
	for _, domain := range domains {
		if name, err := domain.GetName(); err != nil {
			return err
		} else {
			if checkChange {
				if _, ok := col.domains[name]; !ok {
					return MetricsChanged
				}
			}
			col.domains[name] = domain
		}
	}
	return nil
}

func (col *LibvirtCollector) updateVms() error {
	for _, vmReader := range col.vmReaders {
		if err := vmReader.update(); err != nil {
			return err
		}
	}
	return nil
}

// ==================== Metrics ====================
type vmMetricsReader struct {
	col      *LibvirtCollector
	name     string
	memStats []libvirt.VirDomainMemoryStat
}

func (reader *vmMetricsReader) update() error {
	if domain, ok := reader.col.domains[reader.name]; !ok {
		return fmt.Errorf("Warning: libvirt domain %v not found", reader.name)
	} else {
		var err error
		if reader.memStats, err = domain.MemoryStats(MAX_NUM_MEMORY_STATS, 0); err != nil {
			return err
		}
	}
	return nil
}

func (reader *vmMetricsReader) readMemStat(index int32) Value {
	// TODO short linear search for every single metric...
	for _, stat := range reader.memStats {
		if stat.Tag == index {
			return Value(stat.Val)
		}
	}
	return Value(-1)
}

func (reader *vmMetricsReader) readSwap() Value {
	return reader.readMemStat(VIR_DOMAIN_MEMORY_STAT_SWAP_OUT)
}

func (reader *vmMetricsReader) readAvailable() Value {
	return reader.readMemStat(VIR_DOMAIN_MEMORY_STAT_AVAILABLE)
}

func (reader *vmMetricsReader) readBalloon() Value {
	return reader.readMemStat(VIR_DOMAIN_MEMORY_STAT_ACTUAL_BALLOON)
}

func (reader *vmMetricsReader) readRss() Value {
	return reader.readMemStat(VIR_DOMAIN_MEMORY_STAT_RSS)
}
