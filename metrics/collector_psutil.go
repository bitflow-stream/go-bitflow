package metrics

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"runtime"
	"strings"
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

// ==================== CPU ====================
type PsutilCpuCollector struct {
	AbstractCollector
	ring ValueRing
}

func (col *PsutilCpuCollector) Init() error {
	col.Reset(col)
	col.ring = NewValueRing(CpuTimeLogback)
	col.readers = map[string]MetricReader{
		"cpu": col.readCpu,
	}
	return nil
}

func (col *PsutilCpuCollector) Update() (err error) {
	times, err := cpu.CPUTimes(false)
	if err == nil {
		if len(times) != 1 {
			err = fmt.Errorf("warning: gopsutil/cpu.CPUTimes() returned %v CPUTimes instead of %v", len(times), 1)
		} else {
			col.ring.Add(&cpuTime{times[0]})
			col.UpdateMetrics()
		}
	}
	return
}

func (col *PsutilCpuCollector) readCpu() Value {
	return col.ring.GetDiff(CpuInterval)
}

type cpuTime struct {
	cpu.CPUTimesStat
}

func (t *cpuTime) getAllBusy() (float64, float64) {
	busy := t.User + t.System + t.Nice + t.Iowait + t.Irq +
		t.Softirq + t.Steal + t.Guest + t.GuestNice + t.Stolen
	return busy + t.Idle, busy
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
			cpu.CPUTimesStat{
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
	load *load.LoadAvgStat
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
	col.load, err = load.LoadAvg()
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
	bytes   ValueRing
	packets ValueRing
	errors  ValueRing
	dropped ValueRing
}

func (col *PsutilNetCollector) Init() error {
	col.Reset(col)
	col.counters = newNetIoCounters()
	// TODO separate in/out metrics
	col.readers = map[string]MetricReader{
		"net-io/bytes":   col.counters.readBytes,
		"net-io/packets": col.counters.readPackets,
		"net-io/errors":  col.counters.readErrors,
		"net-io/dropped": col.counters.readDropped,
	}
	return nil
}

func (col *PsutilNetCollector) Update() (err error) {
	counters, err := psnet.NetIOCounters(false)
	if err == nil && len(counters) != 1 {
		err = fmt.Errorf("gopsutil/net.NetIOCounters() returned %v NetIOCountersStat instead of %v", len(counters), 1)
	}
	if err == nil {
		col.counters.Add(&counters[0])
		col.UpdateMetrics()
	}
	return
}

func newNetIoCounters() netIoCounters {
	return netIoCounters{
		bytes:   NewValueRing(NetIoLogback),
		packets: NewValueRing(NetIoLogback),
		errors:  NewValueRing(NetIoLogback),
		dropped: NewValueRing(NetIoLogback),
	}
}

func (counters *netIoCounters) Add(stat *psnet.NetIOCountersStat) {
	counters.AddToHead(stat)
	counters.FlushHead()
}

func (counters *netIoCounters) AddToHead(stat *psnet.NetIOCountersStat) {
	counters.bytes.AddToHead(Value(stat.BytesSent + stat.BytesRecv))
	counters.packets.AddToHead(Value(stat.PacketsSent + stat.PacketsRecv))
	counters.errors.AddToHead(Value(stat.Errin + stat.Errout))
	counters.dropped.AddToHead(Value(stat.Dropin + stat.Dropout))
}

func (counters *netIoCounters) FlushHead() {
	counters.bytes.FlushHead()
	counters.packets.FlushHead()
	counters.errors.FlushHead()
	counters.dropped.FlushHead()
}

func (counters *netIoCounters) readBytes() Value {
	return counters.bytes.GetDiff(NetIoInterval)
}

func (counters *netIoCounters) readPackets() Value {
	return counters.packets.GetDiff(NetIoInterval)
}

func (counters *netIoCounters) readErrors() Value {
	return counters.errors.GetDiff(NetIoInterval)
}

func (counters *netIoCounters) readDropped() Value {
	return counters.dropped.GetDiff(NetIoInterval)
}

// ==================== Net Protocol Counters ====================
type PsutilNetProtoCollector struct {
	AbstractCollector
	protocols map[string]psnet.NetProtoCountersStat
}

func (col *PsutilNetProtoCollector) Init() error {
	col.Reset(col)
	col.protocols = make(map[string]psnet.NetProtoCountersStat)

	// TODO missing: metrics about individual connections and NICs
	if err := col.update(false); err != nil {
		return err
	}
	col.readers = make(map[string]MetricReader)
	for proto, counters := range col.protocols {
		for statName, _ := range counters.Stats {
			name := "net-proto/" + proto + "/" + statName
			col.readers[name] = (&protoStatReader{
				col:      col,
				protocol: proto,
				field:    statName,
			}).read
		}
	}
	return nil
}

func (col *PsutilNetProtoCollector) update(checkChange bool) error {
	counters, err := psnet.NetProtoCounters(nil)
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
		col.UpdateMetrics()
	}
	return
}

