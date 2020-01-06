package parser

import (
	"encoding/json"
	"testing"
)

var badJSON = `
[
	{
		"Kind": "ConfigMap",
		"apiVersion": "v1"
	},
	{
		"interal_": {
			"bad_field": "can't have a string here"
		}
	}
]
`

func TestWalk(t *testing.T) {
	var data interface{}
	if err := json.Unmarshal([]byte(badJSON), &data); err != nil {
		t.Fatal(err)
	}
	if _, err := walk(data); err == nil {
		t.Errorf("expected an error from parsing bad_field")
	}
}
