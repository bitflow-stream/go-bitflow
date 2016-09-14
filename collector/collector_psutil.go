package collector

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/antongulenko/data2go/sample"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
	psnet "github.com/shirou/gopsutil/net"
	"github.com/shirou/gopsutil/process"
)

func RegisterPsutilCollectors(osInfoUpdate time.Duration, factory *ValueRingFactory) {
	RegisterCollector(new(PsutilMemCollector))
	RegisterCollector(&PsutilCpuCollector{Factory: factory})
	RegisterCollector(new(PsutilLoadCollector))
	RegisterCollector(&PsutilNetCollector{Factory: factory})
	RegisterCollector(&PsutilNetProtoCollector{Factory: factory})
	RegisterCollector(&PsutilDiskIOCollector{Factory: factory})
	RegisterCollector(new(PsutilDiskUsageCollector))
	RegisterCollector(new(PsutilMiscCollector))
	go UpdateRunningPids(osInfoUpdate)
}

// ==================== Memory ====================
type PsutilMemCollector struct {
	AbstractCollector
	memory *mem.VirtualMemoryStat
}

func (col *PsutilMemCollector) Init() error {
	col.Reset(col)
	col.readers = map[string]MetricReader{
		"mem/free":    col.readFreeMem,
		"mem/used":    col.readUsedMem,
		"mem/percent": col.readUsedPercentMem,
	}
	return nil
}

func (col *PsutilMemCollector) Update() (err error) {
	col.memory, err = mem.VirtualMemory()
	if err == nil {
		col.UpdateMetrics()
	}
	return
}

func (col *PsutilMemCollector) readFreeMem() sample.Value {
	return sample.Value(col.memory.Available)
}

func (col *PsutilMemCollector) readUsedMem() sample.Value {
	return sample.Value(col.memory.Used)
}

func (col *PsutilMemCollector) readUsedPercentMem() sample.Value {
	return sample.Value(col.memory.UsedPercent)
}

func hostProcFile(parts ...string) string {
	// Forbidden import: "github.com/shirou/gopsutil/internal/common"
	// return common.HostProc(parts...)
	all := make([]string, len(parts)+1)
	all[0] = "/proc"
	copy(all[1:], parts)
	return filepath.Join(all...)
}

// ==================== CPU ====================
type PsutilCpuCollector struct {
	AbstractCollector
	Factory *ValueRingFactory
	ring    *ValueRing
}

func (col *PsutilCpuCollector) Init() error {
	col.Reset(col)
	col.ring = col.Factory.NewValueRing()
	col.readers = map[string]MetricReader{
		"cpu": col.ring.GetDiff,
	}
	return nil
}

func (col *PsutilCpuCollector) Update() (err error) {
	times, err := cpu.Times(false)
	if err == nil {
		if len(times) != 1 {
			err = fmt.Errorf("warning: gopsutil/cpu.Times() returned %v CPUTimes instead of %v", len(times), 1)
		} else {
			col.ring.Add(&cpuTime{times[0]})
			col.UpdateMetrics()
		}
	}
	return
}

type cpuTime struct {
	cpu.TimesStat
}

func (t *cpuTime) getAllBusy() (float64, float64) {
	busy := t.User + t.System + t.Nice + t.Irq +
		t.Softirq + t.Steal + t.Guest + t.GuestNice + t.Stolen
	return busy + t.Idle + t.Iowait, busy
}

func (t *cpuTime) DiffValue(logback LogbackValue, _ time.Duration) sample.Value {
	if previous, ok := logback.(*cpuTime); ok {
		// Calculation based on https://github.com/shirou/gopsutil/blob/master/cpu/cpu_unix.go
		t1All, t1Busy := previous.getAllBusy()
		t2All, t2Busy := t.getAllBusy()

		if t2Busy <= t1Busy {
			return 0
		}
		if t2All <= t1All {
			return 1
		}
		return sample.Value((t2Busy - t1Busy) / (t2All - t1All) * 100)
	} else {
		log.Errorf("Cannot diff %v (%T) and %v (%T)", t, t, logback, logback)
		return sample.Value(0)
	}
}

func (t *cpuTime) AddValue(incoming LogbackValue) LogbackValue {
	if other, ok := incoming.(*cpuTime); ok {
		return &cpuTime{
			cpu.TimesStat{
				User:      t.User + other.User,
				System:    t.System + other.System,
				Idle:      t.Idle + other.Idle,
				Nice:      t.Nice + other.Nice,
				Iowait:    t.Iowait + other.Iowait,
				Irq:       t.Irq + other.Irq,
				Softirq:   t.Softirq + other.Softirq,
				Steal:     t.Steal + other.Steal,
				Guest:     t.Guest + other.Guest,
				GuestNice: t.GuestNice + other.GuestNice,
				Stolen:    t.Stolen + other.Stolen,
			},
		}
	} else {
		log.Errorf("Cannot add %v (%T) and %v (%T)", t, t, incoming, incoming)
		return StoredValue(0)
	}
}

