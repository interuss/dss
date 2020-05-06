// Our OpenAPI interface -> openapi2proto -> grpc-gateway -> grpc toolchain
// does not currently deal with enums that use "string" as the underlying type.
// This tool rewrites an OpenAPI yaml file to change the type of string-based
// enums to just plain strings for a wire-compatible data format that the
// toolchain can handle properly.  Enum value checking must then be performed
// manually in each handler.

package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

// Whenever node or a descendant contains a list in its "enum" key and "string"
// in its "type" key, replace the "enum" key with an "example" key containing
// the first value in the original "enum" value list
func fixStringEnums(node *map[interface{}]interface{}) error {
	for _, v := range *node {
		m, ok := v.(map[interface{}]interface{})
		if ok {
			enumListNode, hasEnum := m["enum"]
			if hasEnum {
				enumList, enumNodeIsList := enumListNode.([]interface{})
				if enumNodeIsList && m["type"] == "string" {
					delete(m, "enum")
					m["example"] = enumList[0]
				}
			}
			fixStringEnums(&m)
		}
	}
	return nil
}

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage:")
		fmt.Println("  go get gopkg.in/yaml.v2")
		fmt.Println("  go run adjust_openapi_yaml.go -- INPUT_YAML_FILENAME OUTPUT_YAML_FILENAME")
	}

	inputFilename := os.Args[1]
	outputFilename := os.Args[2]

	data, err := ioutil.ReadFile(inputFilename)
	check(err)

	m := make(map[interface{}]interface{})

	err = yaml.Unmarshal([]byte(data), &m)
	check(err)

	err = fixStringEnums(&m)
	check(err)

	data, err = yaml.Marshal(&m)
	check(err)

	err = ioutil.WriteFile(outputFilename, data, 0644)
	check(err)
}
