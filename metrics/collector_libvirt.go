package metrics

import (
	"fmt"

	"github.com/rgbkrk/libvirt-go"
)

const (
	NO_FLAGS = 0
)

/*
	// ?
	v.GetInterfaceParameters()
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

// ==================== Libvirt collector ====================
type LibvirtCollector struct {
	AbstractCollector
	ConnectUri string
	conn       libvirt.VirConnection
	domains    map[string]libvirt.VirDomain
	vmReaders  []*vmMetricsCollector
}

func (col *LibvirtCollector) Init() error {
	col.Reset(col)
	col.domains = make(map[string]libvirt.VirDomain)
	if err := col.fetchDomains(false); err != nil {
		return err
	}
	col.readers = make(map[string]MetricReader)
	col.vmReaders = make([]*vmMetricsCollector, 0, len(col.domains))
	for name, _ := range col.domains {
		vmReader := &vmMetricsCollector{
			col:  col,
			name: name,
			readers: []*activatedMetricsReader{
				&activatedMetricsReader{reader: &vmGeneralReader{}},
				&activatedMetricsReader{reader: &memoryStatReader{}},
				&activatedMetricsReader{reader: &cpuStatReader{}},
				&activatedMetricsReader{reader: &blockStatReader{}},
				&activatedMetricsReader{reader: &interfaceStatReader{}},
			},
		}
		for _, collector := range vmReader.readers {
			for metric, reader := range collector.reader.register(name) {
				// The notify mechanism here is to avoid unnecessary libvirt
				// API-calls for metrics that filtered out
				col.readers[metric] = reader
				col.notify[metric] = collector.activate
			}
		}
		col.vmReaders = append(col.vmReaders, vmReader)
	}
	return nil
}

func (col *LibvirtCollector) Update() (err error) {
	if err = col.fetchDomains(true); err == nil {
		if err = col.updateVms(); err == nil {
			col.UpdateMetrics()
		}
	}
	return
}

func (col *LibvirtCollector) fetchDomains(checkChange bool) error {
	conn, err := libvirt.NewVirConnection(col.ConnectUri)
	if err != nil {
		return err
	}
	col.conn = conn
	domains, err := conn.ListAllDomains(NO_FLAGS) // No flags: return all domains
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

// ==================== VM Collector ====================
type vmMetricsCollector struct {
	col     *LibvirtCollector
	name    string
	readers []*activatedMetricsReader
}

func (reader *vmMetricsCollector) update() error {
	if domain, ok := reader.col.domains[reader.name]; !ok {
		return fmt.Errorf("Warning: libvirt domain %v not found", reader.name)
	} else {
		for _, collector := range reader.readers {
			if collector.active {
				if err := collector.reader.update(domain); err != nil {
					return fmt.Errorf("Failed to update domain %s: %v", domain, err)
				}
			}
		}
	}
	return nil
}

type activatedMetricsReader struct {
	active bool
	reader vmMetricsReader
}

func (reader *activatedMetricsReader) activate() {
	reader.active = true
}

type vmMetricsReader interface {
	update(domain libvirt.VirDomain) error
	register(domainName string) map[string]MetricReader
}

// ==================== General VM info ====================
type vmGeneralReader struct {
	info libvirt.VirDomainInfo
}

func (reader *vmGeneralReader) register(domainName string) map[string]MetricReader {
	return map[string]MetricReader{
		"libvirt/vm/" + domainName + "/general/cpu":    reader.readCpu,
		"libvirt/vm/" + domainName + "/general/maxMem": reader.readMaxMem,
		"libvirt/vm/" + domainName + "/general/mem":    reader.readMem,
	}
}

func (reader *vmGeneralReader) update(domain libvirt.VirDomain) (err error) {
	reader.info, err = domain.GetInfo()
	return
}

func (reader *vmGeneralReader) readCpu() Value {
	return Value(reader.info.GetCpuTime())
}

func (reader *vmGeneralReader) readMaxMem() Value {
	return Value(reader.info.GetMaxMem())
}

func (reader *vmGeneralReader) readMem() Value {
	return Value(reader.info.GetMemory())
}

// ==================== Memory Stats ====================
const (
	// virDomainMemoryStatStruct.Tag
	VIR_DOMAIN_MEMORY_STAT_SWAP_OUT       = 1
	VIR_DOMAIN_MEMORY_STAT_AVAILABLE      = 5 // Max usable memory
	VIR_DOMAIN_MEMORY_STAT_ACTUAL_BALLOON = 6 // Used memory?
	VIR_DOMAIN_MEMORY_STAT_RSS            = 7 // Occuppied by VM process
	MAX_NUM_MEMORY_STATS                  = 8
)

type memoryStatReader struct {
	memStats []libvirt.VirDomainMemoryStat
}

func (reader *memoryStatReader) register(domainName string) map[string]MetricReader {
	return map[string]MetricReader{
		"libvirt/vm/" + domainName + "/swap":      reader.readSwap,
		"libvirt/vm/" + domainName + "/available": reader.readAvailable,
		"libvirt/vm/" + domainName + "/balloon":   reader.readBalloon,
		"libvirt/vm/" + domainName + "/rss":       reader.readRss,
	}
}

func (reader *memoryStatReader) update(domain libvirt.VirDomain) (err error) {
	reader.memStats, err = domain.MemoryStats(MAX_NUM_MEMORY_STATS, NO_FLAGS)
	return
}

func (reader *memoryStatReader) readMemStat(index int32) Value {
	// TODO short linear search for every single metric...
	for _, stat := range reader.memStats {
		if stat.Tag == index {
			return Value(stat.Val)
		}
	}
	return Value(-1)
}

func (reader *memoryStatReader) readSwap() Value {
	return reader.readMemStat(VIR_DOMAIN_MEMORY_STAT_SWAP_OUT)
}

func (reader *memoryStatReader) readAvailable() Value {
	return reader.readMemStat(VIR_DOMAIN_MEMORY_STAT_AVAILABLE)
}

func (reader *memoryStatReader) readBalloon() Value {
	return reader.readMemStat(VIR_DOMAIN_MEMORY_STAT_ACTUAL_BALLOON)
}

func (reader *memoryStatReader) readRss() Value {
	return reader.readMemStat(VIR_DOMAIN_MEMORY_STAT_RSS)
}

// ==================== CPU Stats ====================
type cpuStatReader struct {
	stats    *libvirt.VirTypedParameters
	numStats int
}

func (reader *cpuStatReader) register(domainName string) map[string]MetricReader {
	// TODO implement
	return nil
}

func (reader *cpuStatReader) update(domain libvirt.VirDomain) (err error) {
	// TODO Alternative, if GetCPUStats() is not necessary
	// v.GetVcpus() - CPU time per CPU
	_, err = domain.GetCPUStats(reader.stats, reader.numStats, -1, 1, NO_FLAGS)
	return
}

// ==================== Block info ====================
type blockStatReader struct {
}

func (reader *blockStatReader) register(domainName string) map[string]MetricReader {
	// TODO
	return nil
}

func (reader *blockStatReader) update(domain libvirt.VirDomain) (err error) {
	// TODO need to get all block devices from domain.GetXMLDesc()
	// domain.GetBlockInfo()
	// More detailed alternative: domain.BlockStatsFlags()
	return
}

// ==================== Interface info ====================
type interfaceStatReader struct {
}

func (reader *interfaceStatReader) register(domainName string) map[string]MetricReader {
	// TODO
	return nil
}

func (reader *interfaceStatReader) update(domain libvirt.VirDomain) (err error) {
	// More detailed alternative: domain.GetInterfaceParameters()
	//	stats, err := domain.InterfaceStats("paath")
	return
}