// ==================== Load ====================
type PsutilLoadCollector struct {
	AbstractCollector
	load *load.AvgStat
}

func (col *PsutilLoadCollector) Init() error {
	col.Reset(col)
	col.readers = map[string]MetricReader{
		"load/1":  col.readLoad1,
		"load/5":  col.readLoad5,
		"load/15": col.readLoad15,
	}
	return nil
}

func (col *PsutilLoadCollector) Update() (err error) {
	col.load, err = load.Avg()
	if err == nil {
		col.UpdateMetrics()
	}
	return
}

func (col *PsutilLoadCollector) readLoad1() sample.Value {
	return sample.Value(col.load.Load1)
}

func (col *PsutilLoadCollector) readLoad5() sample.Value {
	return sample.Value(col.load.Load5)
}

func (col *PsutilLoadCollector) readLoad15() sample.Value {
	return sample.Value(col.load.Load15)
}

// ==================== Net IO Counters ====================
type PsutilNetCollector struct {
	AbstractCollector
	Factory *ValueRingFactory

	counters netIoCounters
}

type netIoCounters struct {
	bytes      *ValueRing
	packets    *ValueRing
	rx_bytes   *ValueRing
	rx_packets *ValueRing
	tx_bytes   *ValueRing
	tx_packets *ValueRing
	errors     *ValueRing
	dropped    *ValueRing
}

func (col *PsutilNetCollector) Init() error {
	col.Reset(col)
	col.counters = NewNetIoCounters(col.Factory)
	col.readers = make(map[string]MetricReader)
	col.counters.Register(col.readers, "net-io")
	return nil
}

func (col *PsutilNetCollector) Update() (err error) {
	counters, err := psnet.IOCounters(false)
	if err == nil && len(counters) != 1 {
		err = fmt.Errorf("gopsutil/net.NetIOCounters() returned %v NetIOCountersStat instead of %v", len(counters), 1)
	}
	if err == nil {
		col.counters.Add(&counters[0])
		col.UpdateMetrics()
	}
	return
}

func NewNetIoCounters(factory *ValueRingFactory) netIoCounters {
	return netIoCounters{
		bytes:      factory.NewValueRing(),
		packets:    factory.NewValueRing(),
		rx_bytes:   factory.NewValueRing(),
		rx_packets: factory.NewValueRing(),
		tx_bytes:   factory.NewValueRing(),
		tx_packets: factory.NewValueRing(),
		errors:     factory.NewValueRing(),
		dropped:    factory.NewValueRing(),
	}
}

func (counters *netIoCounters) Add(stat *psnet.IOCountersStat) {
	counters.AddToHead(stat)
	counters.FlushHead()
}

func (counters *netIoCounters) AddToHead(stat *psnet.IOCountersStat) {
	counters.bytes.AddToHead(StoredValue(stat.BytesSent + stat.BytesRecv))
	counters.packets.AddToHead(StoredValue(stat.PacketsSent + stat.PacketsRecv))
	counters.rx_bytes.AddToHead(StoredValue(stat.BytesRecv))
	counters.rx_packets.AddToHead(StoredValue(stat.PacketsRecv))
	counters.tx_bytes.AddToHead(StoredValue(stat.BytesSent))
	counters.tx_packets.AddToHead(StoredValue(stat.PacketsSent))
	counters.errors.AddToHead(StoredValue(stat.Errin + stat.Errout))
	counters.dropped.AddToHead(StoredValue(stat.Dropin + stat.Dropout))
}

func (counters *netIoCounters) FlushHead() {
	counters.bytes.FlushHead()
	counters.packets.FlushHead()
	counters.rx_bytes.FlushHead()
	counters.rx_packets.FlushHead()
	counters.tx_bytes.FlushHead()
	counters.tx_packets.FlushHead()
	counters.errors.FlushHead()
	counters.dropped.FlushHead()
}

func (counters *netIoCounters) Register(target map[string]MetricReader, prefix string) {
	target[prefix+"/bytes"] = counters.bytes.GetDiff
	target[prefix+"/packets"] = counters.packets.GetDiff
	target[prefix+"/rx_bytes"] = counters.rx_bytes.GetDiff
	target[prefix+"/rx_packets"] = counters.rx_packets.GetDiff
	target[prefix+"/tx_bytes"] = counters.tx_bytes.GetDiff
	target[prefix+"/tx_packets"] = counters.tx_packets.GetDiff
	target[prefix+"/errors"] = counters.errors.GetDiff
	target[prefix+"/dropped"] = counters.dropped.GetDiff
}

