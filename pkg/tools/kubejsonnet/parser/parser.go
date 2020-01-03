// Package parser parses evaluated jsonnet files and outputs
// KubernetesResources in a Prodspec asset.
package parser

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/google/go-jsonnet"
)

var (
	vm = jsonnet.MakeVM()
)

// Metadata is parsed from the json files and is a ConfigMap that holds
// metadata about the cluster.
type Metadata map[string]string

func (m Metadata) valid() bool {
	return m["clusterName"] != ""
}

// Parse takes a path to a jsonnet or json file and parses out the
// Kubernetes unstructured.Unstructured objects.
func Parse(filename string) (Metadata, []*unstructured.Unstructured, error) {
	bytes, err := readFile(filename)
	if err != nil {
		return nil, nil, err
	}
	uList, err := parse(bytes)
	if err != nil {
		return nil, nil, err
	}

	m, err := parseMetamap(uList)
	if err != nil {
		return nil, nil, err
	}
	return m, uList, nil
}

func readFile(filename string) ([]byte, error) {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	switch {
	case strings.HasSuffix(filename, ".jsonnet"):
		str, err := vm.EvaluateSnippet(filename, string(bytes))
		if err != nil {
			return nil, err
		}
		bytes = []byte(str)
	case strings.HasSuffix(filename, ".json"):
	default:
		return nil, fmt.Errorf("unsupported filetype: %s", filename)
	}
	return bytes, err
}

func parseMetamap(uList []*unstructured.Unstructured) (Metadata, error) {
	var m Metadata
	for _, u := range uList {
		if isMetaMap(u) {
			m = parseMetaMap(u)
			break
		}
	}
	if m == nil || !m.valid() {
		return nil, fmt.Errorf("could not determine cluster metadata")
	}
	return m, nil
}

func parse(bytes []byte) ([]*unstructured.Unstructured, error) {
	var data interface{}
	if err := json.Unmarshal(bytes, &data); err != nil {
		return nil, err
	}

	objs, err := walk(data)
	if err != nil {
		return nil, err
	}

	uList := make([]*unstructured.Unstructured, 0, len(objs))
	for _, o := range objs {
		uList = append(uList, &unstructured.Unstructured{Object: o.(map[string]interface{})})
	}

	return uList, nil
}

func parseMetaMap(u *unstructured.Unstructured) Metadata {
	data := make(map[string]string)
	for k, v := range u.Object["data"].(map[string]interface{}) {
		data[k] = v.(string)
	}
	return data
}

func isMetaMap(u *unstructured.Unstructured) bool {
	return u.GetKind() == "ConfigMap" && u.GetName() == "cluster-metadata"
}

func isKubeObj(obj map[string]interface{}) bool {
	return obj["kind"] != nil && obj["apiVersion"] != nil
}

// Unstructured takes a list of interfaces
func walk(obj interface{}) ([]interface{}, error) {
	switch o := obj.(type) {
	case nil:
		return []interface{}{}, nil
	case []interface{}:
		ret := make([]interface{}, 0, len(o))
		for _, v := range o {
			children, err := walk(v)
			if err != nil {
				return nil, err
			}
			ret = append(ret, children...)
		}
		return ret, nil
	case map[string]interface{}:
		if isKubeObj(o) {
			return []interface{}{o}, nil
		}
		ret := []interface{}{}
		for _, v := range o {
			children, err := walk(v)
			if err != nil {
				return nil, err
			}
			ret = append(ret, children...)
		}
		return ret, nil
	default:
		return nil, fmt.Errorf("couldn't parse kubernetes object %+v", o)
	}
}
