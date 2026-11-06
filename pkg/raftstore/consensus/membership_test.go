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
	removed := map[uint64]member{5: {URL: "http://node5:9011"}}
	add := func(id uint64) *raftpb.ConfChangeSingle {
		return &raftpb.ConfChangeSingle{Type: raftpb.ConfChangeAddNode.Enum(), NodeId: new(id)}
	}
	remove := func(id uint64) *raftpb.ConfChangeSingle {
		return &raftpb.ConfChangeSingle{Type: raftpb.ConfChangeRemoveNode.Enum(), NodeId: new(id)}
	}

	node6 := member{URL: "http://node6:9011"}

	tests := []struct {
		name     string
		change   *raftpb.ConfChangeSingle
		node     member
		voters   []uint64
		wantErr  bool
		wantCode stacktrace.ErrorCode
	}{
		{name: "add new node", change: add(6), node: node6, voters: voters},
		{name: "add invalid ID", change: add(raft.None), node: node6, voters: voters, wantErr: true},
		{name: "add removed ID", change: add(5), node: member{URL: "http://node5:9011"}, voters: voters, wantErr: true},
		{name: "add existing member", change: add(2), node: member{URL: "http://node2:9011"}, voters: voters, wantErr: true},
		{name: "add without URL", change: add(6), node: member{}, voters: voters, wantErr: true},
		{name: "add with URL without port", change: add(6), node: member{URL: "http://node6"}, voters: voters, wantErr: true},
		{name: "add with the URL of a member", change: add(6), node: member{URL: "http://node2:9011"}, voters: voters, wantErr: true, wantCode: dsserr.AlreadyExists},
		{name: "add with the URL of a removed node", change: add(6), node: member{URL: "http://node5:9011"}, voters: voters},
		{name: "remove member", change: remove(4), voters: voters},
		{name: "remove removed ID", change: remove(5), voters: voters, wantErr: true},
		{name: "remove non member", change: remove(6), voters: voters, wantErr: true},
		{name: "remove below minimum voters", change: remove(3), voters: []uint64{1, 2, 3}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			members := map[uint64]member{}
			for _, id := range tt.voters {
				members[id] = member{URL: fmt.Sprintf("http://node%d:9011", id)}
			}
			m := &membership{members: members, removed: removed}
			err := m.validateConfChange(tt.change, tt.node)
			if !tt.wantErr {
				require.NoError(t, err)
				return
			}
			wantCode := dsserr.BadRequest
			if tt.wantCode != 0 {
				wantCode = tt.wantCode
			}
			require.Equal(t, wantCode, stacktrace.GetCode(err))
		})
	}
}