// ==================== Net Protocol Counters ====================
var absoluteNetProtoValues = map[string]bool{
	// These values will not be aggregated through ValueRing
	"NoPorts":      true, // udp, udplite
	"CurrEstab":    true, // tcp
	"MaxConn":      true,
	"RtpAlgorithm": true,
	"RtoMax":       true,
	"RtpMin":       true,
	"DefaultTTL":   true, // ip
	"Forwarding":   true,
}

type PsutilNetProtoCollector struct {
	AbstractCollector
	Factory *ValueRingFactory

	protocols    map[string]psnet.ProtoCountersStat
	protoReaders []*protoStatReader
}

func (col *PsutilNetProtoCollector) Init() error {
	col.Reset(col)
	col.protocols = make(map[string]psnet.ProtoCountersStat)
	col.protoReaders = nil

	// TODO missing: metrics about individual connections and NICs
	if err := col.update(false); err != nil {
		return err
	}
	col.readers = make(map[string]MetricReader)
	for proto, counters := range col.protocols {
		for statName, _ := range counters.Stats {
			var ring *ValueRing
			if !absoluteNetProtoValues[statName] {
				ring = col.Factory.NewValueRing()
			}
			name := "net-proto/" + proto + "/" + statName
			protoReader := &protoStatReader{
				col:      col,
				protocol: proto,
				field:    statName,
				ring:     ring,
			}
			col.readers[name] = protoReader.read
			col.protoReaders = append(col.protoReaders, protoReader)
		}
	}
	return nil
}

func (col *PsutilNetProtoCollector) update(checkChange bool) error {
	counters, err := psnet.ProtoCounters(nil)
	if err != nil {
		return err
	}
	for _, counters := range counters {
		if checkChange {
			if _, ok := col.protocols[counters.Protocol]; !ok {
				return MetricsChanged
			}
		}
		col.protocols[counters.Protocol] = counters
	}
	if checkChange && len(counters) != len(col.protocols) {
		// This means some previous metric is not available anymore
		return MetricsChanged
	}
	return nil
}

func (col *PsutilNetProtoCollector) Update() (err error) {
	if err = col.update(true); err == nil {
		for _, protoReader := range col.protoReaders {
			if err := protoReader.update(); err != nil {
				return err
			}
		}
		col.UpdateMetrics()
	}
	return
}

type protoStatReader struct {
	col      *PsutilNetProtoCollector
	protocol string
	field    string

	// Only one of the following 2 fields is used
	ring  *ValueRing
	value sample.Value
}

func (reader *protoStatReader) update() error {
	if counters, ok := reader.col.protocols[reader.protocol]; ok {
		if val, ok := counters.Stats[reader.field]; ok {
			if reader.ring != nil {
				reader.ring.Add(StoredValue(val))
			} else {
				reader.value = sample.Value(val)
			}
			return nil
		} else {
			return fmt.Errorf("Counter %v not found in protocol %v in PsutilNetProtoCollector", reader.field, reader.protocol)
		}
	} else {
		return fmt.Errorf("Protocol %v not found in PsutilNetProtoCollector", reader.protocol)
	}
}

func (reader *protoStatReader) read() sample.Value {
	if ring := reader.ring; ring != nil {
		return ring.GetDiff()
	} else {
		return reader.value
	}
}

// ==================== Disk IO ====================
type PsutilDiskIOCollector struct {
	AbstractCollector
	Factory *ValueRingFactory

	disks map[string]disk.IOCountersStat
}

func (col *PsutilDiskIOCollector) Init() error {
	col.Reset(col)
	col.disks = make(map[string]disk.IOCountersStat)

	if err := col.update(false); err != nil {
		return err
	}
	col.readers = make(map[string]MetricReader)
	for disk, _ := range col.disks {
		name := "disk-io/" + disk + "/"
		reader := &diskIOReader{
			col:            col,
			disk:           disk,
			readRing:       col.Factory.NewValueRing(),
			writeRing:      col.Factory.NewValueRing(),
			ioRing:         col.Factory.NewValueRing(),
			readBytesRing:  col.Factory.NewValueRing(),
			writeBytesRing: col.Factory.NewValueRing(),
			ioBytesRing:    col.Factory.NewValueRing(),
			readTimeRing:   col.Factory.NewValueRing(),
			writeTimeRing:  col.Factory.NewValueRing(),
			ioTimeRing:     col.Factory.NewValueRing(),
		}
		col.readers[name+"read"] = reader.readRead
		col.readers[name+"write"] = reader.readWrite
		col.readers[name+"io"] = reader.readIo
		col.readers[name+"readBytes"] = reader.readReadBytes
		col.readers[name+"writeBytes"] = reader.readWriteBytes
		col.readers[name+"ioBytes"] = reader.readIoBytes
		col.readers[name+"readTime"] = reader.readReadTime
		col.readers[name+"writeTime"] = reader.readWriteTime
		col.readers[name+"ioTime"] = reader.readIoTime
	}
	return nil
}

