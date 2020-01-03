package kubejsonnet

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/interuss/dss/pkg/tools/kubejsonnet/parser"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	kubectl "k8s.io/kubernetes/pkg/kubectl/cmd"
)

type Command struct {
	filename string
	command  string
	args     []string
	cluster  string
	tempfile string
	data     string
}

func checkArgs(args []string) error {
	// Don't supply file via -f, don't supply context
	probhibited := []string{"-f", "--context"}
	for _, arg := range args {
		for _, p := range probhibited {
			if arg == p {
				return fmt.Errorf("cannot provide arg: %s", arg)
			}
		}
	}
	return nil
}

func New(filename string, args []string) (*Command, error) {
	var data []byte
	var err error

	if err := checkArgs(args); err != nil {
		return nil, err
	}

	// At this point data is an unstructured s
	metadata, uList, err := parser.Parse(filename)
	if err != nil {
		return nil, err
	}

	for _, obj := range uList {
		objData, err := json.MarshalIndent(obj, "", " ")
		if err != nil {
			return nil, err
		}
		data = append(data, []byte("---\n")...)
		data = append(data, objData...)
		data = append(data, '\n')
	}

	tempfile, err := storeJson(data)
	if err != nil {
		return nil, err
	}

	/// todo store the output so we can call apply
	cmd := &Command{
		command:  args[0],
		args:     args[1:],
		filename: filename,
		cluster:  metadata["clusterName"],
		data:     string(data),
		tempfile: tempfile,
	}
	return cmd, err
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
	out := new(bytes.Buffer)
	cmd := kubectl.NewKubectlCommand(os.Stdin, out, os.Stderr)
	cmd.SetArgs([]string{"diff", "-f", c.tempfile, "--context", c.cluster})
	// The exit error will have a status code of 1 when there is a diff and when
	// there is an actual error :/
	// https://github.com/kubernetes/kubectl/issues/765
	cmdutil.BehaviorOnFatal(func(s string, code int) {
		if code > 1 {
			c.Cleanup()
			os.Exit(code)
		}
	})
	defer cmdutil.DefaultBehaviorOnFatal()
	err := cmd.Execute()
	return out.String(), err
}

func (c *Command) Run(ctx context.Context) (string, error) {
	// Run diff separately which requires special casing for the bad exit code
	// handling
	if c.command == "diff" {
		return c.Diff()
	}
	out, errOut := new(bytes.Buffer), new(bytes.Buffer)
	cmd := kubectl.NewKubectlCommand(os.Stdin, out, errOut)
	// Ignore errors on diff, which returns a bad exit code.
	args := append([]string{c.command, "-f", c.tempfile, "--context", c.cluster},
		c.args...)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), err
}

func (c *Command) Cleanup() error {
	return os.Remove(c.tempfile)
}

func storeJson(buf []byte) (string, error) {
	file, err := ioutil.TempFile(os.TempDir(), "parsed_json-*.json")
	if err != nil {
		return "", err
	}
	if _, err := file.Write(buf); err != nil {
		return "", err
	}
	return file.Name(), nil
}
