package raftstore

import (
	"context"
	"encoding/json"
	"slices"

	"github.com/golang/geo/s2"
	"github.com/interuss/dss/pkg/memstore"
	dssmodels "github.com/interuss/dss/pkg/models"
	"github.com/interuss/dss/pkg/raftstore"
	raftstoreadmin "github.com/interuss/dss/pkg/raftstore/admin"
	"github.com/interuss/dss/pkg/raftstore/consensus"
	ridmodels "github.com/interuss/dss/pkg/rid/models"
	"github.com/interuss/dss/pkg/rid/repos"
	ridmemstore "github.com/interuss/dss/pkg/rid/store/memstore"
	ridraftparams "github.com/interuss/dss/pkg/rid/store/raftstore/params"
	"github.com/interuss/stacktrace"
	"go.uber.org/zap"
)

const storeID = "rid"

const (
	getISA          raftstore.RequestType = "getISA"
	deleteISA       raftstore.RequestType = "deleteISA"
	insertISA       raftstore.RequestType = "insertISA"
	updateISA       raftstore.RequestType = "updateISA"
	searchISAs      raftstore.RequestType = "searchISAs"
	listExpiredISAs raftstore.RequestType = "listExpiredISAs"
	countISAs       raftstore.RequestType = "countISAs"

	DeleteISATransaction raftstore.RequestType = "deleteISATransaction"
	InsertISATransaction raftstore.RequestType = "insertISATransaction"
	UpdateISATransaction raftstore.RequestType = "updateISATransaction"

	getSubscription                    raftstore.RequestType = "getSubscription"
	deleteSubscription                 raftstore.RequestType = "deleteSubscription"
	insertSubscription                 raftstore.RequestType = "insertSubscription"
	updateSubscription                 raftstore.RequestType = "updateSubscription"
	searchSubscriptions                raftstore.RequestType = "searchSubscriptions"
	searchSubscriptionsByOwner         raftstore.RequestType = "searchSubscriptionsByOwner"
	updateNotificationIdxsInCells      raftstore.RequestType = "updateNotificationIdxsInCells"
	maxSubscriptionCountInCellsByOwner raftstore.RequestType = "maxSubscriptionCountInCellsByOwner"
	listExpiredSubscriptions           raftstore.RequestType = "listExpiredSubscriptions"
	countSubscriptions                 raftstore.RequestType = "countSubscriptions"

	DeleteSubscriptionTransaction raftstore.RequestType = "deleteSubscriptionTransaction"
	InsertSubscriptionTransaction raftstore.RequestType = "insertSubscriptionTransaction"
	UpdateSubscriptionTransaction raftstore.RequestType = "updateSubscriptionTransaction"
)

var readOnlyRequests = []raftstore.RequestType{
	getISA,
	searchISAs,
	listExpiredISAs,
	countISAs,

	getSubscription,
	searchSubscriptions,
	searchSubscriptionsByOwner,
	maxSubscriptionCountInCellsByOwner,
	listExpiredSubscriptions,
	countSubscriptions,
}

// repo is a full implementation of rid.repos.Repository for Raft-based storage.
type repo struct {
	consensus *consensus.Consensus
	memStore  *memstore.Store[repos.Repository]
}

func Init(ctx context.Context, logger *zap.Logger) (*raftstore.Store[repos.Repository], error) {
	params, err := ridraftparams.GetConnectParameters()
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get rid raft parameters")
	}

	memStore, err := ridmemstore.Init(ctx, logger)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to initialize RID memstore")
	}

	r := &repo{memStore: memStore}
	store, err := raftstore.Init(ctx, logger, params, r)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to initialize RID raftstore")
	}

	r.consensus = store.Consensus

	raftstoreadmin.Register(ctx, logger, "rid", store)

	return store, nil
}

func (r *repo) GetRepo() repos.Repository { return r }

func (r *repo) IsReadOnly(requestType raftstore.RequestType) bool {
	return slices.Contains(readOnlyRequests, requestType)
}

func (r *repo) GetSnapshot() ([]byte, error) {
	return r.memStore.GetSnapshot()
}

func (r *repo) RestoreFromSnapshot(data []byte) error {
	return r.memStore.RestoreFromSnapshot(data)
}