func (col *PsutilDiskIOCollector) update(checkChange bool) error {
	disks, err := disk.IOCounters()
	if err != nil {
		return err
	}
	if checkChange {
		for k, _ := range col.disks {
			if _, ok := disks[k]; !ok {
				return MetricsChanged
			}
		}
		if len(col.disks) != len(disks) {
			return MetricsChanged
		}
	}
	col.disks = disks
	return nil
}

func (col *PsutilDiskIOCollector) Update() (err error) {
	if err = col.update(true); err == nil {
		col.UpdateMetrics()
	}
	return
}

type diskIOReader struct {
	col  *PsutilDiskIOCollector
	disk string

	readRing       *ValueRing
	writeRing      *ValueRing
	ioRing         *ValueRing
	readBytesRing  *ValueRing
	writeBytesRing *ValueRing
	ioBytesRing    *ValueRing
	readTimeRing   *ValueRing
	writeTimeRing  *ValueRing
	ioTimeRing     *ValueRing
}

func (reader *diskIOReader) checkDisk() *disk.IOCountersStat {
	if disk, ok := reader.col.disks[reader.disk]; ok {
		return &disk
	} else {
		log.Warnf("disk-io counters for disk %v not found", reader.disk)
		return nil
	}
}

func (reader *diskIOReader) value(val uint64, ring *ValueRing) sample.Value {
	ring.Add(StoredValue(val))
	return ring.GetDiff()
}

func (reader *diskIOReader) readRead() sample.Value {
	if disk := reader.checkDisk(); disk != nil {
		return reader.value(disk.ReadCount, reader.readRing)
	}
	return sample.Value(0)
}

func (reader *diskIOReader) readWrite() sample.Value {
	if disk := reader.checkDisk(); disk != nil {
		return reader.value(disk.WriteCount, reader.writeRing)
	}
	return sample.Value(0)
}

func (reader *diskIOReader) readIo() sample.Value {
	if disk := reader.checkDisk(); disk != nil {
		return reader.value(disk.ReadCount+disk.WriteCount, reader.ioRing)
	}
	return sample.Value(0)
}

func (reader *diskIOReader) readReadBytes() sample.Value {
	if disk := reader.checkDisk(); disk != nil {
		return reader.value(disk.ReadBytes, reader.readBytesRing)
	}
	return sample.Value(0)
}

func (reader *diskIOReader) readWriteBytes() sample.Value {
	if disk := reader.checkDisk(); disk != nil {
		return reader.value(disk.WriteBytes, reader.writeBytesRing)
	}
	return sample.Value(0)
}

func (reader *diskIOReader) readIoBytes() sample.Value {
	if disk := reader.checkDisk(); disk != nil {
		return reader.value(disk.ReadBytes+disk.WriteBytes, reader.ioBytesRing)
	}
	return sample.Value(0)
}

func (reader *diskIOReader) readReadTime() sample.Value {
	if disk := reader.checkDisk(); disk != nil {
		return reader.value(disk.ReadTime, reader.readTimeRing)
	}
	return sample.Value(0)
}

func (reader *diskIOReader) readWriteTime() sample.Value {
	if disk := reader.checkDisk(); disk != nil {
		return reader.value(disk.WriteTime, reader.writeTimeRing)
	}
	return sample.Value(0)
}

func (reader *diskIOReader) readIoTime() sample.Value {
	if disk := reader.checkDisk(); disk != nil {
		return reader.value(disk.IoTime, reader.ioTimeRing)
	}
	return sample.Value(0)
}

// ==================== Disk Usage ====================
type PsutilDiskUsageCollector struct {
	AbstractCollector
	allPartitions      map[string]bool
	observedPartitions map[string]bool
	usage              map[string]*disk.UsageStat
}

func (col *PsutilDiskUsageCollector) Init() error {
	col.Reset(col)
	col.usage = make(map[string]*disk.UsageStat)
	col.observedPartitions = make(map[string]bool)

	var err error
	col.allPartitions, err = col.getAllPartitions()
	if err != nil {
		return err
	}

	col.readers = make(map[string]MetricReader)
	for partition, _ := range col.allPartitions {
		name := "disk-usage/" + partition + "/"
		reader := &diskUsageReader{
			col:       col,
			partition: partition,
		}
		col.readers[name+"free"] = reader.readFree
		col.readers[name+"used"] = reader.readPercent
	}
	return nil
}

