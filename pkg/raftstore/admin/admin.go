package admin

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"sync"

	"github.com/interuss/dss/pkg/logging"
	"github.com/interuss/stacktrace"
	"go.etcd.io/raft/v3/raftpb"
	"go.uber.org/zap"
)

var adminAddr string

const raftAdminAddrFlag = "raft_admin_addr"

func init() {
	flag.StringVar(&adminAddr, raftAdminAddrFlag, "", "If set and the raft store backend is in use, determines the address and port for the raft admin endpoint (POST /admin/members), used to add/remove/update a node across all of the rid/scd/aux raft clusters in a single request. Disabled if empty.")
}

var (
	globalRegistry = &storeRegistry{stores: map[string]MembershipManager{}}
	startAdminOnce sync.Once
)

// AdminServer exposes an HTTP endpoint, that can add, update or remove a node
// from every raftstore.Store instance registered with it.
type AdminServer struct {
	logger     *zap.Logger
	registry   *storeRegistry
	httpServer *http.Server
}

// Register is called by each store that wants to be discoverable through the admin endpoint.
// The first time any store registers, the admin endpoint is started in the backgroundd.
func Register(ctx context.Context, logger *zap.Logger, name string, store MembershipManager) {
	globalRegistry.register(name, store)

	startAdminOnce.Do(func() {
		if adminAddr == "" {
			logger.Fatal(fmt.Sprintf("--%s is required", raftAdminAddrFlag))
		}

		adminServer := &AdminServer{
			logger:   logging.WithValuesFromContext(ctx, logger),
			registry: globalRegistry,
		}

		adminServer.httpServer = &http.Server{
			Addr:    adminAddr,
			Handler: http.HandlerFunc(adminServer.handleMemberChange),
		}

		go func() {
			if err := adminServer.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				logger.Error("raft admin server error", zap.Error(err))
			}
		}()
	})
}

type changeRequest struct {
	Action    raftpb.ConfChangeType `json:"action"`
	NodeID    uint64                `json:"node_id"`
	Addresses map[string]string     `json:"addresses,omitempty"` // required for add/update
}

type changeResponse struct {
	Results map[string]storeResult `json:"results"`
}
type storeResult struct {
	ConfState *raftpb.ConfState `json:"conf_state,omitempty"`
	Error     string            `json:"error,omitempty"`
}

func (a *AdminServer) handleMemberChange(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req changeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		a.writeJSON(w, http.StatusBadRequest, changeResponse{Results: map[string]storeResult{
			"": {Error: stacktrace.Propagate(err, "failed to decode request body").Error()},
		}})
		return
	}

	if req.NodeID == 0 {
		a.writeAdminError(w, http.StatusBadRequest, "node_id is required")
		return
	}

	if (req.Action == raftpb.ConfChangeAddNode || req.Action == raftpb.ConfChangeUpdateNode) && len(req.Addresses) != 3 {
		a.writeAdminError(w, http.StatusBadRequest, "scd, rid and aux addresses are required for action %q", req.Action)
		return
	}

	stores := a.registry.clone()
	var mu sync.Mutex
	results := make(map[string]storeResult, len(stores))

	var wg sync.WaitGroup
	for name, manager := range stores {
		wg.Go(func() {
			var address string
			if req.Action == raftpb.ConfChangeAddNode || req.Action == raftpb.ConfChangeUpdateNode {
				var ok bool
				address, ok = req.Addresses[name]
				if !ok {
					mu.Lock()
					results[name] = storeResult{Error: fmt.Sprintf("no address provided for store %q", name)}
					mu.Unlock()
					return
				}
			}

			confState, err := manager.ProposeConfChange(r.Context(), req.Action, req.NodeID, address)

			mu.Lock()
			if err != nil {
				results[name] = storeResult{Error: err.Error()}
			} else {
				results[name] = storeResult{ConfState: confState}
			}
			mu.Unlock()
		})
	}
	wg.Wait()

	a.logger.Info("processed membership change", zap.String("action", req.Action.String()), zap.Uint64("nodeID", req.NodeID), zap.Any("stores", stores))
	a.writeJSON(w, http.StatusOK, changeResponse{Results: results})
}

func (a *AdminServer) writeAdminError(w http.ResponseWriter, status int, format string, args ...any) {
	a.writeJSON(w, status, changeResponse{Results: map[string]storeResult{
		"": {Error: stacktrace.NewError(format, args...).Error()},
	}})
}

func (a *AdminServer) writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		a.logger.Error("failed to encode admin response", zap.Error(err))
	}
}
