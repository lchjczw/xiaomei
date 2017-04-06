package release

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
)

type Stack struct {
	Version  string
	Services map[string]Service `yaml:",omitempty"`
	Volumes  map[string]Volume  `yaml:",omitempty"`
	Networks map[string]Network `yaml:",omitempty"`
}
type Service map[string]interface{}
type Volume map[string]interface{}
type Network map[string]interface{}

var theStack *Stack

func GetStack() Stack {
	if theStack == nil {
		content, err := ioutil.ReadFile(filepath.Join(Root(), `stack.yml`))
		if err != nil {
			panic(err)
		}
		stack := &Stack{Services: make(map[string]Service)}
		if err := yaml.Unmarshal(content, stack); err != nil {
			panic(err)
		}
		theStack = stack
	}
	return *theStack
}

func GetService(svcName string) Service {
	service := GetStack().Services[svcName]
	if service == nil {
		panic(fmt.Sprintf(`stack.yml: services.%s: undefined.`, svcName))
	}
	return service
}

func ImageNameOf(svcName string) string {
	service := GetService(svcName)
	image := service[`image`]
	if image == nil {
		panic(fmt.Sprintf(`stack.yml: services.%s.image: undefined.`, svcName))
	}
	if str, ok := image.(string); ok && str != `` {
		return str
	} else {
		panic(fmt.Sprintf(`stack.yml: services.%s.image: should be a string.`, svcName))
	}
}

func PortsOf(svcName string) (ports []string) {
	service := GetService(svcName)
	iports, _ := service[`ports`].([]interface{})
	for i, iport := range iports {
		switch port := iport.(type) {
		case string:
			if index := strings.IndexByte(port, ':'); index < 0 {
				panic(fmt.Sprintf(
					`stack.yml: services.%s.ports: random host port is not allowed: %s.`, svcName, port,
				))
			}
			ports = append(ports, port)
		case map[interface{}]interface{}:
			ports = append(ports, fmt.Sprint(port[`published`])+`:`+fmt.Sprint(port[`target`]))
		default:
			panic(fmt.Sprintf(
				`stack.yml: services.%s.ports[%d]: should be a string, got: %s.`, svcName, i, iport,
			))
		}
	}
	return ports
}

func EachServiceDo(work func(svcName string) error) error {
	for svcName := range GetStack().Services {
		if svcName != `` {
			if err := work(svcName); err != nil {
				return err
			}
		}
	}
	return nil
}