func (col *PsutilDiskUsageCollector) Collect(metric *Metric) error {
	lastSlash := strings.LastIndex(metric.Name, "/")
	partition := metric.Name[len("disk-usage/"):lastSlash]
	col.observedPartitions[partition] = true
	return col.AbstractCollector.Collect(metric)
}

func (col *PsutilDiskUsageCollector) getAllPartitions() (map[string]bool, error) {
	partitions, err := disk.Partitions(true)
	if err != nil {
		return nil, err
	}
	result := make(map[string]bool, len(partitions))
	for _, partition := range partitions {
		result[partition.Mountpoint] = true
	}
	return result, nil
}

func (col *PsutilDiskUsageCollector) update() error {
	for partition, _ := range col.observedPartitions {
		usage, err := disk.Usage(partition)
		if err != nil {
			return err
		}
		col.usage[partition] = usage
	}
	return nil
}

func (col *PsutilDiskUsageCollector) checkChangedPartitions() error {
	partitions, err := disk.Partitions(true)
	if err != nil {
		return err
	}
	checked := make(map[string]bool, len(partitions))
	for _, partition := range partitions {
		if _, ok := col.allPartitions[partition.Mountpoint]; !ok {
			return MetricsChanged
		}
		checked[partition.Mountpoint] = true
	}
	if len(checked) != len(col.allPartitions) {
		return MetricsChanged
	}
	return nil
}

func (col *PsutilDiskUsageCollector) Update() error {
	if err := col.checkChangedPartitions(); err != nil {
		return err
	}
	if err := col.update(); err == nil {
		col.UpdateMetrics()
		return nil
	} else {
		return err
	}
}

type diskUsageReader struct {
	col       *PsutilDiskUsageCollector
	partition string
}

func (reader *diskUsageReader) checkDisk() *disk.UsageStat {
	if disk, ok := reader.col.usage[reader.partition]; ok {
		return disk
	} else {
		log.Warnf("disk-usage counters for partition %v not found", reader.partition)
		return nil
	}
}

func (reader *diskUsageReader) readFree() sample.Value {
	if disk := reader.checkDisk(); disk != nil {
		return sample.Value(disk.Free)
	}
	return sample.Value(0)
}

func (reader *diskUsageReader) readPercent() sample.Value {
	if disk := reader.checkDisk(); disk != nil {
		return sample.Value(disk.UsedPercent)
	}
	return sample.Value(0)
}

// ==================== Misc OS Metrics ====================
// Global information updated regularly
var osInformation struct {
	pids []int32
}

// This is required for PsutilMiscCollector and PsutilProcessCollector
func UpdateRunningPids(interval time.Duration) {
	for {
		if pids, err := process.Pids(); err != nil {
			log.Errorln("Failed to update PIDs:", err)
		} else {
			osInformation.pids = pids
		}
		time.Sleep(interval)
	}
}

type PsutilMiscCollector struct {
	AbstractCollector
}

func (col *PsutilMiscCollector) Init() error {
	// TODO missing: number of open files, threads, etc in entire OS
	col.Reset(col)
	col.readers = map[string]MetricReader{
		"num_procs": col.readNumProcs,
	}
	return nil
}

func (col *PsutilMiscCollector) Update() error {
	col.UpdateMetrics()
	return nil
}

func (col *PsutilMiscCollector) readNumProcs() sample.Value {
	return sample.Value(len(osInformation.pids))
}

// ==================== Process Metrics ====================
type PsutilProcessCollector struct {
	AbstractCollector
	Factory *ValueRingFactory

	// Settings
	CmdlineFilter     []*regexp.Regexp
	GroupName         string
	PrintErrors       bool
	PidUpdateInterval time.Duration

	pidsUpdated bool
	own_pid     int32
	cpu_factor  float64
	pids        map[int32]*SingleProcessCollector
}

func (col *PsutilProcessCollector) Init() error {
	col.own_pid = int32(os.Getpid())
	col.cpu_factor = 100 / float64(runtime.NumCPU())
	col.Reset(col)

	prefix := "proc/" + col.GroupName
	col.readers = map[string]MetricReader{
		prefix + "/num": col.readNumProc,
		prefix + "/cpu": col.readCpu,

		prefix + "/disk/read":       col.readIoRead,
		prefix + "/disk/write":      col.readIoWrite,
		prefix + "/disk/io":         col.readIo,
		prefix + "/disk/readBytes":  col.readBytesRead,
		prefix + "/disk/writeBytes": col.readBytesWrite,
		prefix + "/disk/ioBytes":    col.readBytes,

		prefix + "/ctxSwitch":             col.readCtxSwitch,
		prefix + "/ctxSwitch/voluntary":   col.readCtxSwitchVoluntary,
		prefix + "/ctxSwitch/involuntary": col.readCtxSwitchInvoluntary,

		prefix + "/mem/rss":  col.readMemRss,
		prefix + "/mem/vms":  col.readMemVms,
		prefix + "/mem/swap": col.readMemSwap,
		prefix + "/fds":      col.readFds,
		prefix + "/threads":  col.readThreads,

		prefix + "/net-io/bytes":      col.readNetBytes,
		prefix + "/net-io/packets":    col.readNetPackets,
		prefix + "/net-io/rx_bytes":   col.readNetRxBytes,
		prefix + "/net-io/rx_packets": col.readNetRxPackets,
		prefix + "/net-io/tx_bytes":   col.readNetTxBytes,
		prefix + "/net-io/tx_packets": col.readNetTxPackets,
		prefix + "/net-io/errors":     col.readNetErrors,
		prefix + "/net-io/dropped":    col.readNetDropped,
	}
	return nil
}