type protoStatReader struct {
	col      *PsutilNetProtoCollector
	protocol string
	field    string
}

func (reader *protoStatReader) read() Value {
	if counters, ok := reader.col.protocols[reader.protocol]; ok {
		if val, ok := counters.Stats[reader.field]; ok {
			return Value(val)
		} else {
			log.Printf("Warning: Counter %v not found in protocol %v in PsutilNetProtoCollector\n", reader.field, reader.protocol)
		}
	} else {
		log.Printf("Warning: Protocol %v not found in PsutilNetProtoCollector\n", reader.protocol)
	}
	return Value(0)
}

// ==================== Disk IO ====================
type PsutilDiskIOCollector struct {
	AbstractCollector
	disks map[string]disk.DiskIOCountersStat
}

func (col *PsutilDiskIOCollector) Init() error {
	col.Reset(col)
	col.disks = make(map[string]disk.DiskIOCountersStat)

	if err := col.update(false); err != nil {
		return err
	}
	col.readers = make(map[string]MetricReader)
	for disk, _ := range col.disks {
		name := "disk-io/" + disk + "/"
		reader := &diskIOReader{
			col:            col,
			disk:           disk,
			readRing:       NewValueRing(DiskIoLogback),
			writeRing:      NewValueRing(DiskIoLogback),
			readBytesRing:  NewValueRing(DiskIoLogback),
			writeBytesRing: NewValueRing(DiskIoLogback),
			readTimeRing:   NewValueRing(DiskIoLogback),
			writeTimeRing:  NewValueRing(DiskIoLogback),
			ioTimeRing:     NewValueRing(DiskIoLogback),
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
	disks, err := disk.DiskIOCounters()
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

func (reader *diskIOReader) checkDisk() *disk.DiskIOCountersStat {
	if disk, ok := reader.col.disks[reader.disk]; ok {
		return &disk
	} else {
		log.Printf("Warning: disk-io counters for disk %v not found\n", reader.disk)
		return nil
	}
}

func (reader *diskIOReader) value(val uint64, ring *ValueRing) Value {
	ring.Add(Value(val))
	return ring.GetDiff(DiskIoInterval)
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
	usage              map[string]*disk.DiskUsageStat
}

func (col *PsutilDiskUsageCollector) Init() error {
	col.Reset(col)
	col.usage = make(map[string]*disk.DiskUsageStat)
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
	partitions, err := disk.DiskPartitions(true)
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
		usage, err := disk.DiskUsage(partition)
		if err != nil {
			return err
		}
		col.usage[partition] = usage
	}
	return nil
}

func (col *PsutilDiskUsageCollector) checkChangedPartitions() error {
	partitions, err := disk.DiskPartitions(true)
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

func (reader *diskUsageReader) checkDisk() *disk.DiskUsageStat {
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
	pids        map[int32]*process.Process

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
	col.cpu = NewValueRing(CpuTimeLogback)
	col.ioRead = NewValueRing(DiskIoLogback)
	col.ioWrite = NewValueRing(DiskIoLogback)
	col.ioReadBytes = NewValueRing(DiskIoLogback)
	col.ioWriteBytes = NewValueRing(DiskIoLogback)
	col.ctx_switch_voluntary = NewValueRing(CtxSwitchLogback)
	col.ctx_switch_involuntary = NewValueRing(CtxSwitchLogback)
	col.net = newNetIoCounters()

	col.readers = map[string]MetricReader{
		"proc/" + col.GroupName + "/num": col.readNumProc,
		"proc/" + col.GroupName + "/cpu": col.readCpu,

		"proc/" + col.GroupName + "/disk/read":       col.readDiskRead,
		"proc/" + col.GroupName + "/disk/write":      col.readDiskWrite,
		"proc/" + col.GroupName + "/disk/readBytes":  col.readDiskReadBytes,
		"proc/" + col.GroupName + "/disk/writeBytes": col.readDiskWriteBytes,

		"proc/" + col.GroupName + "/net-io/bytes":   col.net.readBytes,
		"proc/" + col.GroupName + "/net-io/packets": col.net.readPackets,
		"proc/" + col.GroupName + "/net-io/errors":  col.net.readErrors,
		"proc/" + col.GroupName + "/net-io/dropped": col.net.readDropped,

		"proc/" + col.GroupName + "/ctxSwitch/voluntary":   col.readCtxSwitchVol,
		"proc/" + col.GroupName + "/ctxSwitch/involuntary": col.readCtxSwitchInvol,

		"proc/" + col.GroupName + "/mem/rss":  col.readMemRss,
		"proc/" + col.GroupName + "/mem/vms":  col.readMemVms,
		"proc/" + col.GroupName + "/mem/swap": col.readMemSwap,
		"proc/" + col.GroupName + "/fds":      col.readFds,
		"proc/" + col.GroupName + "/threads":  col.readThreads,
	}
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

	col.pids = make(map[int32]*process.Process)
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
				col.pids[pid] = proc
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
		err := col.updateProcess(proc)
		if err != nil {
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

func (col *PsutilProcessCollector) updateProcess(proc *process.Process) error {
	// CPU
	if cpu, err := proc.CPUTimes(); err != nil {
		return err
	} else {
		busy := (cpu.Total() - cpu.Idle) * col.cpu_factor
		col.cpu.AddToHead(Value(busy))
	}

	// Disk
	if io, err := proc.IOCounters(); err != nil {
		return err
	} else {
		col.ioRead.AddToHead(Value(io.ReadCount))
		col.ioWrite.AddToHead(Value(io.WriteCount))
		col.ioReadBytes.AddToHead(Value(io.ReadBytes))
		col.ioWriteBytes.AddToHead(Value(io.WriteBytes))
	}

	// Memory, Alternative: proc.MemoryInfoEx()
	if mem, err := proc.MemoryInfo(); err != nil {
		return err
	} else {
		col.mem_rss += mem.RSS
		col.mem_vms += mem.VMS
		col.mem_swap += mem.Swap
	}

	// Network, Alternative: proc.Connections()
	if counters, err := proc.NetIOCounters(false); err != nil {
		return err
	} else {
		if len(counters) != 1 {
			return fmt.Errorf("gopsutil/process/Process.NetIOCounters() returned %v NetIOCountersStat instead of %v", len(counters), 1)
		}
		col.net.AddToHead(&counters[0])
	}

	// Misc, Alternative: proc.OpenFiles()
	if num, err := proc.NumFDs(); err != nil {
		return err
	} else {
		col.numFds += num
	}
	if num, err := proc.NumThreads(); err != nil {
		return err
	} else {
		col.numThreads += num
	}
	if ctxSwitches, err := proc.NumCtxSwitches(); err != nil {
		return err
	} else {
		col.ctx_switch_voluntary.AddToHead(Value(ctxSwitches.Voluntary))
		col.ctx_switch_involuntary.AddToHead(Value(ctxSwitches.Involuntary))
	}

	return nil
}

func (col *PsutilProcessCollector) readNumProc() Value {
	return Value(len(col.pids))
}

func (col *PsutilProcessCollector) readCpu() Value {
	return col.cpu.GetDiff(CpuInterval)
}

func (col *PsutilProcessCollector) readDiskRead() Value {
	return col.ioRead.GetDiff(DiskIoInterval)
}

func (col *PsutilProcessCollector) readDiskWrite() Value {
	return col.ioWrite.GetDiff(DiskIoInterval)
}

func (col *PsutilProcessCollector) readDiskReadBytes() Value {
	return col.ioReadBytes.GetDiff(DiskIoInterval)
}

func (col *PsutilProcessCollector) readDiskWriteBytes() Value {
	return col.ioWriteBytes.GetDiff(DiskIoInterval)
}

func (col *PsutilProcessCollector) readCtxSwitchVol() Value {
	return col.ctx_switch_voluntary.GetDiff(CtxSwitchInterval)
}

func (col *PsutilProcessCollector) readCtxSwitchInvol() Value {
	return col.ctx_switch_involuntary.GetDiff(CtxSwitchInterval)
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
