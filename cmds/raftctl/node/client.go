// Package node implements the `raftctl node` subcommands, which change membership of a DSS
// instance's raftstore clusters by talking to its admin endpoint (see pkg/raftstore/admin).
package node

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"

	"go.etcd.io/raft/v3/raftpb"
)

const membersPath = "/admin/members"

type memberChangeRequest struct {
	Action    raftpb.ConfChangeType `json:"action"`
	NodeID    uint64                `json:"node_id"`
	Addresses map[string]string     `json:"addresses,omitempty"`
}

type storeResult struct {
	ConfState any    `json:"conf_state,omitempty"`
	Error     string `json:"error,omitempty"`
}

type memberChangeResponse struct {
	Results map[string]storeResult `json:"results"`
}

func changeMembership(addr string, action raftpb.ConfChangeType, nodeID uint64, addresses map[string]string) error {
	body, err := json.Marshal(memberChangeRequest{Action: action, NodeID: nodeID, Addresses: addresses})
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := http.Post(addr+membersPath, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to send request to %s: %w", addr, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	var result memberChangeResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return fmt.Errorf("failed to decode response (status %s): %s", resp.Status, respBody)
	}

	names := make([]string, 0, len(result.Results))
	for name := range result.Results {
		names = append(names, name)
	}
	sort.Strings(names)

	failed := false
	for _, name := range names {
		r := result.Results[name]
		if r.Error != "" {
			failed = true
			fmt.Printf("%s: error: %s\n", name, r.Error)
			continue
		}
		fmt.Printf("%s: ok, conf state: %v\n", name, r.ConfState)
	}

	if failed {
		return fmt.Errorf("one or more stores failed to apply the membership change")
	}

	return nil
}