func (col *PsutilProcessCollector) Update() (err error) {
	if err := col.updatePids(); err != nil {
		return err
	}
	col.updateProcesses()
	if err == nil {
		col.UpdateMetrics()
	}
	return
}

func (col *PsutilProcessCollector) updatePids() error {
	if col.pidsUpdated {
		return nil
	}

	newPids := make(map[int32]*SingleProcessCollector)
	errors := 0
	pids := osInformation.pids
	for _, pid := range pids {
		if pid == col.own_pid {
			continue
		}
		proc, err := process.NewProcess(pid)
		if err != nil {
			// Process does not exist anymore
			errors++
			if col.PrintErrors {
				log.WithField("pid", pid).Warnln("Checking process failed:", err)
			}
			continue
		}
		cmdline, err := proc.Cmdline()
		if err != nil {
			// Probably a permission error
			errors++
			if col.PrintErrors {
				log.WithField("pid", pid).Warnln("Obtaining process cmdline failed:", err)
			}
			continue
		}
		for _, regex := range col.CmdlineFilter {
			if regex.MatchString(cmdline) {
				procCollector, ok := col.pids[pid]
				if !ok {
					procCollector = col.MakeProcessCollector(proc, col.Factory)
				}
				newPids[pid] = procCollector
				break
			}
		}
	}
	if len(newPids) == 0 && errors > 0 && col.PrintErrors {
		log.Errorln("Warning: Observing no processes, failed to check", errors, "out of", len(pids), "PIDs")
	}
	col.pids = newPids

	if col.PidUpdateInterval > 0 {
		col.pidsUpdated = true
		time.AfterFunc(col.PidUpdateInterval, func() {
			col.pidsUpdated = false
		})
	} else {
		col.pidsUpdated = false
	}
	return nil
}

func (col *PsutilProcessCollector) updateProcesses() {
	for pid, proc := range col.pids {
		if err := proc.update(); err != nil {
			// Process probably does not exist anymore
			delete(col.pids, pid)
			if col.PrintErrors {
				log.WithField("pid", pid).Warnln("Process info update failed:", err)
			}
		}
	}
}

func (col *PsutilProcessCollector) readNumProc() sample.Value {
	return sample.Value(len(col.pids))
}

func (col *PsutilProcessCollector) readCpu() (res sample.Value) {
	for _, proc := range col.pids {
		res += sample.Value(proc.cpu.GetDiff())
	}
	return
}

func (col *PsutilProcessCollector) readIoRead() (res sample.Value) {
	for _, proc := range col.pids {
		res += sample.Value(proc.ioRead.GetDiff())
	}
	return
}

func (col *PsutilProcessCollector) readIoWrite() (res sample.Value) {
	for _, proc := range col.pids {
		res += sample.Value(proc.ioWrite.GetDiff())
	}
	return
}

func (col *PsutilProcessCollector) readIo() (res sample.Value) {
	for _, proc := range col.pids {
		res += sample.Value(proc.ioTotal.GetDiff())
	}
	return
}

func (col *PsutilProcessCollector) readBytesRead() (res sample.Value) {
	for _, proc := range col.pids {
		res += sample.Value(proc.ioReadBytes.GetDiff())
	}
	return
}

func (col *PsutilProcessCollector) readBytesWrite() (res sample.Value) {
	for _, proc := range col.pids {
		res += sample.Value(proc.ioWriteBytes.GetDiff())
	}
	return
}

func (col *PsutilProcessCollector) readBytes() (res sample.Value) {
	for _, proc := range col.pids {
		res += sample.Value(proc.ioBytesTotal.GetDiff())
	}
	return
}

func (col *PsutilProcessCollector) readCtxSwitchVoluntary() (res sample.Value) {
	for _, proc := range col.pids {
		res += sample.Value(proc.ctxSwitchVoluntary.GetDiff())
	}
	return
}