func (r *repo) Apply(ctx context.Context, proposal consensus.Proposal) (any, error) {
	mem, err := r.memStore.Interact(ctx)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to obtain rid memstore repository")
	}

	switch raftstore.RequestType(proposal.RequestType) {
	// ISAs

	case getISA:
		var p getISAPayload
		if err := json.Unmarshal(proposal.Value, &p); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", getISA)
		}
		return mem.GetISA(ctx, p.ID, p.ForUpdate)

	case deleteISA:
		var isa ridmodels.IdentificationServiceArea
		if err := json.Unmarshal(proposal.Value, &isa); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", deleteISA)
		}
		return mem.DeleteISA(ctx, &isa)

	case insertISA:
		var isa ridmodels.IdentificationServiceArea
		if err := json.Unmarshal(proposal.Value, &isa); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", insertISA)
		}
		return mem.InsertISA(ctx, &isa)

	case updateISA:
		var isa ridmodels.IdentificationServiceArea
		if err := json.Unmarshal(proposal.Value, &isa); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", updateISA)
		}
		return mem.UpdateISA(ctx, &isa)

	case searchISAs:
		var p searchISAsPayload
		if err := json.Unmarshal(proposal.Value, &p); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", searchISAs)
		}
		return mem.SearchISAs(ctx, p.Cells, p.Earliest, p.Latest)

	case listExpiredISAs:
		var p listExpiredISAsPayload
		if err := json.Unmarshal(proposal.Value, &p); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", listExpiredISAs)
		}
		return mem.ListExpiredISAs(ctx, p.Writer, p.Threshold)

	case countISAs:
		return mem.CountISAs(ctx)

	case DeleteISATransaction:
		return r.deleteISATransactionApplier(ctx, proposal, mem)

	case InsertISATransaction:
		return r.insertISATransactionApplier(ctx, proposal, mem)

	case UpdateISATransaction:
		return r.updateISATransactionApplier(ctx, proposal, mem)

	// Subscriptions

	case getSubscription:
		var id dssmodels.ID
		if err := json.Unmarshal(proposal.Value, &id); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", getSubscription)
		}
		return mem.GetSubscription(ctx, id)

	case deleteSubscription:
		var sub ridmodels.Subscription
		if err := json.Unmarshal(proposal.Value, &sub); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", deleteSubscription)
		}
		return mem.DeleteSubscription(ctx, &sub)

	case insertSubscription:
		var sub ridmodels.Subscription
		if err := json.Unmarshal(proposal.Value, &sub); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", insertSubscription)
		}
		return mem.InsertSubscription(ctx, &sub)

	case updateSubscription:
		var sub ridmodels.Subscription
		if err := json.Unmarshal(proposal.Value, &sub); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", updateSubscription)
		}
		return mem.UpdateSubscription(ctx, &sub)

	case searchSubscriptions:
		var cells s2.CellUnion
		if err := json.Unmarshal(proposal.Value, &cells); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", searchSubscriptions)
		}
		return mem.SearchSubscriptions(ctx, cells)

	case searchSubscriptionsByOwner:
		var p searchSubscriptionsByOwnerPayload
		if err := json.Unmarshal(proposal.Value, &p); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", searchSubscriptionsByOwner)
		}
		return mem.SearchSubscriptionsByOwner(ctx, p.Cells, p.Owner)

	case updateNotificationIdxsInCells:
		var cells s2.CellUnion
		if err := json.Unmarshal(proposal.Value, &cells); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", updateNotificationIdxsInCells)
		}
		return mem.UpdateNotificationIdxsInCells(ctx, cells)

	case maxSubscriptionCountInCellsByOwner:
		var p maxSubscriptionCountInCellsByOwnerPayload
		if err := json.Unmarshal(proposal.Value, &p); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", maxSubscriptionCountInCellsByOwner)
		}
		return mem.MaxSubscriptionCountInCellsByOwner(ctx, p.Cells, p.Owner)

	case listExpiredSubscriptions:
		var p listExpiredSubscriptionsPayload
		if err := json.Unmarshal(proposal.Value, &p); err != nil {
			return nil, stacktrace.Propagate(err, "failed to unmarshal %s payload", listExpiredSubscriptions)
		}
		return mem.ListExpiredSubscriptions(ctx, p.Writer, p.Threshold)

	case countSubscriptions:
		return mem.CountSubscriptions(ctx)

	case DeleteSubscriptionTransaction:
		return r.deleteSubscriptionTransactionApplier(ctx, proposal, mem)

	case InsertSubscriptionTransaction:
		return r.insertSubscriptionTransactionApplier(ctx, proposal, mem)

	case UpdateSubscriptionTransaction:
		return r.updateSubscriptionTransactionApplier(ctx, proposal, mem)

	default:
		return nil, stacktrace.NewError("unknown request type: %q", proposal.RequestType)
	}
}
