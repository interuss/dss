package kubejsonnet

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type Command struct {
	metaname string
	filename string
	command  string
	cluster  string
	tempFile string
}

func New(command, filename, metaname string) (*Command, error) {
	cmd := &Command{
		command:  command,
		filename: filename,
		metaname: metaname,
	}
	return cmd, cmd.parseClusterContextFromFile()
}

func (c *Command) IsMutatingCommand() bool {
	mutatingCommands := map[string]bool{
		"apply":   true,
		"create":  true,
		"replace": true,
	}
	_, exists := mutatingCommands[c.command]
	return exists
}

func (c *Command) Diff() (string, error) {
	out, err := exec.Command("kubectl", "diff", "-f", c.tempFile, "--context", c.cluster).Output()
	// The exit error will have a status code of 1 when there is a diff and when
	// there is an actual error :/
	// https://github.com/kubernetes/kubectl/issues/765
	if exitError, ok := err.(*exec.ExitError); ok {
		errorStr := string(exitError.Stderr)
		if strings.HasPrefix(errorStr, "error: ") {
			return "", errors.New(errorStr)
		}
	}
	return "will apply the following diff:\n\n" + string(out), nil
}

func (c *Command) Run(ctx context.Context) (string, error) {
	out, err := exec.CommandContext(ctx, "kubectl", "diff", "--context", c.cluster, "-f", c.tempFile).Output()
	if exitError, ok := err.(*exec.ExitError); ok {
		err = errors.New(string(exitError.Stderr))
	}
	return string(out), err
}

func (c *Command) Cleanup() error {
	return os.Remove(c.tempFile)
}

func (c *Command) storeYaml(buf []byte) (string, error) {
	file, err := ioutil.TempFile(os.TempDir(), filepath.Base(c.filename)+"-*.yaml")
	if err != nil {
		return "", err
	}
	if _, err := file.Write(buf); err != nil {
		return "", err
	}
	return file.Name(), nil
}

func (c *Command) parseClusterContextFromFile() error {
	b, err := getYaml(c.filename)
	if err != nil {
		return fmt.Errorf("error parsing jsonnet file with kubecfg show: %v", err)
	}
	c.tempFile, err = c.storeYaml(b)
	if err != nil {
		return err
	}
	decoder := yaml.NewDecoder(bytes.NewBuffer(b))
	var o map[string]interface{}
	for {
		err := decoder.Decode(&o)
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error parsing yaml %v", err)
		}
		u := &unstructured.Unstructured{Object: o}
		if u.IsList() {
			return fmt.Errorf("unsupported UnstructuredList for %s/%s/%s",
				u.GetKind(), u.GetNamespace(), u.GetName())
		}
		if isMetaMap(u, c.metaname) {
			c.parseMetaMap(u)
		}
	}
	if c.cluster == "" {
		return errors.New("could not find metamap with cluster name")
	}
	return nil
}

func (c *Command) parseMetaMap(u *unstructured.Unstructured) {
	data := u.Object["data"].(map[interface{}]interface{})
	c.cluster = data["clusterName"].(string)
}

func isMetaMap(u *unstructured.Unstructured, metaname string) bool {
	metadata := u.Object["metadata"].(map[interface{}]interface{})
	// TODO(steeling) why is u.GetName() and u.GetNamespace not working??
	return u.GetKind() == "ConfigMap" && metadata["name"] == metaname
}

func getYaml(filename string) ([]byte, error) {
	out, err := exec.Command("kubecfg", "show", filename).Output()
	if exitError, ok := err.(*exec.ExitError); ok {
		err = errors.New(string(exitError.Stderr))
	}
	return out, err
}
