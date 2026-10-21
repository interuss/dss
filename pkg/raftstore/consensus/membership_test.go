package consensus

import (
	"fmt"
	"testing"

	dsserr "github.com/interuss/dss/pkg/errors"
	"github.com/interuss/stacktrace"
	"github.com/stretchr/testify/require"
	"go.etcd.io/raft/v3"
	"go.etcd.io/raft/v3/raftpb"
)

func TestValidateConfChange(t *testing.T) {
	voters := []uint64{1, 2, 3, 4}
	removed := []uint64{5}
	add := func(id uint64) raftpb.ConfChangeSingle {
		return raftpb.ConfChangeSingle{Type: raftpb.ConfChangeAddNode, NodeID: id}
	}
	remove := func(id uint64) raftpb.ConfChangeSingle {
		return raftpb.ConfChangeSingle{Type: raftpb.ConfChangeRemoveNode, NodeID: id}
	}

	tests := []struct {
		name    string
		change  raftpb.ConfChangeSingle
		url     string
		voters  []uint64
		wantErr bool
	}{
		{name: "add new node", change: add(6), url: "http://node6:9011", voters: voters},
		{name: "add invalid ID", change: add(raft.None), url: "http://node6:9011", voters: voters, wantErr: true},
		{name: "add removed ID", change: add(5), url: "http://node5:9011", voters: voters, wantErr: true},
		{name: "add existing member", change: add(2), url: "http://node2:9011", voters: voters, wantErr: true},
		{name: "add without URL", change: add(6), voters: voters, wantErr: true},
		{name: "add with URL without port", change: add(6), url: "http://node6", voters: voters, wantErr: true},
		{name: "remove member", change: remove(4), voters: voters},
		{name: "remove removed ID", change: remove(5), voters: voters, wantErr: true},
		{name: "remove non member", change: remove(6), voters: voters, wantErr: true},
		{name: "remove below minimum voters", change: remove(3), voters: []uint64{1, 2, 3}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			members := map[uint64]string{}
			for _, id := range tt.voters {
				members[id] = fmt.Sprintf("http://node%d:9011", id)
			}
			m := &membership{members: members, removedIDs: removed}
			err := m.validateConfChange(tt.change, tt.url)
			if !tt.wantErr {
				require.NoError(t, err)
				return
			}
			require.Equal(t, dsserr.BadRequest, stacktrace.GetCode(err))
		})
	}
}
