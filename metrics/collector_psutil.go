package metrics

import (
	"fmt"
	"log"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
	psnet "github.com/shirou/gopsutil/net"
)

const (
	NetIoLogback   = 50
	NetIoInterval  = 1 * time.Second
	CpuTimeLogback = 10
	CpuInterval    = 1 * time.Second
)

type psutilMetric struct {
	*Metric
	psutilReader
}

type psutilReader func() Value

type PsutilCollector struct {
	metrics []*psutilMetric
}

func init() {
	registerCollector("mem", &PsutilMemCollector{})
	registerCollector("cpu", &PsutilCpuCollector{ring: NewValueRing(CpuTimeLogback)})
	registerCollector("net-io", &PsutilNetCollector{
		bytes:   NewValueRing(NetIoLogback),
		packets: NewValueRing(NetIoLogback),
		errors:  NewValueRing(NetIoLogback),
		dropped: NewValueRing(NetIoLogback),
	})
	//	registerCollector("net-proto", &PsutilNetProtoCollector{})
}

func (col *PsutilCollector) doCollect(metric *Metric, readers map[string]psutilReader) error {
	tags := make([]string, 0, len(readers))
	for tag, reader := range readers {
		tagStr := metric.Tag
		if tagStr == tag {
			col.metrics = append(col.metrics, &psutilMetric{
				Metric:       metric,
				psutilReader: reader,
			})
			return nil
		}
		tags = append(tags, tagStr)
	}
	return fmt.Errorf("Cannot handle tag %v, expected one of %v", metric.Tag, tags)
}

func (col *PsutilCollector) updateMetrics() {
	for _, metric := range col.metrics {
		metric.Val = metric.psutilReader()
	}
}

func (col *PsutilCollector) String() string {
	return fmt.Sprintf("PsutilCollector collecting %v", col.metrics)
}

// ==================== Memory ====================
type PsutilMemCollector struct {
	PsutilCollector
	stat *mem.VirtualMemoryStat
}

func (col *PsutilMemCollector) SupportedMetrics() []string {
	return []string{
		"mem/free",
		"mem/used",
		"mem/percent",
	}
}

func (col *PsutilMemCollector) Collect(metric *Metric) error {
	return col.doCollect(metric, map[string]psutilReader{
		"mem/free":    col.readFreeMem,
		"mem/used":    col.readUsedMem,
		"mem/percent": col.readUsedPercentMem,
	})
}

func (col *PsutilMemCollector) Update() (err error) {
	col.stat, err = mem.VirtualMemory()
	if err == nil {
		col.updateMetrics()
	}
	return
}

func (col *PsutilMemCollector) readFreeMem() Value {
	return Value(col.stat.Available)
}

func (col *PsutilMemCollector) readUsedMem() Value {
	return Value(col.stat.Used)
}

func (col *PsutilMemCollector) readUsedPercentMem() Value {
	return Value(col.stat.UsedPercent)
}

// ==================== CPU ====================
type PsutilCpuCollector struct {
	PsutilCollector
	ring ValueRing
}

func (col *PsutilCpuCollector) SupportedMetrics() []string {
	return []string{
		"cpu",
	}
}

func (col *PsutilCpuCollector) Collect(metric *Metric) error {
	return col.doCollect(metric, map[string]psutilReader{
		"cpu": col.readCpu,
	})
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
	stat *load.LoadAvgStat
}

func (col *PsutilLoadCollector) SupportedMetrics() []string {
	return []string{
		"load/1",
		"load/5",
		"load/15",
	}
}

func (col *PsutilLoadCollector) Collect(metric *Metric) error {
	return col.doCollect(metric, map[string]psutilReader{
		"load/1":  col.readLoad1,
		"load/5":  col.readLoad5,
		"load/15": col.readLoad15,
	})
}

func (col *PsutilLoadCollector) Update() (err error) {
	col.stat, err = load.LoadAvg()
	if err == nil {
		col.updateMetrics()
	}
	return
}

func (col *PsutilLoadCollector) readLoad1() Value {
	return Value(col.stat.Load1)
}

func (col *PsutilLoadCollector) readLoad5() Value {
	return Value(col.stat.Load5)
}

func (col *PsutilLoadCollector) readLoad15() Value {
	return Value(col.stat.Load15)
}

// ==================== Net IO Counters ====================
type PsutilNetCollector struct {
	PsutilCollector
	bytes   ValueRing
	packets ValueRing
	errors  ValueRing
	dropped ValueRing
}

func (col *PsutilNetCollector) SupportedMetrics() []string {
	// TODO separate in/out metrics
	return []string{
		"net-io/bytes",
		"net-io/packets",
		"net-io/errors",
		"net-io/dropped",
	}
}

func (col *PsutilNetCollector) Collect(metric *Metric) error {
	return col.doCollect(metric, map[string]psutilReader{
		"net-io/bytes":   col.readBytes,
		"net-io/packets": col.readPackets,
		"net-io/errors":  col.readErrors,
		"net-io/dropped": col.readDropped,
	})
}

func (col *PsutilNetCollector) Update() (err error) {
	stats, err := psnet.NetIOCounters(false)
	if err == nil && len(stats) != 1 {
		err = fmt.Errorf("gopsutil/net.NetIOCounters() returned %v NetIOCountersStat instead of %v", len(stats), 1)
	}
	if err == nil {
		stat := stats[0]
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
	stat map[string]*psnet.NetProtoCountersStat
}

type protoStatReader struct {
	col      *PsutilNetProtoCollector
	protocol string
	stat     string
}

func (col *PsutilNetProtoCollector) SupportedMetrics() []string {
	// TODO missing: metrics about individual connections and NICs
	if err := col.update(); err != nil {
		log.Println("Warning: Failed to update PsutilNetProtoCollector:", err)
		return nil
	}

	var res []string
	for proto, stat := range col.stat {
		for statName, _ := range stat.Stats {
			name := "net-proto/" + proto + "/" + statName
			res = append(res, name)
		}
	}
	return res
}

func (col *PsutilNetProtoCollector) Collect(metric *Metric) error {
	if err := col.update(); err != nil {
		return err
	}

	readers := make(map[string]psutilReader)
	for proto, stat := range col.stat {
		for statName, _ := range stat.Stats {
			name := "net-proto/" + proto + "/" + statName
			readers[name] = (&protoStatReader{
				col:      col,
				protocol: proto,
				stat:     statName,
			}).read
		}
	}
	return col.doCollect(metric, readers)
}

func (col *PsutilNetProtoCollector) update() error {
	stats, err := psnet.NetProtoCounters(nil)
	if err != nil {
		return err
	}
	col.stat = make(map[string]*psnet.NetProtoCountersStat)
	for _, stat := range stats {
		col.stat[stat.Protocol] = &stat
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

func (reader *protoStatReader) read() Value {
	if stat, ok := reader.col.stat[reader.protocol]; ok {
		if val, ok := stat.Stats[reader.stat]; ok {
			return Value(val)
		} else {
			log.Printf("Warning: Stat %v not found in protocol %v in PsutilNetProtoCollector\n", reader.stat, reader.protocol)
		}
	} else {
		log.Printf("Warning: Protocol %v not found in PsutilNetProtoCollector\n", reader.protocol)
	}
	return Value(0)
}