func (col *PsutilProcessCollector) readCtxSwitchInvoluntary() (res sample.Value) {
	for _, proc := range col.pids {
		res += sample.Value(proc.ctxSwitchInvoluntary.GetDiff())
	}
	return
}

func (col *PsutilProcessCollector) readCtxSwitch() (res sample.Value) {
	for _, proc := range col.pids {
		res += sample.Value(proc.ctxSwitchInvoluntary.GetDiff())
		res += sample.Value(proc.ctxSwitchVoluntary.GetDiff())
	}
	return
}

func (col *PsutilProcessCollector) readMemRss() (res sample.Value) {
	for _, proc := range col.pids {
		res += sample.Value(proc.mem_rss)
	}
	return
}

func (col *PsutilProcessCollector) readMemVms() (res sample.Value) {
	for _, proc := range col.pids {
		res += sample.Value(proc.mem_vms)
	}
	return
}

func (col *PsutilProcessCollector) readMemSwap() (res sample.Value) {
	for _, proc := range col.pids {
		res += sample.Value(proc.mem_swap)
	}
	return
}

func (col *PsutilProcessCollector) readFds() (res sample.Value) {
	for _, proc := range col.pids {
		res += sample.Value(proc.numFds)
	}
	return
}

func (col *PsutilProcessCollector) readThreads() (res sample.Value) {
	for _, proc := range col.pids {
		res += sample.Value(proc.numThreads)
	}
	return
}

func (col *PsutilProcessCollector) readNetBytes() (res sample.Value) {
	for _, proc := range col.pids {
		res += sample.Value(proc.net.bytes.GetDiff())
	}
	return
}

func (col *PsutilProcessCollector) readNetPackets() (res sample.Value) {
	for _, proc := range col.pids {
		res += sample.Value(proc.net.packets.GetDiff())
	}
	return
}

func (col *PsutilProcessCollector) readNetRxBytes() (res sample.Value) {
	for _, proc := range col.pids {
		res += sample.Value(proc.net.rx_bytes.GetDiff())
	}
	return
}

func (col *PsutilProcessCollector) readNetRxPackets() (res sample.Value) {
	for _, proc := range col.pids {
		res += sample.Value(proc.net.rx_packets.GetDiff())
	}
	return
}

func (col *PsutilProcessCollector) readNetTxBytes() (res sample.Value) {
	for _, proc := range col.pids {
		res += sample.Value(proc.net.tx_bytes.GetDiff())
	}
	return
}

func (col *PsutilProcessCollector) readNetTxPackets() (res sample.Value) {
	for _, proc := range col.pids {
		res += sample.Value(proc.net.tx_packets.GetDiff())
	}
	return
}

func (col *PsutilProcessCollector) readNetErrors() (res sample.Value) {
	for _, proc := range col.pids {
		res += sample.Value(proc.net.errors.GetDiff())
	}
	return
}

func (col *PsutilProcessCollector) readNetDropped() (res sample.Value) {
	for _, proc := range col.pids {
		res += sample.Value(proc.net.dropped.GetDiff())
	}
	return
}

type SingleProcessCollector struct {
	parent *PsutilProcessCollector
	*process.Process
	tasks CollectorTasks

	updateLock           sync.Mutex
	cpu                  *ValueRing
	ioRead               *ValueRing
	ioWrite              *ValueRing
	ioTotal              *ValueRing
	ioReadBytes          *ValueRing
	ioWriteBytes         *ValueRing
	ioBytesTotal         *ValueRing
	ctxSwitchVoluntary   *ValueRing
	ctxSwitchInvoluntary *ValueRing
	net                  netIoCounters
	mem_rss              uint64
	mem_vms              uint64
	mem_swap             uint64
	numFds               int32
	numThreads           int32
}

func (collector *PsutilProcessCollector) MakeProcessCollector(proc *process.Process, factory *ValueRingFactory) *SingleProcessCollector {
	col := &SingleProcessCollector{
		parent:  collector,
		Process: proc,

		cpu:                  factory.NewValueRing(),
		ioRead:               factory.NewValueRing(),
		ioWrite:              factory.NewValueRing(),
		ioTotal:              factory.NewValueRing(),
		ioReadBytes:          factory.NewValueRing(),
		ioWriteBytes:         factory.NewValueRing(),
		ioBytesTotal:         factory.NewValueRing(),
		ctxSwitchVoluntary:   factory.NewValueRing(),
		ctxSwitchInvoluntary: factory.NewValueRing(),
		net:                  NewNetIoCounters(factory),
	}
	col.tasks = CollectorTasks{
		col.updateCpu,
		col.updateDisk,
		col.updateMemory,
		col.updateNet,
		col.updateOpenFiles,
		col.updateMisc,
	}
	return col
}

