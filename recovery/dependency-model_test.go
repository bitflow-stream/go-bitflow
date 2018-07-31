package recovery

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type dependencyModelTestSuite struct {
	t *testing.T
	*require.Assertions
}

func TestDependencyModel(t *testing.T) {
	suite.Run(t, new(dependencyModelTestSuite))
}

func (suite *dependencyModelTestSuite) T() *testing.T {
	return suite.t
}

func (suite *dependencyModelTestSuite) SetT(t *testing.T) {
	suite.t = t
	suite.Assertions = require.New(t)
}

func (suite *dependencyModelTestSuite) TestModel() {
	model, err := ParseDependencyModel([]byte(jsonTestData))
	suite.NoError(err)
	suite.NotNil(model)

	suite.Len(model.Layers[ServiceLayer], 10)
	suite.Len(model.Layers[VMLayer], 10)
	suite.Len(model.Layers[HypervisorLayer], 6)
	suite.Len(model.Layers[UnknownLayer], 5)
	suite.Len(model.Layers, 4)
	suite.Len(model.Ids, 31)
	suite.Len(model.Names, 31)
	suite.Len(model.Groups, 10)

	// Check consistency of nodes and index-maps
	for id, node := range model.Ids {
		suite.Equal(id, node.Id)
	}
	for name, node := range model.Names {
		suite.Equal(name, node.Name)
	}
	for layer, nodes := range model.Layers {
		for _, node := range nodes {
			suite.Equal(layer, node.Layer)
		}
	}
	for _, node := range model.Ids {
		suite.Contains(model.Layers[node.Layer], node)
	}
	for name, nodes := range model.Groups {
		for _, node := range nodes {
			suite.Contains(node.Groups, name)
		}
	}
	for _, node := range model.Ids {
		for _, group := range node.Groups {
			suite.Contains(model.Groups[group], node)
		}
	}

	// Check parents/children
	for _, hv := range model.Layers[HypervisorLayer] {
		suite.Nil(hv.Parent)
		suite.NotEmpty(hv.Children)
	}
	for _, vm := range model.Layers[VMLayer] {
		suite.NotNil(vm.Parent)
		suite.NotEmpty(vm.Children)
	}
	for _, service := range model.Layers[ServiceLayer] {
		suite.NotNil(service.Parent)
		suite.Empty(service.Children)
	}
	for _, node := range model.Layers[UnknownLayer] {
		suite.Nil(node.Parent)
		suite.Empty(node.Children)
	}
}

