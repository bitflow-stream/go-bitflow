package metrics

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
	psnet "github.com/shirou/gopsutil/net"
)

const (
	NetIoLogback   = 50
	NetIoInterval  = 1 * time.Second
	CpuTimeLogback = 10
	CpuInterval    = 1 * time.Second
	DiskIoLogback  = 50
	DiskIoInterval = 1 * time.Second
)

func init() {
	RegisterCollector(new(PsutilMemCollector))
	RegisterCollector(new(PsutilCpuCollector))
	RegisterCollector(new(PsutilLoadCollector))
	RegisterCollector(new(PsutilNetCollector))
	RegisterCollector(new(PsutilNetProtoCollector))
	RegisterCollector(new(PsutilDiskIOCollector))
	RegisterCollector(new(PsutilDiskUsageCollector))
}

type psutilMetric struct {
	*Metric
	psutilReader
}

type psutilReader func() Value

type PsutilCollector struct {
	metrics []*psutilMetric
	readers map[string]psutilReader // Must be filled in Init() implementations
}

func (col *PsutilCollector) SupportedMetrics() (res []string) {
	res = make([]string, 0, len(col.readers))
	for metric, _ := range col.readers {
		res = append(res, metric)
	}
	return
}

func (col *PsutilCollector) SupportsMetric(metric string) bool {
	_, ok := col.readers[metric]
	return ok
}

func (col *PsutilCollector) Collect(metric *Metric) error {
	tags := make([]string, 0, len(col.readers))
	for metricName, reader := range col.readers {
		if metric.Name == metricName {
			col.metrics = append(col.metrics, &psutilMetric{
				Metric:       metric,
				psutilReader: reader,
			})
			return nil
		}
		tags = append(tags, metric.Name)
	}
	return fmt.Errorf("Cannot handle metric %v, expected one of %v", metric.Name, tags)
}

func (col *PsutilCollector) updateMetrics() {
	for _, metric := range col.metrics {
		metric.Set(metric.psutilReader())
	}
}

func (col *PsutilCollector) String() string {
	return fmt.Sprintf("PsutilCollector collecting %v", col.metrics)
}

// ==================== Memory ====================
type PsutilMemCollector struct {
	PsutilCollector
	memory *mem.VirtualMemoryStat
}

func (col *PsutilMemCollector) Init() error {
	col.readers = map[string]psutilReader{
		"mem/free":    col.readFreeMem,
		"mem/used":    col.readUsedMem,
		"mem/percent": col.readUsedPercentMem,
	}
	return nil
}