func (col *SingleProcessCollector) update() error {
	return col.tasks.Run()
}

func (col *SingleProcessCollector) updateCpu() error {
	if cpu, err := col.Times(); err != nil {
		return fmt.Errorf("Failed to get CPU info: %v", err)
	} else {
		busy := (cpu.Total() - cpu.Idle) * col.parent.cpu_factor
		col.cpu.Add(StoredValue(busy))
	}
	return nil
}

func (col *SingleProcessCollector) updateDisk() error {
	if io, err := col.IOCounters(); err != nil {
		return fmt.Errorf("Failed to get disk-IO info: %v", err)
	} else {
		col.ioRead.Add(StoredValue(io.ReadCount))
		col.ioWrite.Add(StoredValue(io.WriteCount))
		col.ioTotal.Add(StoredValue(io.ReadCount + io.WriteCount))
		col.ioReadBytes.Add(StoredValue(io.ReadBytes))
		col.ioWriteBytes.Add(StoredValue(io.WriteBytes))
		col.ioBytesTotal.Add(StoredValue(io.ReadBytes + io.WriteBytes))
	}
	return nil
}

func (col *SingleProcessCollector) updateMemory() error {
	// Alternative: col.MemoryInfoEx()
	if mem, err := col.MemoryInfo(); err != nil {
		return fmt.Errorf("Failed to get memory info: %v", err)
	} else {
		col.mem_rss = mem.RSS
		col.mem_vms = mem.VMS
		col.mem_swap = mem.Swap
	}
	return nil
}

func (col *SingleProcessCollector) updateNet() error {
	// Alternative: col.Connections()
	if counters, err := col.NetIOCounters(false); err != nil {
		return fmt.Errorf("Failed to get net-IO info: %v", err)
	} else {
		if len(counters) != 1 {
			return fmt.Errorf("gopsutil/process/Process.NetIOCounters() returned %v NetIOCountersStat instead of %v", len(counters), 1)
		}
		col.net.Add(&counters[0])
	}
	return nil
}

func (col *SingleProcessCollector) updateOpenFiles() error {
	// Alternative: col.NumFDs(), proc.OpenFiles()
	if num, err := col.procNumFds(); err != nil {
		return fmt.Errorf("Failed to get number of open files: %v", err)
	} else {
		col.numFds = num
	}
	return nil
}

func (col *SingleProcessCollector) updateMisc() error {
	// Misc, Alternatice: col.NumThreads(), col.NumCtxSwitches()
	if numThreads, ctxSwitches, err := col.procGetMisc(); err != nil {
		return fmt.Errorf("Failed to get number of threads/ctx-switches: %v", err)
	} else {
		col.numThreads = numThreads
		col.ctxSwitchVoluntary.Add(StoredValue(ctxSwitches.Voluntary))
		col.ctxSwitchInvoluntary.Add(StoredValue(ctxSwitches.Involuntary))
	}
	return nil
}

func (col *SingleProcessCollector) procNumFds() (int32, error) {
	// This is part of gopsutil/process.Process.fillFromfd()
	pid := col.Pid
	statPath := hostProcFile(strconv.Itoa(int(pid)), "fd")
	d, err := os.Open(statPath)
	if err != nil {
		return 0, err
	}
	defer d.Close()
	fnames, err := d.Readdirnames(-1)
	numFDs := int32(len(fnames))
	return numFDs, err
}

func (col *SingleProcessCollector) procGetMisc() (numThreads int32, numCtxSwitches process.NumCtxSwitchesStat, err error) {
	// This is part of gopsutil/process.Process.fillFromStatus()
	pid := col.Pid
	statPath := hostProcFile(strconv.Itoa(int(pid)), "status")
	var contents []byte
	contents, err = ioutil.ReadFile(statPath)
	if err != nil {
		return
	}
	lines := strings.Split(string(contents), "\n")
	leftover_fields := 3
	for _, line := range lines {
		tabParts := strings.SplitN(line, "\t", 2)
		if len(tabParts) < 2 {
			continue
		}
		value := tabParts[1]
		var v int64
		switch strings.TrimRight(tabParts[0], ":") {
		case "Threads":
			v, err = strconv.ParseInt(value, 10, 32)
			if err != nil {
				return
			}
			numThreads = int32(v)
			leftover_fields--
		case "voluntary_ctxt_switches":
			v, err = strconv.ParseInt(value, 10, 64)
			if err != nil {
				return
			}
			numCtxSwitches.Voluntary = v
			leftover_fields--
		case "nonvoluntary_ctxt_switches":
			v, err = strconv.ParseInt(value, 10, 64)
			if err != nil {
				return
			}
			numCtxSwitches.Involuntary = v
			leftover_fields--
		}
		if leftover_fields <= 0 {
			return
		}
	}
	return
}
