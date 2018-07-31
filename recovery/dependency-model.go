package recovery

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	log "github.com/sirupsen/logrus"
)

const (
	UnknownLayer    = Layer(0)
	HypervisorLayer = Layer(1)
	VMLayer         = Layer(2)
	ServiceLayer    = Layer(3)
)

type Layer int

type DependencyModel struct {
	Groups map[string][]*DependencyNode
	Names  map[string]*DependencyNode
	Ids    map[string]*DependencyNode
	Layers map[Layer][]*DependencyNode
}

type DependencyNode struct {
	Vars     map[string]string
	Name     string
	Id       string
	Layer    Layer
	Parent   *DependencyNode
	Children []*DependencyNode
	Groups   []string
}

func LoadDependencyModel(filePath string) (*DependencyModel, error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	return ParseDependencyModel(data)
}

func ParseDependencyModel(jsonData []byte) (*DependencyModel, error) {
	var jsonMap map[string]interface{}
	if err := json.Unmarshal(jsonData, &jsonMap); err != nil {
		return nil, err
	}
	res := new(DependencyModel)
	return res, res.FillFrom(jsonMap)
}

func (m *DependencyModel) FillFrom(root map[string]interface{}) (err error) {
	m.Names = make(map[string]*DependencyNode)
	m.Ids = make(map[string]*DependencyNode)
	m.Groups = make(map[string][]*DependencyNode)
	m.Layers = make(map[Layer][]*DependencyNode)

	// Use the panic mechanism to make the JSON traversal code more readable
	defer func() {
		if rec := recover(); rec != nil {
			err = fmt.Errorf("Error traversing dependency graph JSON data: %v", rec)
		}
	}()
	m.fillOrPanic(root)
	return
}

func (m *DependencyModel) fillOrPanic(root map[string]interface{}) {
	for name, objInterface := range root {
		obj := objInterface.(map[string]interface{})
		if strings.HasSuffix(name, "_vms") {
			// Intermediate group associating VMs with hypervisors
			hvId := name[0 : len(name)-len("_vms")]
			hv := m.getById(hvId)
			hv.setLayer(HypervisorLayer)
			for _, vmId := range obj["children"].([]interface{}) {
				vmNode := m.getById(vmId.(string))
				vmNode.setLayer(VMLayer)
				hv.addChild(vmNode)
			}
		} else if strings.HasSuffix(name, "_services") {
			// Intermediate group associating services with VMs
			vmId := name[0 : len(name)-len("_services")]
			vm := m.getById(vmId)
			vm.setLayer(VMLayer)
			for _, serviceId := range obj["children"].([]interface{}) {
				serviceNode := m.getById(serviceId.(string))
				serviceNode.setLayer(ServiceLayer)
				vm.addChild(serviceNode)
			}
		} else if childrenI, ok := obj["children"]; ok {
			// Named group of hosts
			m.resolveGroup(name, childrenI.([]interface{}), root)
		} else {
			// Regular node
			hosts := obj["hosts"].([]interface{})
			if len(hosts) > 1 {
				log.Warnf("Dependency-Model: Ignoring group with more than 1 host definition: %v", name)
				continue
			}
			vars := obj["vars"].(map[string]interface{})
			strVars := make(map[string]string, len(vars))
			for name, val := range vars {
				strVars[name] = val.(string)
			}
			node := m.getById(name)
			node.Name = hosts[0].(string)
			node.Vars = strVars
		}
	}
	for _, node := range m.Ids {
		if node.Name == "" {
			panic(fmt.Sprintf("DependencyNode with ID %v has no name: %v", node.Id, node))
		}
		m.Names[node.Name] = node
		m.Layers[node.Layer] = append(m.Layers[node.Layer], node)
	}
}

func (m *DependencyModel) getById(id string) *DependencyNode {
	node, ok := m.Ids[id]
	if !ok {
		node = &DependencyNode{
			Vars: make(map[string]string),
			Id:   id,
		}
		m.Ids[id] = node
	}
	return node
}

func (m *DependencyModel) resolveGroup(name string, children []interface{}, root map[string]interface{}) []*DependencyNode {
	if group, hasGroup := m.Groups[name]; hasGroup {
		return group
	}

	var res []*DependencyNode
	for _, child := range children {
		id := child.(string)
		if children, ok := root[id].(map[string]interface{})["children"]; ok {
			// Recursively points  to another group
			subGroup := m.resolveGroup(id, children.([]interface{}), root)
			res = append(res, subGroup...)
		} else {
			res = append(res, m.getById(id))
		}
	}
	m.Groups[name] = res
	for _, node := range res {
		node.Groups = append(node.Groups, name)
	}
	return res
}

func (n *DependencyNode) setLayer(layer Layer) {
	if n.Layer != UnknownLayer && n.Layer != layer {
		panic(fmt.Sprintf("DependencyNode already has layer %v, cannot change to %v: %v", n.Layer, layer, n))
	}
	n.Layer = layer
}

func (n *DependencyNode) addChild(child *DependencyNode) {
	if child.Parent != nil {
		panic(fmt.Sprintf("DependencyNode %v already has parent %v, cannot switch to %v", child, child.Parent, n))
	}
	n.Children = append(n.Children, child)
	child.Parent = n
}
