package metrics

import (
	"fmt"
	"sync"
	"time"

	"github.com/socketplane/libovsdb"
)

const (
	OvsdbLogback  = 50
	OvsdbInterval = 5 * time.Second

	DefaultOvsdbPort = libovsdb.DEFAULT_PORT
)

func RegisterOvsdbCollector(Host string) {
	RegisterOvsdbCollectorPort(Host, 0)
}

func RegisterOvsdbCollectorPort(host string, port int) {
	RegisterCollector(&OvsdbCollector{Host: host, Port: port})
}

type OvsdbCollector struct {
	AbstractCollector
	Host string
	Port int

	client           *libovsdb.OvsdbClient
	lastUpdateError  error
	notifier         ovsdbNotifier
	interfaceReaders map[string]*ovsdbInterfaceReader
	readersLock      sync.Mutex
}

func (col *OvsdbCollector) Init() error {
	col.Reset(col)
	col.Close()
	col.notifier.col = col
	col.lastUpdateError = nil
	col.interfaceReaders = make(map[string]*ovsdbInterfaceReader)
	if err := col.update(false); err != nil {
		return err
	}

	col.readers = make(map[string]MetricReader)
	for _, reader := range col.interfaceReaders {
		base := "ovsdb/" + reader.name
		for name, reader := range map[string]MetricReader{
			base + "/errors":     reader.readErrors,
			base + "/dropped":    reader.readDropped,
			base + "/bytes":      reader.readBytes,
			base + "/packets":    reader.readPackets,
			base + "/rx_bytes":   reader.readRxBytes,
			base + "/rx_packets": reader.readRxPackets,
			base + "/tx_bytes":   reader.readTxBytes,
			base + "/tx_packets": reader.readTxPackets,
		} {
			col.readers[name] = reader
		}
	}
	return nil
}

func (col *OvsdbCollector) getReader(name string) *ovsdbInterfaceReader {
	if reader, ok := col.interfaceReaders[name]; ok {
		return reader
	}
	reader := &ovsdbInterfaceReader{
		col:        col,
		name:       name,
		errors:     NewValueRing(OvsdbLogback),
		dropped:    NewValueRing(OvsdbLogback),
		bytes:      NewValueRing(OvsdbLogback),
		packets:    NewValueRing(OvsdbLogback),
		rx_bytes:   NewValueRing(OvsdbLogback),
		rx_packets: NewValueRing(OvsdbLogback),
		tx_bytes:   NewValueRing(OvsdbLogback),
		tx_packets: NewValueRing(OvsdbLogback),
	}
	col.interfaceReaders[name] = reader
	return reader
}

func (col *OvsdbCollector) Close() {
	if client := col.client; client != nil {
		client.Disconnect()
		col.client = nil
	}
}

func (col *OvsdbCollector) Update() (err error) {
	if err = col.update(true); err == nil {
		col.UpdateMetrics()
	}
	return
}

func (col *OvsdbCollector) update(checkChange bool) error {
	if col.lastUpdateError != nil {
		col.Close()
		return col.lastUpdateError
	}
	if err := col.ensureConnection(checkChange); err != nil {
		return err
	}
	return nil
}

func (col *OvsdbCollector) ensureConnection(checkChange bool) error {
	if col.client == nil {
		initialTables, ovs, err := col.openConnection()
		if err == nil {
			col.client = ovs
			return col.updateTables(checkChange, initialTables.Updates)
		} else {
			return err
		}
	}
	return nil
}

func (col *OvsdbCollector) openConnection() (*libovsdb.TableUpdates, *libovsdb.OvsdbClient, error) {
	ovs, err := libovsdb.Connect(col.Host, col.Port)
	if err != nil {
		return nil, nil, err
	}
	ovs.Register(&col.notifier)

	// Request all updates for all Interface statistics
	requests := map[string]libovsdb.MonitorRequest{
		"Interface": libovsdb.MonitorRequest{
			Columns: []string{"name", "statistics"},
		},
	}

	initial, err := ovs.Monitor("Open_vSwitch", "", requests)
	if err != nil {
		ovs.Disconnect()
		return nil, nil, err
	}
	return initial, ovs, nil
}

func (col *OvsdbCollector) updateTables(checkChange bool, updates map[string]libovsdb.TableUpdate) error {
	update, ok := updates["Interface"]
	if !ok {
		return fmt.Errorf("OVSDB update did not contain requested table 'Interface'. Instead: %v", updates)
	}

	// TODO periodically check, if all monitored interfaces still exist
	col.readersLock.Lock()
	defer col.readersLock.Unlock()
	for _, rowUpdate := range update.Rows {
		if name, stats, err := col.parseRowUpdate(&rowUpdate.New); err != nil {
			return err
		} else {
			if checkChange {
				if _, ok := col.interfaceReaders[name]; !ok {
					return MetricsChanged
				}
			}
			col.getReader(name).update(stats)
		}
	}
	return nil
}

