package dssproto

import (
	"context"

	"google.golang.org/grpc"
)

// DSSServiceClient is the client API for DSSService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type DSSServiceClient interface {
	// DSS: /uas_reporters/{id}
	//
	// Delete reference to an Entity.  USSs should not delete PositionReporting Entities before the end of the last managed flight plus the retention period.
	DeleteUASReporter(ctx context.Context, in *DeleteUASReporterRequest, opts ...grpc.CallOption) (*DeleteUASReporterResponse, error)
	// DSS: /subscriptions/{id}
	//
	// Delete a subscription.
	DeleteSubscription(ctx context.Context, in *DeleteSubscriptionRequest, opts ...grpc.CallOption) (*DeleteSubscriptionResponse, error)
	// DSS: /uas_reporters
	//
	// Retrieve references to all visible airspace Entities in the DAR for a given area during the given time.  Note that some Entities returned will lie entirely outside the requested area because an individual DAR cell cannot filter EntityReferences by exact geography.
	//
	// Only PositionReporting Entities shall be visible to clients providing the `dss.read.position_reporting_entities` scope.
	SearchEntityReferences(ctx context.Context, in *SearchUASReportersRequest, opts ...grpc.CallOption) (*SearchUASReportersResponse, error)
	// DSS: /subscriptions
	//
	// Retrieve subscriptions intersecting an area of interest.  Subscription notifications are only triggered by (and contain details of) changes to Entities in the DSS; they do not involve any other data transfer such as remote ID telemetry updates.  Only Subscriptions belonging to the caller are returned.
	SearchSubscriptions(ctx context.Context, in *SearchSubscriptionsRequest, opts ...grpc.CallOption) (*SearchSubscriptionsResponse, error)
	// DSS: /subscriptions/{id}
	//
	// Verify the existence/valdity and state of a particular subscription.
	GetSubscription(ctx context.Context, in *GetSubscriptionRequest, opts ...grpc.CallOption) (*GetSubscriptionResponse, error)
	// DSS: /uas_reporters/{id}
	//
	// Create or update reference to an Entity.  Unless otherwise specified, the EntityType of an existing Entity may not be changed.
	//
	// The `details_url` field in the request body is required for all Entities except PositionReporting Entities.
	//
	// `PositionReporting Entities`:
	// Authorization scope `dss.write.position_reporting_entities` is required.  The DSS assumes the USS has already added the appropriate retention period to operation end time in EntityReference's `time_end` extents field before storing it.  Updating `time_start` is not allowed if it is before the current time.
	//
	// `Operation Entities`:
	// Authorization scope `dss.write.operation_entities` is required.
	PutUASReporter(ctx context.Context, in *PutUASReporterRequest, opts ...grpc.CallOption) (*PutUASReporterResponse, error)
	// DSS: /subscriptions/{id}
	//
	// Create or update a subscription.  Subscription notifications are only triggered by (and contain details of) changes to Entities in the DSS; they do not involve any other data transfer such as remote ID telemetry updates.
	//
	// Note that the types of content that should be sent to the created subscription depends on the scope in the provided access token.
	PutSubscription(ctx context.Context, in *PutSubscriptionRequest, opts ...grpc.CallOption) (*PutSubscriptionResponse, error)
}

type dSSServiceClient struct {
	cc *grpc.ClientConn
}

func NewDSSServiceClient(cc *grpc.ClientConn) DSSServiceClient {
	return &dSSServiceClient{cc}
}

func (c *dSSServiceClient) DeleteUASReporter(ctx context.Context, in *DeleteUASReporterRequest, opts ...grpc.CallOption) (*DeleteUASReporterResponse, error) {
	out := new(DeleteUASReporterResponse)
	err := c.cc.Invoke(ctx, "/dss.DSSService/DeleteUASReporter", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *dSSServiceClient) DeleteSubscription(ctx context.Context, in *DeleteSubscriptionRequest, opts ...grpc.CallOption) (*DeleteSubscriptionResponse, error) {
	out := new(DeleteSubscriptionResponse)
	err := c.cc.Invoke(ctx, "/dss.DSSService/DeleteSubscription", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *dSSServiceClient) SearchEntityReferences(ctx context.Context, in *SearchUASReportersRequest, opts ...grpc.CallOption) (*SearchUASReportersResponse, error) {
	out := new(SearchUASReportersResponse)
	err := c.cc.Invoke(ctx, "/dss.DSSService/SearchEntityReferences", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *dSSServiceClient) SearchSubscriptions(ctx context.Context, in *SearchSubscriptionsRequest, opts ...grpc.CallOption) (*SearchSubscriptionsResponse, error) {
	out := new(SearchSubscriptionsResponse)
	err := c.cc.Invoke(ctx, "/dss.DSSService/SearchSubscriptions", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *dSSServiceClient) GetSubscription(ctx context.Context, in *GetSubscriptionRequest, opts ...grpc.CallOption) (*GetSubscriptionResponse, error) {
	out := new(GetSubscriptionResponse)
	err := c.cc.Invoke(ctx, "/dss.DSSService/GetSubscription", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *dSSServiceClient) PutUASReporter(ctx context.Context, in *PutUASReporterRequest, opts ...grpc.CallOption) (*PutUASReporterResponse, error) {
	out := new(PutUASReporterResponse)
	err := c.cc.Invoke(ctx, "/dss.DSSService/PutUASReporter", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *dSSServiceClient) PutSubscription(ctx context.Context, in *PutSubscriptionRequest, opts ...grpc.CallOption) (*PutSubscriptionResponse, error) {
	out := new(PutSubscriptionResponse)
	err := c.cc.Invoke(ctx, "/dss.DSSService/PutSubscription", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}