func (col *PsutilMemCollector) Update() (err error) {
	col.memory, err = mem.VirtualMemory()
	if err == nil {
		col.updateMetrics()
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
	PsutilCollector
	ring ValueRing
}

func (col *PsutilCpuCollector) Init() error {
	col.ring = NewValueRing(CpuTimeLogback)
	col.readers = map[string]psutilReader{
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
			col.updateMetrics()
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

func (t *cpuTime) DiffValue(logback LogbackValue, _ time.Duration) Value {
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

// ==================== Load ====================
type PsutilLoadCollector struct {
	PsutilCollector
	load *load.LoadAvgStat
}

func (col *PsutilLoadCollector) Init() error {
	col.readers = map[string]psutilReader{
		"load/1":  col.readLoad1,
		"load/5":  col.readLoad5,
		"load/15": col.readLoad15,
	}
	return nil
}

func (col *PsutilLoadCollector) Update() (err error) {
	col.load, err = load.LoadAvg()
	if err == nil {
		col.updateMetrics()
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
	PsutilCollector
	bytes   ValueRing
	packets ValueRing
	errors  ValueRing
	dropped ValueRing
}

func (col *PsutilNetCollector) Init() error {
	col.bytes = NewValueRing(NetIoLogback)
	col.packets = NewValueRing(NetIoLogback)
	col.errors = NewValueRing(NetIoLogback)
	col.dropped = NewValueRing(NetIoLogback)
	// TODO separate in/out metrics
	col.readers = map[string]psutilReader{
		"net-io/bytes":   col.readBytes,
		"net-io/packets": col.readPackets,
		"net-io/errors":  col.readErrors,
		"net-io/dropped": col.readDropped,
	}
	return nil
}

func (col *PsutilNetCollector) Update() (err error) {
	counters, err := psnet.NetIOCounters(false)
	if err == nil && len(counters) != 1 {
		err = fmt.Errorf("gopsutil/net.NetIOCounters() returned %v NetIOCountersStat instead of %v", len(counters), 1)
	}
	if err == nil {
		stat := counters[0]
		col.bytes.Add(Value(stat.BytesSent + stat.BytesRecv))
		col.packets.Add(Value(stat.PacketsSent + stat.PacketsRecv))
		col.errors.Add(Value(stat.Errin + stat.Errout))
		col.dropped.Add(Value(stat.Dropin + stat.Dropout))
		col.updateMetrics()
	}
	return
}

func (col *PsutilNetCollector) readBytes() Value {
	return col.bytes.GetDiff(NetIoInterval)
}

func (col *PsutilNetCollector) readPackets() Value {
	return col.packets.GetDiff(NetIoInterval)
}

func (col *PsutilNetCollector) readErrors() Value {
	return col.errors.GetDiff(NetIoInterval)
}

func (col *PsutilNetCollector) readDropped() Value {
	return col.dropped.GetDiff(NetIoInterval)
}

// ==================== Net Protocol Counters ====================
type PsutilNetProtoCollector struct {
	PsutilCollector
	protocols map[string]psnet.NetProtoCountersStat
}

func (col *PsutilNetProtoCollector) Init() error {
	col.protocols = make(map[string]psnet.NetProtoCountersStat)

	// TODO missing: metrics about individual connections and NICs
	if err := col.update(); err != nil {
		return err
	}
	col.readers = make(map[string]psutilReader)
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

func (col *PsutilNetProtoCollector) update() error {
	counters, err := psnet.NetProtoCounters(nil)
	if err != nil {
		return err
	}
	for _, counters := range counters {
		col.protocols[counters.Protocol] = counters
	}
	return nil
}

func (col *PsutilNetProtoCollector) Update() error {
	err := col.update()
	if err == nil {
		col.updateMetrics()
	}
	return err
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
	PsutilCollector
	disks map[string]disk.DiskIOCountersStat
}

func (col *PsutilDiskIOCollector) Init() error {
	col.disks = make(map[string]disk.DiskIOCountersStat)

	if err := col.update(); err != nil {
		return err
	}
	col.readers = make(map[string]psutilReader)
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

func (col *PsutilDiskIOCollector) update() error {
	disks, err := disk.DiskIOCounters()
	if err != nil {
		return err
	}
	col.disks = disks
	return nil
}

func (col *PsutilDiskIOCollector) Update() error {
	err := col.update()
	if err == nil {
		col.updateMetrics()
	}
	return err
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
	PsutilCollector
	allPartitions      []string
	observedPartitions map[string]bool
	usage              map[string]*disk.DiskUsageStat
}

func (col *PsutilDiskUsageCollector) Init() error {
	col.usage = make(map[string]*disk.DiskUsageStat)
	col.observedPartitions = make(map[string]bool)

	var err error
	col.allPartitions, err = col.getAllPartitions()
	if err != nil {
		return err
	}

	col.readers = make(map[string]psutilReader)
	for _, partition := range col.allPartitions {
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
	return col.PsutilCollector.Collect(metric)
}

func (col *PsutilDiskUsageCollector) getAllPartitions() ([]string, error) {
	partitions, err := disk.DiskPartitions(true)
	if err != nil {
		return nil, err
	}
	result := make([]string, 0, len(partitions))
	for _, partition := range partitions {
		result = append(result, partition.Mountpoint)
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

func (col *PsutilDiskUsageCollector) Update() error {
	err := col.update()
	if err == nil {
		col.updateMetrics()
	}
	return err
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
