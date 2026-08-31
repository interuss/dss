package application

import (
	"context"
	"time"

	"github.com/golang/geo/s2"
	dssmodels "github.com/interuss/dss/pkg/models"
	ridmodels "github.com/interuss/dss/pkg/rid/models"
)

type subscriptionStore struct {
	subs map[dssmodels.ID]*ridmodels.Subscription
}

func (store *subscriptionStore) GetSubscription(ctx context.Context, id dssmodels.ID) (*ridmodels.Subscription, error) {
	if sub, ok := store.subs[id]; ok {
		return sub, nil
	}
	return nil, nil
}

// DeleteSubscription deletes the Subscription identified by "id" and owned by "owner".
// Returns the delete Subscription and all IdentificationServiceAreas affected by the delete.
func (store *subscriptionStore) DeleteSubscription(ctx context.Context, s *ridmodels.Subscription) (*ridmodels.Subscription, error) {
	if sub, ok := store.subs[s.ID]; ok {
		delete(store.subs, s.ID)
		return sub, nil
	}
	return nil, nil
}

func (store *subscriptionStore) InsertSubscription(ctx context.Context, s *ridmodels.Subscription) (*ridmodels.Subscription, error) {
	storedCopy := *s
	storedCopy.Version = dssmodels.VersionFromTime(time.Now())
	store.subs[s.ID] = &storedCopy

	returnedCopy := storedCopy
	return &returnedCopy, nil
}

func (store *subscriptionStore) UpdateSubscription(ctx context.Context, s *ridmodels.Subscription) (*ridmodels.Subscription, error) {
	storedCopy := *s
	storedCopy.Version = dssmodels.VersionFromTime(time.Now())
	store.subs[s.ID] = &storedCopy

	returnedCopy := storedCopy
	return &returnedCopy, nil
}

func (store *subscriptionStore) SearchSubscriptionsByOwner(ctx context.Context, cells s2.CellUnion, owner dssmodels.Owner) ([]*ridmodels.Subscription, error) {
	var subs []*ridmodels.Subscription

	res, _ := store.SearchSubscriptions(ctx, cells)
	for _, s := range res {
		if s.Owner == owner {
			subs = append(subs, s)
		}
	}
	return subs, nil
}

func (store *subscriptionStore) UpdateNotificationIdxsInCells(ctx context.Context, cells s2.CellUnion) ([]*ridmodels.Subscription, error) {
	subs, _ := store.SearchSubscriptions(ctx, cells)
	for i := range subs {
		subs[i].NotificationIndex++
	}
	return subs, nil
}

func (store *subscriptionStore) MaxSubscriptionCountInCellsByOwner(ctx context.Context, cells s2.CellUnion, owner dssmodels.Owner) (int, error) {
	maxValue := 0
	subs, _ := store.SearchSubscriptionsByOwner(ctx, cells, owner)

	cellMap := make(map[s2.CellID]int)
	for _, s := range subs {
		for _, cid := range s.Cells {
			if _, ok := cellMap[cid]; !ok {
				cellMap[cid] = 1
			} else {
				cellMap[cid]++
			}
			if cellMap[cid] > maxValue {
				maxValue = cellMap[cid]
			}
		}
	}
	return maxValue, nil
}

func (store *subscriptionStore) SearchSubscriptions(ctx context.Context, cells s2.CellUnion) ([]*ridmodels.Subscription, error) {
	var subs []*ridmodels.Subscription
	for _, s := range store.subs {
		// Don't call Intersects, since that's smarter code than we implement in the DB.
		appended := false
		for _, c1 := range s.Cells {
			for _, c2 := range cells {
				if c1 == c2 {
					subs = append(subs, s)
					appended = true
					break
				}
			}
			if appended {
				break
			}
		}
	}
	return subs, nil
}

func (store *subscriptionStore) ListExpiredSubscriptions(ctx context.Context, writer string, threshold time.Time) ([]*ridmodels.Subscription, error) {
	return make([]*ridmodels.Subscription, 0), nil
}

// Implements repos.ISA.CountSubscriptions
func (store *subscriptionStore) CountSubscriptions(ctx context.Context) (int64, error) {
	return int64(len(store.subs)), nil
}
