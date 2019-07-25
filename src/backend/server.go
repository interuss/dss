package backend

import (
	dspb "InterUSS-Platform/src/dssproto"
	"context"
)

type DSSServer struct {
	store DSSStorageInterface
}

func NewServer() (*DSSServer, error) {
	crdb, err := NewCRDB()
	if err != nil {
		return nil, err
	}
	return &DSSServer{
		store: crdb,
	}, nil
}

func (s *DSSServer) DeleteUASReporter(ctx context.Context, req *dspb.DeleteUASReporterRequest) (*dspb.DeleteUASReporterResponse, error) {
	return nil, nil
}

func (s *DSSServer) DeleteSubscription(ctx context.Context, req *dspb.DeleteSubscriptionRequest) (*dspb.DeleteSubscriptionResponse, error) {
	return nil, nil
}

func (s *DSSServer) SearchEntityReferences(ctx context.Context, req *dspb.SearchUASReportersRequest) (*dspb.SearchUASReportersResponse, error) {
	return nil, nil
}

func (s *DSSServer) SearchSubscriptions(ctx context.Context, req *dspb.SearchSubscriptionsRequest) (*dspb.SearchSubscriptionsResponse, error) {
	return nil, nil
}

func (s *DSSServer) GetSubscription(ctx context.Context, req *dspb.GetSubscriptionRequest) (*dspb.GetSubscriptionResponse, error) {
	return nil, nil
}

func (s *DSSServer) PutUASReporter(ctx context.Context, req *dspb.PutUASReporterRequest) (*dspb.PutUASReporterResponse, error) {
	return nil, nil
}

func (s *DSSServer) PutSubscription(ctx context.Context, req *dspb.PutSubscriptionRequest) (*dspb.PutSubscriptionResponse, error) {
	return nil, nil
}
