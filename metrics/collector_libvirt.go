package metrics

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"gopkg.in/xmlpath.v1"

	"github.com/antongulenko/data2go/sample"
	"github.com/antongulenko/golib"
	"github.com/rgbkrk/libvirt-go"
)

const (
	NO_FLAGS              = 0
	DomainReparseInterval = 5 * time.Minute
)
const (
	LibvirtNetIoLogback   = 50
	LibvirtNetIoInterval  = 1 * time.Second
	LibvirtCpuTimeLogback = 10
	LibvirtCpuInterval    = 1 * time.Second
)

/*
	// TODO info about the node/hypervisor?

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
	conn       *libvirt.VirConnection
	domains    map[string]libvirt.VirDomain
	vmReaders  []*vmMetricsCollector
}

func (col *LibvirtCollector) Init() error {
	col.Close()
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
				&activatedMetricsReader{NewVmGeneralReader(), false},
				&activatedMetricsReader{new(memoryStatReader), false},
				&activatedMetricsReader{NewCpuStatReader(), false},
				&activatedMetricsReader{new(blockStatReader), false},
				&activatedMetricsReader{NewInterfaceStatReader(), false},
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
	conn, err := col.connection()
	if err != nil {
		return err
	}
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

func (col *LibvirtCollector) connection() (*libvirt.VirConnection, error) {
	conn := col.conn
	if conn != nil {
		if alive, err := conn.IsAlive(); err != nil || !alive {
			log.Println("Libvirt alive connection check failed:", err)
			col.Close()
			conn = nil
		}
	}
	if conn == nil {
		newConn, err := libvirt.NewVirConnection(col.ConnectUri)
		if err != nil {
			return nil, err
		}
		conn = &newConn
		col.conn = conn
	}
	return conn, nil
}

func (col *LibvirtCollector) Close() {
	if col.conn != nil {
		if err := col.conn.UnrefAndCloseConnection(); err != nil {
			log.Println("Error closing libvirt connection:", err)
		}
		col.conn = nil
	}
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
	var res golib.MultiError
	for _, reader := range collector.readers {
		if reader.active {
			if err := reader.update(domain); err == UpdateXmlDescription {
				collector.needXmlDesc = true
				updateDesc = true
			} else if err != nil {
				res.Add(fmt.Errorf("Failed to update domain %s: %v", collector.name, err))
				updateDesc = true
				break
			}
		}
	}
	if !updateDesc && time.Now().Sub(collector.xmlDescParsed) >= DomainReparseInterval {
		updateDesc = true
	}
	if collector.needXmlDesc && updateDesc {
		res.Add(collector.updateXmlDesc(domain))
	}
	return res.NilOrError()
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
	cpu  ValueRing
}

func NewVmGeneralReader() *vmGeneralReader {
	return &vmGeneralReader{
		cpu: NewValueRing(LibvirtCpuTimeLogback, LibvirtCpuInterval),
	}
}

type LogbackCpuVal uint64

func (val LogbackCpuVal) DiffValue(logback LogbackValue, interval time.Duration) sample.Value {
	switch previous := logback.(type) {
	case LogbackCpuVal:
		return sample.Value(val-previous) / sample.Value(interval.Nanoseconds())
	case *LogbackCpuVal:
		return sample.Value(val-*previous) / sample.Value(interval.Nanoseconds())
	default:
		log.Printf("Error: Cannot diff %v (%T) and %v (%T)\n", val, val, logback, logback)
		return sample.Value(0)
	}
}

func (val LogbackCpuVal) AddValue(logback LogbackValue) LogbackValue {
	switch previous := logback.(type) {
	case LogbackCpuVal:
		return StoredValue(val + previous)
	case *LogbackCpuVal:
		return StoredValue(val + *previous)
	default:
		log.Printf("Error: Cannot add %v (%T) and %v (%T)\n", val, val, logback, logback)
		return StoredValue(0)
	}
}

func (reader *vmGeneralReader) register(domainName string) map[string]MetricReader {
	return map[string]MetricReader{
		"libvirt/" + domainName + "/general/cpu":    reader.cpu.GetDiff,
		"libvirt/" + domainName + "/general/maxMem": reader.readMaxMem,
		"libvirt/" + domainName + "/general/mem":    reader.readMem,
	}
}

func (reader *vmGeneralReader) description(xmlDesc *xmlpath.Node) {
}

func (reader *vmGeneralReader) update(domain libvirt.VirDomain) (err error) {
	reader.info, err = domain.GetInfo()
	if err == nil {
		reader.cpu.Add(LogbackCpuVal(reader.info.GetCpuTime()))
	}
	return
}

func (reader *vmGeneralReader) readMaxMem() sample.Value {
	return sample.Value(reader.info.GetMaxMem())
}

func (reader *vmGeneralReader) readMem() sample.Value {
	return sample.Value(reader.info.GetMemory())
}

// ==================== Memory Stats ====================
const (
	VIR_DOMAIN_MEMORY_STAT_SWAP_OUT       = 1
	VIR_DOMAIN_MEMORY_STAT_AVAILABLE      = 5 // Max usable memory
	VIR_DOMAIN_MEMORY_STAT_ACTUAL_BALLOON = 6 // Used memory?
	VIR_DOMAIN_MEMORY_STAT_RSS            = 7 // Occuppied by VM process
	MAX_NUM_MEMORY_STATS                  = 8
)

type memoryStatReader struct {
	swap      uint64
	available uint64
	balloon   uint64
	rss       uint64
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

func (reader *memoryStatReader) update(domain libvirt.VirDomain) error {
	if memStats, err := domain.MemoryStats(MAX_NUM_MEMORY_STATS, NO_FLAGS); err != nil {
		return err
	} else {
		for _, stat := range memStats {
			switch stat.Tag {
			case VIR_DOMAIN_MEMORY_STAT_SWAP_OUT:
				reader.swap = stat.Val
			case VIR_DOMAIN_MEMORY_STAT_AVAILABLE:
				reader.available = stat.Val
			case VIR_DOMAIN_MEMORY_STAT_ACTUAL_BALLOON:
				reader.balloon = stat.Val
			case VIR_DOMAIN_MEMORY_STAT_RSS:
				reader.rss = stat.Val
			}
		}
		return nil
	}
}

func (reader *memoryStatReader) readSwap() sample.Value {
	return sample.Value(reader.swap)
}

func (reader *memoryStatReader) readAvailable() sample.Value {
	return sample.Value(reader.available)
}

func (reader *memoryStatReader) readBalloon() sample.Value {
	return sample.Value(reader.balloon)
}

func (reader *memoryStatReader) readRss() sample.Value {
	return sample.Value(reader.rss)
}

// ==================== CPU Stats ====================
const (
	MAX_NUM_CPU_STATS               = 4
	VIR_DOMAIN_CPU_STATS_CPUTIME    = "cpu_time" // Total CPU (VM + hypervisor)
	VIR_DOMAIN_CPU_STATS_SYSTEMTIME = "system_time"
	VIR_DOMAIN_CPU_STATS_USERTIME   = "user_time"
	VIR_DOMAIN_CPU_STATS_VCPUTIME   = "vcpu_time" // Excluding hypervisor usage
)

type cpuStatReader struct {
	cpu_total  ValueRing
	cpu_system ValueRing
	cpu_user   ValueRing
	cpu_virt   ValueRing
}

func NewCpuStatReader() *cpuStatReader {
	return &cpuStatReader{
		cpu_system: NewValueRing(LibvirtCpuTimeLogback, LibvirtCpuInterval),
		cpu_user:   NewValueRing(LibvirtCpuTimeLogback, LibvirtCpuInterval),
		cpu_total:  NewValueRing(LibvirtCpuTimeLogback, LibvirtCpuInterval),
		cpu_virt:   NewValueRing(LibvirtCpuTimeLogback, LibvirtCpuInterval),
	}
}

func (reader *cpuStatReader) register(domainName string) map[string]MetricReader {
	return map[string]MetricReader{
		"libvirt/" + domainName + "/cpu":        reader.cpu_total.GetDiff,
		"libvirt/" + domainName + "/cpu/user":   reader.cpu_user.GetDiff,
		"libvirt/" + domainName + "/cpu/system": reader.cpu_system.GetDiff,
		"libvirt/" + domainName + "/cpu/virt":   reader.cpu_virt.GetDiff,
	}
}

func (reader *cpuStatReader) description(xmlDesc *xmlpath.Node) {
}

func (reader *cpuStatReader) update(domain libvirt.VirDomain) error {
	stats := make(libvirt.VirTypedParameters, MAX_NUM_CPU_STATS)
	// Less detailed alternative: domain.GetVcpus()
	if _, err := domain.GetCPUStats(&stats, len(stats), -1, 1, NO_FLAGS); err != nil {
		return err
	} else {
		for _, param := range stats {
			val, ok := param.Value.(uint64)
			if !ok {
				continue
			}
			switch param.Name {
			case VIR_DOMAIN_CPU_STATS_CPUTIME:
				reader.cpu_total.Add(LogbackCpuVal(val))
			case VIR_DOMAIN_CPU_STATS_USERTIME:
				reader.cpu_user.Add(LogbackCpuVal(val))
			case VIR_DOMAIN_CPU_STATS_SYSTEMTIME:
				reader.cpu_system.Add(LogbackCpuVal(val))
			case VIR_DOMAIN_CPU_STATS_VCPUTIME:
				reader.cpu_virt.Add(LogbackCpuVal(val))
			}
		}
		return nil
	}
}

// ==================== Block info ====================
var DomainBlockXPath = xmlpath.MustCompile("/domain/devices/disk[@type=\"file\"]/target/@dev")

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
	reader.devices = reader.devices[0:0]
	for iter := DomainBlockXPath.Iter(xmlDesc); iter.Next(); {
		reader.devices = append(reader.devices, iter.Node().String())
	}
	reader.parsedDevices = true
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

func (reader *blockStatReader) readAllocation() (result sample.Value) {
	for _, info := range reader.info {
		result += sample.Value(info.Allocation())
	}
	return
}

func (reader *blockStatReader) readCapacity() (result sample.Value) {
	for _, info := range reader.info {
		result += sample.Value(info.Capacity())
	}
	return
}

func (reader *blockStatReader) readPhysical() (result sample.Value) {
	for _, info := range reader.info {
		result += sample.Value(info.Physical())
	}
	return
}

// ==================== Interface info ====================
var DomainInterfaceXPath = xmlpath.MustCompile("/domain/devices/interface/target/@dev")

type interfaceStatReader struct {
	parsedInterfaces bool
	interfaces       []string
	net              netIoCounters
}

func NewInterfaceStatReader() *interfaceStatReader {
	return &interfaceStatReader{
		net: NewNetIoCounters(NetIoLogback, NetIoInterval),
	}
}

func (reader *interfaceStatReader) description(xmlDesc *xmlpath.Node) {
	reader.interfaces = reader.interfaces[0:0]
	for iter := DomainInterfaceXPath.Iter(xmlDesc); iter.Next(); {
		reader.interfaces = append(reader.interfaces, iter.Node().String())
	}
	reader.parsedInterfaces = true
}

func (reader *interfaceStatReader) register(domainName string) map[string]MetricReader {
	result := make(map[string]MetricReader)
	reader.net.Register(result, "libvirt/"+domainName+"/net-io")
	return result
}

func (reader *interfaceStatReader) update(domain libvirt.VirDomain) error {
	if !reader.parsedInterfaces {
		return UpdateXmlDescription
	}
	for _, interfaceName := range reader.interfaces {
		// More detailed alternative: domain.GetInterfaceParameters()
		stats, err := domain.InterfaceStats(interfaceName)
		if err == nil {
			reader.net.bytes.Add(StoredValue(stats.RxBytes + stats.TxBytes))
			reader.net.packets.Add(StoredValue(stats.RxPackets + stats.TxPackets))
			reader.net.rx_bytes.Add(StoredValue(stats.RxBytes))
			reader.net.rx_packets.Add(StoredValue(stats.RxPackets))
			reader.net.tx_bytes.Add(StoredValue(stats.TxBytes))
			reader.net.tx_packets.Add(StoredValue(stats.TxPackets))
			reader.net.errors.Add(StoredValue(stats.RxErrs + stats.TxErrs))
			reader.net.dropped.Add(StoredValue(stats.RxDrop + stats.TxDrop))
		} else {
			return fmt.Errorf("Failed to update vNIC stats for %s: %v", interfaceName, err)
		}
	}
	return nil
}
