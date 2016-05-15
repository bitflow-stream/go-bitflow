package metrics

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
	psnet "github.com/shirou/gopsutil/net"
	"github.com/shirou/gopsutil/process"
)

const (
	NetIoLogback      = 50
	NetIoInterval     = 1 * time.Second
	CpuTimeLogback    = 10
	CpuInterval       = 1 * time.Second
	DiskIoLogback     = 50
	DiskIoInterval    = 1 * time.Second
	CtxSwitchLogback  = 50
	CtxSwitchInterval = 1 * time.Second
	NetProtoLogback   = 50
	NetProtoInterval  = 1 * time.Second
)

func RegisterPsutilCollectors() {
	RegisterCollector(new(PsutilMemCollector))
	RegisterCollector(new(PsutilCpuCollector))
	RegisterCollector(new(PsutilLoadCollector))
	RegisterCollector(new(PsutilNetCollector))
	RegisterCollector(new(PsutilNetProtoCollector))
	RegisterCollector(new(PsutilDiskIOCollector))
	RegisterCollector(new(PsutilDiskUsageCollector))
	RegisterCollector(new(PsutilMiscCollector))
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

func (col *PsutilMemCollector) readFreeMem() Value {
	return Value(col.memory.Available)
}

func (col *PsutilMemCollector) readUsedMem() Value {
	return Value(col.memory.Used)
}

func (col *PsutilMemCollector) readUsedPercentMem() Value {
	return Value(col.memory.UsedPercent)
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
	ring ValueRing
}

func (col *PsutilCpuCollector) Init() error {
	col.Reset(col)
	col.ring = NewValueRing(CpuTimeLogback, CpuInterval)
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

func (t *cpuTime) DiffValue(logback LogbackValue, d time.Duration) Value {
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
		return Value((t2Busy - t1Busy) / (t2All - t1All) * 100)
	} else {
		log.Printf("Error: Cannot diff %v (%T) and %v (%T)\n", t, t, logback, logback)
		return Value(0)
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
		log.Printf("Error: Cannot add %v (%T) and %v (%T)\n", t, t, incoming, incoming)
		return Value(0)
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

func (col *PsutilLoadCollector) readLoad1() Value {
	return Value(col.load.Load1)
}

func (col *PsutilLoadCollector) readLoad5() Value {
	return Value(col.load.Load5)
}

func (col *PsutilLoadCollector) readLoad15() Value {
	return Value(col.load.Load15)
}

// ==================== Net IO Counters ====================
type PsutilNetCollector struct {
	AbstractCollector
	counters netIoCounters
}

type netIoCounters struct {
	bytes      ValueRing
	packets    ValueRing
	rx_bytes   ValueRing
	rx_packets ValueRing
	tx_bytes   ValueRing
	tx_packets ValueRing
	errors     ValueRing
	dropped    ValueRing
}

func (col *PsutilNetCollector) Init() error {
	col.Reset(col)
	col.counters = NewNetIoCounters(NetIoLogback, NetIoInterval)
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

func NewNetIoCounters(logback int, interval time.Duration) netIoCounters {
	return netIoCounters{
		bytes:      NewValueRing(logback, interval),
		packets:    NewValueRing(logback, interval),
		rx_bytes:   NewValueRing(logback, interval),
		rx_packets: NewValueRing(logback, interval),
		tx_bytes:   NewValueRing(logback, interval),
		tx_packets: NewValueRing(logback, interval),
		errors:     NewValueRing(logback, interval),
		dropped:    NewValueRing(logback, interval),
	}
}

func (counters *netIoCounters) Add(stat *psnet.IOCountersStat) {
	counters.AddToHead(stat)
	counters.FlushHead()
}

func (counters *netIoCounters) AddToHead(stat *psnet.IOCountersStat) {
	counters.bytes.AddToHead(Value(stat.BytesSent + stat.BytesRecv))
	counters.packets.AddToHead(Value(stat.PacketsSent + stat.PacketsRecv))
	counters.rx_bytes.AddToHead(Value(stat.BytesRecv))
	counters.rx_packets.AddToHead(Value(stat.PacketsRecv))
	counters.tx_bytes.AddToHead(Value(stat.BytesSent))
	counters.tx_packets.AddToHead(Value(stat.PacketsSent))
	counters.errors.AddToHead(Value(stat.Errin + stat.Errout))
	counters.dropped.AddToHead(Value(stat.Dropin + stat.Dropout))
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
				ringVal := NewValueRing(NetProtoLogback, NetProtoInterval)
				ring = &ringVal
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
	value Value
}

func (reader *protoStatReader) update() error {
	if counters, ok := reader.col.protocols[reader.protocol]; ok {
		if val, ok := counters.Stats[reader.field]; ok {
			if reader.ring != nil {
				reader.ring.Add(Value(val))
			} else {
				reader.value = Value(val)
			}
			return nil
		} else {
			return fmt.Errorf("Counter %v not found in protocol %v in PsutilNetProtoCollector\n", reader.field, reader.protocol)
		}
	} else {
		return fmt.Errorf("Protocol %v not found in PsutilNetProtoCollector\n", reader.protocol)
	}
}

func (reader *protoStatReader) read() Value {
	if ring := reader.ring; ring != nil {
		return ring.GetDiff()
	} else {
		return reader.value
	}
}

// ==================== Disk IO ====================
type PsutilDiskIOCollector struct {
	AbstractCollector
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
			readRing:       NewValueRing(DiskIoLogback, DiskIoInterval),
			writeRing:      NewValueRing(DiskIoLogback, DiskIoInterval),
			readBytesRing:  NewValueRing(DiskIoLogback, DiskIoInterval),
			writeBytesRing: NewValueRing(DiskIoLogback, DiskIoInterval),
			readTimeRing:   NewValueRing(DiskIoLogback, DiskIoInterval),
			writeTimeRing:  NewValueRing(DiskIoLogback, DiskIoInterval),
			ioTimeRing:     NewValueRing(DiskIoLogback, DiskIoInterval),
		}
		col.readers[name+"read"] = reader.readRead
		col.readers[name+"write"] = reader.readWrite
		col.readers[name+"readBytes"] = reader.readReadBytes
		col.readers[name+"writeBytes"] = reader.readWriteBytes
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

	readRing       ValueRing
	writeRing      ValueRing
	readBytesRing  ValueRing
	writeBytesRing ValueRing
	readTimeRing   ValueRing
	writeTimeRing  ValueRing
	ioTimeRing     ValueRing
}

func (reader *diskIOReader) checkDisk() *disk.IOCountersStat {
	if disk, ok := reader.col.disks[reader.disk]; ok {
		return &disk
	} else {
		log.Printf("Warning: disk-io counters for disk %v not found\n", reader.disk)
		return nil
	}
}

func (reader *diskIOReader) value(val uint64, ring *ValueRing) Value {
	ring.Add(Value(val))
	return ring.GetDiff()
}

func (reader *diskIOReader) readRead() Value {
	if disk := reader.checkDisk(); disk != nil {
		return reader.value(disk.ReadCount, &reader.readRing)
	}
	return Value(0)
}

func (reader *diskIOReader) readWrite() Value {
	if disk := reader.checkDisk(); disk != nil {
		return reader.value(disk.WriteCount, &reader.writeRing)
	}
	return Value(0)
}

func (reader *diskIOReader) readReadBytes() Value {
	if disk := reader.checkDisk(); disk != nil {
		return reader.value(disk.ReadBytes, &reader.readBytesRing)
	}
	return Value(0)
}

func (reader *diskIOReader) readWriteBytes() Value {
	if disk := reader.checkDisk(); disk != nil {
		return reader.value(disk.WriteBytes, &reader.writeBytesRing)
	}
	return Value(0)
}

func (reader *diskIOReader) readReadTime() Value {
	if disk := reader.checkDisk(); disk != nil {
		return reader.value(disk.ReadTime, &reader.readTimeRing)
	}
	return Value(0)
}

func (reader *diskIOReader) readWriteTime() Value {
	if disk := reader.checkDisk(); disk != nil {
		return reader.value(disk.WriteTime, &reader.writeTimeRing)
	}
	return Value(0)
}

func (reader *diskIOReader) readIoTime() Value {
	if disk := reader.checkDisk(); disk != nil {
		return reader.value(disk.IoTime, &reader.ioTimeRing)
	}
	return Value(0)
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
		log.Printf("Warning: disk-usage counters for partition %v not found\n", reader.partition)
		return nil
	}
}

func (reader *diskUsageReader) readFree() Value {
	if disk := reader.checkDisk(); disk != nil {
		return Value(disk.Free)
	}
	return Value(0)
}

func (reader *diskUsageReader) readPercent() Value {
	if disk := reader.checkDisk(); disk != nil {
		return Value(disk.UsedPercent)
	}
	return Value(0)
}

// ==================== Misc OS Metrics ====================
// Global information updated regularly
var osInformation struct {
	pids []int32
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

func (col *PsutilMiscCollector) Update() (err error) {
	pids, err := process.Pids()
	if err != nil {
		return err
	}
	osInformation.pids = pids
	if err == nil {
		col.UpdateMetrics()
	}
	return
}

func (col *PsutilMiscCollector) readNumProcs() Value {
	return Value(len(osInformation.pids))
}

// ==================== Process Metrics ====================
type PsutilProcessCollector struct {
	AbstractCollector

	// Settings
	CmdlineFilter     []*regexp.Regexp
	GroupName         string
	PrintErrors       bool
	PidUpdateInterval time.Duration

	pidsUpdated bool
	own_pid     int32
	cpu_factor  float64
	pids        map[int32]*SingleProcessCollector

	updateLock             sync.Mutex
	cpu                    ValueRing
	ioRead                 ValueRing
	ioWrite                ValueRing
	ioReadBytes            ValueRing
	ioWriteBytes           ValueRing
	ctx_switch_voluntary   ValueRing
	ctx_switch_involuntary ValueRing
	net                    netIoCounters
	mem_rss                uint64
	mem_vms                uint64
	mem_swap               uint64
	numFds                 int32
	numThreads             int32
}

func (col *PsutilProcessCollector) logErr(pid int32, err error) {
	if err != nil && col.PrintErrors {
		log.Printf("Error getting info about %s process %v: %v\n", col.GroupName, pid, err)
	}
}

func (col *PsutilProcessCollector) Init() error {
	col.own_pid = int32(os.Getpid())
	col.cpu_factor = 100 / float64(runtime.NumCPU())
	col.Reset(col)
	col.cpu = NewValueRing(CpuTimeLogback, CpuInterval)
	col.ioRead = NewValueRing(DiskIoLogback, DiskIoInterval)
	col.ioWrite = NewValueRing(DiskIoLogback, DiskIoInterval)
	col.ioReadBytes = NewValueRing(DiskIoLogback, DiskIoInterval)
	col.ioWriteBytes = NewValueRing(DiskIoLogback, DiskIoInterval)
	col.ctx_switch_voluntary = NewValueRing(CtxSwitchLogback, CtxSwitchInterval)
	col.ctx_switch_involuntary = NewValueRing(CtxSwitchLogback, CtxSwitchInterval)
	col.net = NewNetIoCounters(NetIoLogback, NetIoInterval)

	col.readers = map[string]MetricReader{
		"proc/" + col.GroupName + "/num": col.readNumProc,
		"proc/" + col.GroupName + "/cpu": col.cpu.GetDiff,

		"proc/" + col.GroupName + "/disk/read":       col.ioRead.GetDiff,
		"proc/" + col.GroupName + "/disk/write":      col.ioWrite.GetDiff,
		"proc/" + col.GroupName + "/disk/readBytes":  col.ioReadBytes.GetDiff,
		"proc/" + col.GroupName + "/disk/writeBytes": col.ioWriteBytes.GetDiff,

		"proc/" + col.GroupName + "/ctxSwitch/voluntary":   col.ctx_switch_voluntary.GetDiff,
		"proc/" + col.GroupName + "/ctxSwitch/involuntary": col.ctx_switch_involuntary.GetDiff,

		"proc/" + col.GroupName + "/mem/rss":  col.readMemRss,
		"proc/" + col.GroupName + "/mem/vms":  col.readMemVms,
		"proc/" + col.GroupName + "/mem/swap": col.readMemSwap,
		"proc/" + col.GroupName + "/fds":      col.readFds,
		"proc/" + col.GroupName + "/threads":  col.readThreads,
	}
	col.net.Register(col.readers, "proc/"+col.GroupName+"/net-io")

	return nil
}

func (col *PsutilProcessCollector) Update() (err error) {
	if err := col.updatePids(); err != nil {
		return err
	}
	col.updateValues()
	if err == nil {
		col.UpdateMetrics()
	}
	return
}

func (col *PsutilProcessCollector) updatePids() error {
	if col.pidsUpdated {
		return nil
	}

	col.pids = make(map[int32]*SingleProcessCollector)
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
			col.logErr(pid, err)
			continue
		}
		cmdline, err := proc.Cmdline()
		if err != nil {
			// Probably a permission error
			errors++
			col.logErr(pid, err)
			continue
		}
		for _, regex := range col.CmdlineFilter {
			if regex.MatchString(cmdline) {
				col.pids[pid] = MakeProcessCollector(col, proc)
				break
			}
		}
	}
	if len(col.pids) == 0 && errors > 0 && col.PrintErrors {
		col.logErr(-1, fmt.Errorf("Warning: Observing no processes, failed to check %v out of %v PIDs.\n",
			errors, len(pids)))
	}

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

func (col *PsutilProcessCollector) updateValues() {
	col.numFds = 0
	col.numThreads = 0
	col.mem_rss = 0
	col.mem_vms = 0
	col.mem_swap = 0

	for pid, proc := range col.pids {
		if err := proc.update(); err != nil {
			// Process probably does not exist anymore
			delete(col.pids, pid)
			col.logErr(pid, err)
		}
	}

	col.cpu.FlushHead()
	col.ioRead.FlushHead()
	col.ioWrite.FlushHead()
	col.ioReadBytes.FlushHead()
	col.ioWriteBytes.FlushHead()
	col.ctx_switch_voluntary.FlushHead()
	col.ctx_switch_involuntary.FlushHead()
	col.net.FlushHead()
}

func (col *PsutilProcessCollector) safeUpdate(update func()) {
	col.updateLock.Lock()
	defer col.updateLock.Unlock()
	update()
}

type SingleProcessCollector struct {
	*PsutilProcessCollector
	*process.Process
	tasks CollectorTasks
}

func MakeProcessCollector(collector *PsutilProcessCollector, proc *process.Process) *SingleProcessCollector {
	col := &SingleProcessCollector{
		PsutilProcessCollector: collector,
		Process:                proc,
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
		busy := (cpu.Total() - cpu.Idle) * col.cpu_factor
		col.safeUpdate(func() {
			col.cpu.AddToHead(Value(busy))
		})
	}
	return nil
}

func (col *SingleProcessCollector) updateDisk() error {
	if io, err := col.IOCounters(); err != nil {
		return fmt.Errorf("Failed to get disk-IO info: %v", err)
	} else {
		col.safeUpdate(func() {
			col.ioRead.AddToHead(Value(io.ReadCount))
			col.ioWrite.AddToHead(Value(io.WriteCount))
			col.ioReadBytes.AddToHead(Value(io.ReadBytes))
			col.ioWriteBytes.AddToHead(Value(io.WriteBytes))
		})
	}
	return nil
}

func (col *SingleProcessCollector) updateMemory() error {
	// Alternative: col.MemoryInfoEx()
	if mem, err := col.MemoryInfo(); err != nil {
		return fmt.Errorf("Failed to get memory info: %v", err)
	} else {
		col.safeUpdate(func() {
			col.mem_rss += mem.RSS
			col.mem_vms += mem.VMS
			col.mem_swap += mem.Swap
		})
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
		col.safeUpdate(func() {
			col.net.AddToHead(&counters[0])
		})
	}
	return nil
}

func (col *SingleProcessCollector) updateOpenFiles() error {
	// Alternative: col.NumFDs(), proc.OpenFiles()
	if num, err := col.procNumFds(); err != nil {
		return fmt.Errorf("Failed to get number of open files: %v", err)
	} else {
		col.safeUpdate(func() {
			col.numFds += num
		})
	}
	return nil
}

func (col *SingleProcessCollector) updateMisc() error {
	// Misc, Alternatice: col.NumThreads(), col.NumCtxSwitches()
	if numThreads, ctxSwitches, err := col.procGetMisc(); err != nil {
		return fmt.Errorf("Failed to get number of threads/ctx-switches: %v", err)
	} else {
		col.safeUpdate(func() {
			col.numThreads += numThreads
			col.ctx_switch_voluntary.AddToHead(Value(ctxSwitches.Voluntary))
			col.ctx_switch_involuntary.AddToHead(Value(ctxSwitches.Involuntary))
		})
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

func (col *PsutilProcessCollector) readNumProc() Value {
	return Value(len(col.pids))
}

func (col *PsutilProcessCollector) readMemRss() Value {
	return Value(col.mem_rss)
}

func (col *PsutilProcessCollector) readMemVms() Value {
	return Value(col.mem_vms)
}

func (col *PsutilProcessCollector) readMemSwap() Value {
	return Value(col.mem_swap)
}

func (col *PsutilProcessCollector) readFds() Value {
	return Value(col.numFds)
}

func (col *PsutilProcessCollector) readThreads() Value {
	return Value(col.numThreads)
}
