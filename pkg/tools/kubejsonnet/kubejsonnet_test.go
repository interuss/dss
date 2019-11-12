package kubejsonnet

import (
	"fmt"
	"path/filepath"
	"runtime"
	"testing"
)

func TestIsMutatingCommand(t *testing.T) {
	tests := []struct {
		name    string
		command string
		want    bool
	}{
		{
			name:    "test apply",
			want:    true,
			command: "apply",
		},
		{
			name:    "test create",
			want:    true,
			command: "create",
		},
		{
			name:    "test replace",
			want:    true,
			command: "replace",
		},
		{
			name:    "test diff",
			want:    false,
			command: "diff",
		},
	}
	for _, test := range tests {
		cmd := Command{command: test.command}
		if got := cmd.IsMutatingCommand(); test.want != got {
			t.Errorf("expected %v, got %v for test: %s", test.want, got, test.name)
		}
	}
}

func TestParseClusterContextFromFile(t *testing.T) {
	tests := []struct {
		name        string
		filename    string
		expectError bool
	}{
		{
			name: "test examples/minimum.jsonnet",

			filename: "../../../build/deploy/examples/prod_leaf.jsonnet",

			expectError: false,
		},
		{
			name:     "test examples/prod_leaf.jsonnet",
			filename: "../../../build/deploy/examples/prod_leaf.jsonnet",

			expectError: false,
		},
		{
			name:        "test no metamap found",
			filename:    "testfiles/bad.jsonnet",
			expectError: true,
		},
	}
	_, relpath, _, _ := runtime.Caller(0)
	fmt.Println(relpath)
	fmt.Println(filepath.Dir(relpath))
	for _, test := range tests {
		cmd := Command{filename: test.filename, metaname: "cluster-metadata"}
		if err := cmd.parseClusterContextFromFile(); (err == nil) == test.expectError {
			t.Errorf("test %s: received unexpected error: \"%+v\"", test.name, err)
		}
	}
}
