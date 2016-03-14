package metrics

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"gopkg.in/xmlpath.v1"

	"github.com/rgbkrk/libvirt-go"
)

const (
	NO_FLAGS              = 0
	DomainReparseInterval = 5 * time.Minute
)

/*
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
				&activatedMetricsReader{new(vmGeneralReader), false},
				&activatedMetricsReader{new(memoryStatReader), false},
				&activatedMetricsReader{new(cpuStatReader), false},
				&activatedMetricsReader{new(blockStatReader), false},
				&activatedMetricsReader{new(interfaceStatReader), false},
			},
		}
		for _, reader := range vmReader.readers {
			for metric, registeredReader := range reader.register(name) {
				// The notify mechanism here is to avoid unnecessary libvirt
				// API-calls for metrics that are filtered out
				col.readers[metric] = registeredReader
				col.notify[metric] = reader.activate
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

	needXmlDesc   bool
	xmlDescParsed time.Time
}

var UpdateXmlDescription = errors.New("XML domain description must be updated")

func (collector *vmMetricsCollector) update() error {
	if domain, ok := collector.col.domains[collector.name]; !ok {
		return fmt.Errorf("Warning: libvirt domain %v not found", collector.name)
	} else {
		return collector.updateReaders(domain)
	}
}

func (collector *vmMetricsCollector) updateReaders(domain libvirt.VirDomain) error {
	updateDesc := false
	for _, reader := range collector.readers {
		if reader.active {
			if err := reader.update(domain); err == UpdateXmlDescription {
				collector.needXmlDesc = true
				updateDesc = true
			} else if err != nil {
				return fmt.Errorf("Failed to update domain %s: %v", domain, err)
			}
		}
	}
	if time.Now().Sub(collector.xmlDescParsed) >= DomainReparseInterval {
		updateDesc = true
	}
	if collector.needXmlDesc && updateDesc {
		return collector.updateXmlDesc(domain)
	}
	return nil
}

func (collector *vmMetricsCollector) updateXmlDesc(domain libvirt.VirDomain) error {
	xmlData, err := domain.GetXMLDesc(NO_FLAGS)
	if err != nil {
		return fmt.Errorf("Failed to retrieve XML domain description of %s: %v", collector.name, err)
	}
	xmlDesc, err := xmlpath.Parse(strings.NewReader(xmlData))
	if err != nil {
		return fmt.Errorf("Failed to parse XML domain description of %s: %v", collector.name, err)
	}
	collector.xmlDescParsed = time.Now()
	for _, reader := range collector.readers {
		reader.description(xmlDesc)
	}
	return nil
}

type activatedMetricsReader struct {
	vmMetricsReader
	active bool
}

func (reader *activatedMetricsReader) activate() {
	reader.active = true
}

type vmMetricsReader interface {
	register(domainName string) map[string]MetricReader
	description(xmlDesc *xmlpath.Node)
	update(domain libvirt.VirDomain) error
}

// ==================== General VM info ====================
type vmGeneralReader struct {
	info libvirt.VirDomainInfo
}

func (reader *vmGeneralReader) register(domainName string) map[string]MetricReader {
	return map[string]MetricReader{
		"libvirt/" + domainName + "/general/cpu":    reader.readCpu,
		"libvirt/" + domainName + "/general/maxMem": reader.readMaxMem,
		"libvirt/" + domainName + "/general/mem":    reader.readMem,
	}
}

func (reader *vmGeneralReader) description(xmlDesc *xmlpath.Node) {
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
		"libvirt/" + domainName + "/mem/swap":      reader.readSwap,
		"libvirt/" + domainName + "/mem/available": reader.readAvailable,
		"libvirt/" + domainName + "/mem/balloon":   reader.readBalloon,
		"libvirt/" + domainName + "/mem/rss":       reader.readRss,
	}
}

func (reader *memoryStatReader) description(xmlDesc *xmlpath.Node) {
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

func (reader *cpuStatReader) description(xmlDesc *xmlpath.Node) {
}

func (reader *cpuStatReader) update(domain libvirt.VirDomain) (err error) {
	// TODO Alternative, if GetCPUStats() is not necessary
	// v.GetVcpus() - CPU time per CPU
	_, err = domain.GetCPUStats(reader.stats, reader.numStats, -1, 1, NO_FLAGS)
	return
}

// ==================== Block info ====================
type blockStatReader struct {
	parsedDevices bool
	devices       []string
	info          []libvirt.VirDomainBlockInfo
}

func (reader *blockStatReader) register(domainName string) map[string]MetricReader {
	return map[string]MetricReader{
		"libvirt/" + domainName + "/block/allocation": reader.readAllocation,
		"libvirt/" + domainName + "/block/capacity":   reader.readCapacity,
		"libvirt/" + domainName + "/block/physical":   reader.readPhysical,
	}
}

func (reader *blockStatReader) description(xmlDesc *xmlpath.Node) {
	// TODO get all block devices from domain.GetXMLDesc()
	//	reader.parsedDevices = true
}

func (reader *blockStatReader) update(domain libvirt.VirDomain) error {
	reader.info = reader.info[0:0]
	if !reader.parsedDevices {
		return UpdateXmlDescription
	}
	var resErr error
	for _, dev := range reader.devices {
		// More detailed alternative: domain.BlockStatsFlags()
		if info, err := domain.GetBlockInfo(dev, NO_FLAGS); err == nil {
			reader.info = append(reader.info, info)
		} else {
			return fmt.Errorf("Failed to get block-device info for %s: %v", dev, err)
		}
	}
	return resErr
}

func (reader *blockStatReader) readAllocation() (result Value) {
	for _, info := range reader.info {
		result += Value(info.Allocation())
	}
	return
}

func (reader *blockStatReader) readCapacity() (result Value) {
	for _, info := range reader.info {
		result += Value(info.Capacity())
	}
	return
}

func (reader *blockStatReader) readPhysical() (result Value) {
	for _, info := range reader.info {
		result += Value(info.Physical())
	}
	return
}

// ==================== Interface info ====================
var DomainInterfaceXPath = xmlpath.MustCompile("/domain/devices/interface/target/@dev")

type interfaceStatReader struct {
	parsedInterfaces bool
	interfaces       []string
	bytes            ValueRing
	packets          ValueRing
	errors           ValueRing
	dropped          ValueRing
}

func (reader *interfaceStatReader) description(xmlDesc *xmlpath.Node) {
	reader.interfaces = reader.interfaces[0:0]
	for iter := DomainInterfaceXPath.Iter(xmlDesc); iter.Next(); {
		reader.interfaces = append(reader.interfaces, iter.Node().String())
	}
	reader.parsedInterfaces = true
}

func (reader *interfaceStatReader) register(domainName string) map[string]MetricReader {
	reader.bytes = NewValueRing(NetIoLogback)
	reader.packets = NewValueRing(NetIoLogback)
	reader.errors = NewValueRing(NetIoLogback)
	reader.dropped = NewValueRing(NetIoLogback)
	return map[string]MetricReader{
		"libvirt/" + domainName + "/net-io/bytes":   reader.readBytes,
		"libvirt/" + domainName + "/net-io/packets": reader.readPackets,
		"libvirt/" + domainName + "/net-io/errors":  reader.readErrors,
		"libvirt/" + domainName + "/net-io/dropped": reader.readDropped,
	}
	return nil
}

func (reader *interfaceStatReader) update(domain libvirt.VirDomain) error {
	if !reader.parsedInterfaces {
		return UpdateXmlDescription
	}
	for _, interfaceName := range reader.interfaces {
		// More detailed alternative: domain.GetInterfaceParameters()
		stats, err := domain.InterfaceStats(interfaceName)
		if err == nil {
			reader.bytes.Add(Value(stats.RxBytes + stats.TxBytes))
			reader.packets.Add(Value(stats.RxPackets + stats.TxPackets))
			reader.errors.Add(Value(stats.RxErrs + stats.TxErrs))
			reader.dropped.Add(Value(stats.RxDrop + stats.TxDrop))
		} else {
			return fmt.Errorf("Failed to update vNIC stats for %s: %v", interfaceName, err)
		}
	}
	return nil
}

func (reader *interfaceStatReader) readBytes() Value {
	return reader.bytes.GetDiff(NetIoInterval)
}

func (reader *interfaceStatReader) readPackets() Value {
	return reader.packets.GetDiff(NetIoInterval)
}

func (reader *interfaceStatReader) readErrors() Value {
	return reader.errors.GetDiff(NetIoInterval)
}

func (reader *interfaceStatReader) readDropped() Value {
	return reader.dropped.GetDiff(NetIoInterval)
}