func (col *OvsdbCollector) parseRowUpdate(row *libovsdb.Row) (name string, stats map[string]float64, err error) {
	defer func() {
		// Allow panics for less explicit type checks
		if rec := recover(); rec != nil {
			err = fmt.Errorf("Parsing OVSDB row updated failed: %v", rec)
		}
	}()

	if nameObj, ok := row.Fields["name"]; !ok {
		err = fmt.Errorf("Row update did not include 'name' field")
		return
	} else {
		name = nameObj.(string)
	}
	if statsObj, ok := row.Fields["statistics"]; !ok {
		err = fmt.Errorf("Row update did not include 'statistics' field")
	} else {
		statMap := statsObj.(libovsdb.OvsMap)
		stats = make(map[string]float64)
		for keyObj, valObj := range statMap.GoMap {
			stats[keyObj.(string)] = valObj.(float64)
		}
	}
	return
}

// ==================== Interface Update Collector ====================

type ovsdbInterfaceReader struct {
	name string
	col  *OvsdbCollector

	errors     ValueRing
	dropped    ValueRing
	bytes      ValueRing
	packets    ValueRing
	rx_bytes   ValueRing
	rx_packets ValueRing
	tx_bytes   ValueRing
	tx_packets ValueRing
}

func (col *ovsdbInterfaceReader) fillValues(stats map[string]float64, names []string, ring *ValueRing) {
	for _, name := range names {
		if value, ok := stats[name]; ok {
			ring.AddToHead(Value(value))
		}
	}
	ring.FlushHead()
}

func (col *ovsdbInterfaceReader) update(stats map[string]float64) {
	col.fillValues(stats, []string{
		"collisions",
		"rx_crc_err",
		"rx_errors",
		"rx_frame_err",
		"rx_over_err",
		"tx_errors",
	}, &col.errors)
	col.fillValues(stats, []string{"rx_dropped", "tx_dropped"}, &col.dropped)
	col.fillValues(stats, []string{"rx_bytes", "tx_bytes"}, &col.bytes)
	col.fillValues(stats, []string{"rx_packets", "tx_packets"}, &col.packets)
	col.fillValues(stats, []string{"rx_bytes"}, &col.rx_bytes)
	col.fillValues(stats, []string{"rx_packets"}, &col.rx_packets)
	col.fillValues(stats, []string{"tx_bytes"}, &col.tx_bytes)
	col.fillValues(stats, []string{"tx_packets"}, &col.tx_packets)
}

func (col *ovsdbInterfaceReader) readErrors() Value {
	return col.errors.GetDiff(OvsdbInterval)
}

func (col *ovsdbInterfaceReader) readDropped() Value {
	return col.dropped.GetDiff(OvsdbInterval)
}

func (col *ovsdbInterfaceReader) readBytes() Value {
	return col.bytes.GetDiff(OvsdbInterval)
}

func (col *ovsdbInterfaceReader) readPackets() Value {
	return col.packets.GetDiff(OvsdbInterval)
}

func (col *ovsdbInterfaceReader) readRxBytes() Value {
	return col.rx_bytes.GetDiff(OvsdbInterval)
}

func (col *ovsdbInterfaceReader) readRxPackets() Value {
	return col.rx_packets.GetDiff(OvsdbInterval)
}

func (col *ovsdbInterfaceReader) readTxBytes() Value {
	return col.tx_bytes.GetDiff(OvsdbInterval)
}

func (col *ovsdbInterfaceReader) readTxPackets() Value {
	return col.tx_packets.GetDiff(OvsdbInterval)
}

// ==================== OVSDB Notifications ====================

type ovsdbNotifier struct {
	col *OvsdbCollector
}

func (n *ovsdbNotifier) Update(_ interface{}, tableUpdates libovsdb.TableUpdates) {
	// Note: Do not call n.col.client.Disconnect() from here (deadlock)
	if n.col.lastUpdateError != nil {
		return
	}
	if err := n.col.updateTables(true, tableUpdates.Updates); err != nil {
		n.col.lastUpdateError = err
	}
}
func (n *ovsdbNotifier) Locked([]interface{}) {
}
func (n *ovsdbNotifier) Stolen([]interface{}) {
}
func (n *ovsdbNotifier) Echo([]interface{}) {
}
func (n *ovsdbNotifier) Disconnected(client *libovsdb.OvsdbClient) {
	n.col.client = nil
}