var jsonTestData = `
{
  "bico": {
    "hosts": [
      "bico"
    ],
    "vars": {
      "ansible_host": "192.168.4.164"
    }
  },
  "wally141": {
    "hosts": [
      "wally141"
    ],
    "vars": {
      "iface": "em1",
      "hostname": "wally141",
      "ansible_host": "wally141.cit.tu-berlin.de",
      "injector-endpoint": "wally141.cit.tu-berlin.de:7888"
    }
  },
  "ea95a6d2-eaf9-4126-8f24-17c21b1b4856": {
    "hosts": [
      "bono-1.ims4"
    ],
    "vars": {
      "iface": "tapc833efbf-3a",
      "vnf_service_regex": "bono",
      "hypervisor": "wally136.cit.tu-berlin.de",
      "hostname": "bono-1.ims4",
      "libvirt": "instance-00000b51",
      "vnf_service": "bono",
      "mac_addr": "fa:16:3e:6f:e4:9e",
      "injector-endpoint": "192.168.4.213:7888",
      "port_id": "c833efbf-3a89-446c-aa7a-b2f19a9b99f6",
      "ansible_host": "192.168.4.213",
      "private_ip": "10.0.0.8"
    }
  },
  "sprout": {
    "children": [
      "c0a2bb46-a8b5-408e-a9d7-2aeaaaf91fce",
      "a9a738a8-e40e-44ce-85bb-babc7f892fb2",
      "1a32608c-92e5-41cd-981d-091d83a4f512"
    ]
  },
  "sprout@sprout-1.ims4": {
    "hosts": [
      "sprout@sprout-1.ims4"
    ],
    "vars": {
      "injector-endpoint": "192.168.4.217:7889"
    }
  },
  "7989ead5-14c1-46d8-a94a-0e6b6fe078ec": {
    "hosts": [
      "ellis-0.ims4"
    ],
    "vars": {
      "iface": "tapa655cf85-64",
      "vnf_service_regex": "ellis",
      "hypervisor": "wally147.cit.tu-berlin.de",
      "hostname": "ellis-0.ims4",
      "libvirt": "instance-00000b4b",
      "vnf_service": "ellis",
      "mac_addr": "fa:16:3e:aa:0d:c8",
      "injector-endpoint": "192.168.4.211:7888",
      "port_id": "a655cf85-6424-4e0b-a381-26025b665d63",
      "ansible_host": "192.168.4.211",
      "private_ip": "10.0.0.6"
    }
  },
  "hypervisors": {
    "children": [
      "wally142",
      "wally141",
      "wally147",
      "wally131",
      "wally136",
      "wally134"
    ]
  },
  "79679734-68ac-417d-b7fa-6ff8d17efd7a": {
    "hosts": [
      "ns.ims4"
    ],
    "vars": {
      "iface": "tapacf61926-e3",
      "vnf_service_regex": "named",
      "hypervisor": "wally134.cit.tu-berlin.de",
      "hostname": "ns.ims4",
      "libvirt": "instance-00000b45",
      "vnf_service": "ns",
      "mac_addr": "fa:16:3e:48:0c:9c",
      "injector-endpoint": "192.168.4.209:7888",
      "port_id": "acf61926-e368-48b6-8d88-ddd9c99d445b",
      "ansible_host": "192.168.4.209",
      "private_ip": "10.0.0.4"
    }
  },
  "ellis@ellis-0.ims4": {
    "hosts": [
      "ellis@ellis-0.ims4"
    ],
    "vars": {
      "injector-endpoint": "192.168.4.211:7889"
    }
  },
  "79679734-68ac-417d-b7fa-6ff8d17efd7a_services": {
    "children": [
      "ns@ns.ims4"
    ]
  },
  "sprout@sprout-2.ims4": {
    "hosts": [
      "sprout@sprout-2.ims4"
    ],
    "vars": {
      "injector-endpoint": "192.168.4.219:7889"
    }
  },
  "f58b7393-6e65-4937-bdc2-28548060559e": {
    "hosts": [
      "bono-2.ims4"
    ],
    "vars": {
      "iface": "tap843bce58-ae",
      "vnf_service_regex": "bono",
      "hypervisor": "wally141.cit.tu-berlin.de",
      "hostname": "bono-2.ims4",
      "libvirt": "instance-00000b5a",
      "vnf_service": "bono",
      "mac_addr": "fa:16:3e:85:55:a9",
      "injector-endpoint": "192.168.4.216:7888",
      "port_id": "843bce58-ae32-48ac-b500-8b84dc02425c",
      "ansible_host": "192.168.4.216",
      "private_ip": "10.0.0.11"
    }
  },
  "recovery_engine": {
    "children": [
      "recovery_engine_host"
    ]
  },
  "bono@bono-1.ims4": {
    "hosts": [
      "bono@bono-1.ims4"
    ],
    "vars": {
      "injector-endpoint": "192.168.4.213:7889"
    }
  },
  "ellis": {
    "children": [
      "7989ead5-14c1-46d8-a94a-0e6b6fe078ec"
    ]
  },
  "c4373341-a3dc-4c94-9162-ac3c58b09cb2_services": {
    "children": [
      "bono@bono-0.ims4"
    ]
  },
  "wally142": {
    "hosts": [
      "wally142"
    ],
    "vars": {
      "iface": "em1",
      "hostname": "wally142",
      "ansible_host": "wally142.cit.tu-berlin.de",
      "injector-endpoint": "wally142.cit.tu-berlin.de:7888"
    }
  },
  "recovery_engine_host": {
    "hosts": [
      "recovery_engine_host"
    ],
    "vars": {
      "ansible_host": "192.168.4.178"
    }
  },
  "wally131_vms": {
    "children": [
      "c0a2bb46-a8b5-408e-a9d7-2aeaaaf91fce"
    ]
  },
  "wally147": {
    "hosts": [
      "wally147"
    ],
    "vars": {
      "iface": "em1",
      "hostname": "wally147",
      "ansible_host": "wally147.cit.tu-berlin.de",
      "injector-endpoint": "wally147.cit.tu-berlin.de:7888"
    }
  },
  "homer@homer-0.ims4": {
    "hosts": [
      "homer@homer-0.ims4"
    ],
    "vars": {
      "injector-endpoint": "192.168.4.210:7889"
    }
  },
  "c0a2bb46-a8b5-408e-a9d7-2aeaaaf91fce_services": {
    "children": [
      "sprout@sprout-2.ims4"
    ]
  },
  "clearwater": {
    "children": [
      "bono",
      "homestead",
      "homer",
      "sprout",
      "ellis",
      "ns"
    ]
  },
  "a9a738a8-e40e-44ce-85bb-babc7f892fb2_services": {
    "children": [
      "sprout@sprout-0.ims4"
    ]
  },
  "1a32608c-92e5-41cd-981d-091d83a4f512_services": {
    "children": [
      "sprout@sprout-1.ims4"
    ]
  },
  "ns": {
    "children": [
      "79679734-68ac-417d-b7fa-6ff8d17efd7a"
    ]
  },
  "rca": {
    "hosts": [
      "rca"
    ],
    "vars": {
      "ansible_host": "192.168.4.171"
    }
  },
  "1a32608c-92e5-41cd-981d-091d83a4f512": {
    "hosts": [
      "sprout-1.ims4"
    ],
    "vars": {
      "iface": "tap42988f39-50",
      "vnf_service_regex": "sprout",
      "hypervisor": "wally147.cit.tu-berlin.de",
      "hostname": "sprout-1.ims4",
      "libvirt": "instance-00000b5d",
      "vnf_service": "sprout",
      "mac_addr": "fa:16:3e:47:b8:29",
      "injector-endpoint": "192.168.4.217:7888",
      "port_id": "42988f39-50de-4d50-919d-f5a6f36b5284",
      "ansible_host": "192.168.4.217",
      "private_ip": "10.0.0.12"
    }
  },
  "homer": {
    "children": [
      "62f570a5-e138-4e06-b18e-aaf56e05a9fc"
    ]
  },
  "anomaly_detector": {
    "children": [
      "svm",
      "bico",
      "rca",
      "bico_alex"
    ]
  },
  "ea95a6d2-eaf9-4126-8f24-17c21b1b4856_services": {
    "children": [
      "bono@bono-1.ims4"
    ]
  },
  "bono@bono-2.ims4": {
    "hosts": [
      "bono@bono-2.ims4"
    ],
    "vars": {
      "injector-endpoint": "192.168.4.216:7889"
    }
  },
  "wally141_vms": {
    "children": [
      "f58b7393-6e65-4937-bdc2-28548060559e",
      "2fe3dbdf-7a92-4cfa-87b4-2951283e24d9"
    ]
  },
  "wally147_vms": {
    "children": [
      "1a32608c-92e5-41cd-981d-091d83a4f512",
      "7989ead5-14c1-46d8-a94a-0e6b6fe078ec"
    ]
  },
  "2fe3dbdf-7a92-4cfa-87b4-2951283e24d9_services": {
    "children": [
      "homestead@homestead-0.ims4"
    ]
  },
  "7989ead5-14c1-46d8-a94a-0e6b6fe078ec_services": {
    "children": [
      "ellis@ellis-0.ims4"
    ]
  },
  "c4373341-a3dc-4c94-9162-ac3c58b09cb2": {
    "hosts": [
      "bono-0.ims4"
    ],
    "vars": {
      "iface": "tap0792ebd3-5e",
      "vnf_service_regex": "bono",
      "hypervisor": "wally142.cit.tu-berlin.de",
      "hostname": "bono-0.ims4",
      "libvirt": "instance-00000b54",
      "vnf_service": "bono",
      "mac_addr": "fa:16:3e:89:21:08",
      "injector-endpoint": "192.168.4.214:7888",
      "port_id": "0792ebd3-5ebf-4cb7-ad25-d85af8d0cd3e",
      "ansible_host": "192.168.4.214",
      "private_ip": "10.0.0.9"
    }
  },
  "wally136_vms": {
    "children": [
      "a9a738a8-e40e-44ce-85bb-babc7f892fb2",
      "ea95a6d2-eaf9-4126-8f24-17c21b1b4856"
    ]
  },
  "homestead@homestead-0.ims4": {
    "hosts": [
      "homestead@homestead-0.ims4"
    ],
    "vars": {
      "injector-endpoint": "192.168.4.215:7889"
    }
  },
  "2fe3dbdf-7a92-4cfa-87b4-2951283e24d9": {
    "hosts": [
      "homestead-0.ims4"
    ],
    "vars": {
      "iface": "tapb53a7974-0d",
      "vnf_service_regex": "homestead",
      "hypervisor": "wally141.cit.tu-berlin.de",
      "hostname": "homestead-0.ims4",
      "libvirt": "instance-00000b57",
      "vnf_service": "homestead",
      "mac_addr": "fa:16:3e:3d:97:27",
      "injector-endpoint": "192.168.4.215:7888",
      "port_id": "b53a7974-0dbe-4cca-972c-3fbc3b734e80",
      "ansible_host": "192.168.4.215",
      "private_ip": "10.0.0.10"
    }
  },
  "bico_alex": {
    "hosts": [
      "bico_alex"
    ],
    "vars": {
      "ansible_host": "192.168.4.175"
    }
  },
  "wally142_vms": {
    "children": [
      "c4373341-a3dc-4c94-9162-ac3c58b09cb2"
    ]
  },
  "62f570a5-e138-4e06-b18e-aaf56e05a9fc": {
    "hosts": [
      "homer-0.ims4"
    ],
    "vars": {
      "iface": "tapd24961e3-bb",
      "vnf_service_regex": "homer",
      "hypervisor": "wally134.cit.tu-berlin.de",
      "hostname": "homer-0.ims4",
      "libvirt": "instance-00000b48",
      "vnf_service": "homer",
      "mac_addr": "fa:16:3e:6d:85:44",
      "injector-endpoint": "192.168.4.210:7888",
      "port_id": "d24961e3-bb63-4c16-92b5-d4f0bfa661a4",
      "ansible_host": "192.168.4.210",
      "private_ip": "10.0.0.5"
    }
  },
  "wally131": {
    "hosts": [
      "wally131"
    ],
    "vars": {
      "iface": "em1",
      "hostname": "wally131",
      "ansible_host": "wally131.cit.tu-berlin.de",
      "injector-endpoint": "wally131.cit.tu-berlin.de:7888"
    }
  },
  "homestead": {
    "children": [
      "2fe3dbdf-7a92-4cfa-87b4-2951283e24d9"
    ]
  },
  "bono@bono-0.ims4": {
    "hosts": [
      "bono@bono-0.ims4"
    ],
    "vars": {
      "injector-endpoint": "192.168.4.214:7889"
    }
  },
  "wally134_vms": {
    "children": [
      "62f570a5-e138-4e06-b18e-aaf56e05a9fc",
      "79679734-68ac-417d-b7fa-6ff8d17efd7a"
    ]
  },
  "sprout@sprout-0.ims4": {
    "hosts": [
      "sprout@sprout-0.ims4"
    ],
    "vars": {
      "injector-endpoint": "192.168.4.218:7889"
    }
  },
  "svm": {
    "hosts": [
      "svm"
    ],
    "vars": {
      "ansible_host": "192.168.4.172"
    }
  },
  "wally": {
    "hosts": [
      "wally137.cit.tu-berlin.de",
      "wally133.cit.tu-berlin.de",
      "wally139.cit.tu-berlin.de",
      "wally132.cit.tu-berlin.de",
      "wally138.cit.tu-berlin.de",
      "wally135.cit.tu-berlin.de",
      "wally140.cit.tu-berlin.de",
      "wally145.cit.tu-berlin.de",
      "wally131.cit.tu-berlin.de",
      "wally147.cit.tu-berlin.de",
      "wally134.cit.tu-berlin.de",
      "wally146.cit.tu-berlin.de",
      "wally141.cit.tu-berlin.de",
      "wally136.cit.tu-berlin.de",
      "wally142.cit.tu-berlin.de",
      "wally148.cit.tu-berlin.de",
      "wally143.cit.tu-berlin.de",
      "wally149.cit.tu-berlin.de",
      "wally144.cit.tu-berlin.de"
    ],
    "vars": {
      "ansible_ssh_private_key_file": "~/.ssh/id_wally"
    }
  },
  "a9a738a8-e40e-44ce-85bb-babc7f892fb2": {
    "hosts": [
      "sprout-0.ims4"
    ],
    "vars": {
      "iface": "tap2877fa0f-4b",
      "vnf_service_regex": "sprout",
      "hypervisor": "wally136.cit.tu-berlin.de",
      "hostname": "sprout-0.ims4",
      "libvirt": "instance-00000b60",
      "vnf_service": "sprout",
      "mac_addr": "fa:16:3e:6e:35:f2",
      "injector-endpoint": "192.168.4.218:7888",
      "port_id": "2877fa0f-4bca-4f34-a803-c77039bbfbb0",
      "ansible_host": "192.168.4.218",
      "private_ip": "10.0.0.13"
    }
  },
  "c0a2bb46-a8b5-408e-a9d7-2aeaaaf91fce": {
    "hosts": [
      "sprout-2.ims4"
    ],
    "vars": {
      "iface": "tap57a3765a-c0",
      "vnf_service_regex": "sprout",
      "hypervisor": "wally131.cit.tu-berlin.de",
      "hostname": "sprout-2.ims4",
      "libvirt": "instance-00000b63",
      "vnf_service": "sprout",
      "mac_addr": "fa:16:3e:fc:8b:d4",
      "injector-endpoint": "192.168.4.219:7888",
      "port_id": "57a3765a-c0e5-488d-9418-8ccea6406ff2",
      "ansible_host": "192.168.4.219",
      "private_ip": "10.0.0.14"
    }
  },
  "ns@ns.ims4": {
    "hosts": [
      "ns@ns.ims4"
    ],
    "vars": {
      "injector-endpoint": "192.168.4.209:7889"
    }
  },
  "bono": {
    "children": [
      "f58b7393-6e65-4937-bdc2-28548060559e",
      "c4373341-a3dc-4c94-9162-ac3c58b09cb2",
      "ea95a6d2-eaf9-4126-8f24-17c21b1b4856"
    ]
  },
  "f58b7393-6e65-4937-bdc2-28548060559e_services": {
    "children": [
      "bono@bono-2.ims4"
    ]
  },
  "62f570a5-e138-4e06-b18e-aaf56e05a9fc_services": {
    "children": [
      "homer@homer-0.ims4"
    ]
  },
  "wally136": {
    "hosts": [
      "wally136"
    ],
    "vars": {
      "iface": "em1",
      "hostname": "wally136",
      "ansible_host": "wally136.cit.tu-berlin.de",
      "injector-endpoint": "wally136.cit.tu-berlin.de:7888"
    }
  },
  "wally134": {
    "hosts": [
      "wally134"
    ],
    "vars": {
      "iface": "em1",
      "hostname": "wally134",
      "ansible_host": "wally134.cit.tu-berlin.de",
      "injector-endpoint": "wally134.cit.tu-berlin.de:7888"
    }
  }
}
`
