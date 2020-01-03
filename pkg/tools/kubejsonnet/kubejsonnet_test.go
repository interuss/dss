package kubejsonnet

import (
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

func TestNew(t *testing.T) {
	tests := []struct {
		name        string
		filename    string
		expectError bool
	}{
		{
			name: "test examples/minimum.jsonnet converted to json file",

			filename: "testfiles/good.json",

			expectError: false,
		},
		{
			name: "test examples/minimum.jsonnet",

			filename: "../../../build/deploy/examples/minimum.jsonnet",

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
	for _, test := range tests {
		_, err := New(test.filename, []string{"", ""})
		if (err == nil) == test.expectError {
			t.Errorf("test %s: received unexpected error: \"%+v\"", test.name, err)
		}
	}
}
