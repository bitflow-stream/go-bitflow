package plugin

import (
	"fmt"
	"plugin"

	"github.com/bitflow-stream/go-bitflow-pipeline/bitflow-script/reg"
	log "github.com/sirupsen/logrus"
)

const BitflowPluginSymbol = "Plugin"

type BitflowPlugin interface {
	Init(registry reg.ProcessorRegistry) error
	Name() string
}

func LoadPlugin(registry reg.ProcessorRegistry, path string) error {
	return LoadPluginSymbol(registry, path, BitflowPluginSymbol)
}

func LoadPluginSymbol(registry reg.ProcessorRegistry, path string, symbol string) error {
	log.Debugln("Loading plugin", path)
	openedPlugin, err := plugin.Open(path)
	if err != nil {
		return err
	}
	symbolObject, err := openedPlugin.Lookup(symbol)
	if err != nil {
		return err
	}
	sourcePlugin, ok := symbolObject.(*BitflowPlugin)
	if !ok || sourcePlugin == nil {
		return fmt.Errorf("Symbol '%v' from plugin '%v' has type %T instead of plugin.BitflowPlugin",
			symbol, path, symbolObject)
	}
	p := *sourcePlugin
	log.Debugf("Initializing plugin '%v' loaded from symbol '%v' in %v...", p.Name(), symbol, path)
	return p.Init(registry)
}
