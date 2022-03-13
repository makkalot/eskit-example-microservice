package clients

import (
	"context"
	"fmt"
	"github.com/makkalot/eskit/generated/grpc/go/common"
	store "github.com/makkalot/eskit/generated/grpc/go/eventstore"
	common2 "github.com/makkalot/eskit/lib/common"
	"github.com/makkalot/eskit/lib/eventstore"
	"google.golang.org/grpc"
	"strconv"
)

// estoreGrpc implements eventstore.Store via grpc
type estoreGRPC struct {
	ctx        context.Context
	grpcClient store.EventstoreServiceClient
}

func (e estoreGRPC) Append(event *store.Event) error {
	_, err := e.grpcClient.Append(e.ctx, &store.AppendEventRequest{Event: event})
	return err
}

func (e estoreGRPC) Get(originator *common.Originator, fromVersion bool) ([]*store.Event, error) {
	resp, err := e.grpcClient.GetEvents(e.ctx, &store.GetEventsRequest{
		Originator: originator,
	})
	if err != nil {
		return nil, err
	}

	return resp.GetEvents(), nil
}

func (e estoreGRPC) Logs(fromID uint64, size uint32, pipelineID string) ([]*store.AppLogEntry, error) {
	resp, err := e.grpcClient.Logs(e.ctx, &store.AppLogRequest{
		FromId:     strconv.FormatUint(fromID, 10),
		Size:       size,
		PipelineId: pipelineID,
	})

	if err != nil {
		return nil, err
	}

	return resp.Results, nil
}

// NewStoreClient creates a new instance of eventstore.Store
func NewStoreClient(ctx context.Context, storeEndpoint string) (eventstore.Store, error) {
	client, err := NewStoreClientWithWait(ctx, storeEndpoint)
	if err != nil {
		return nil, err
	}

	return &estoreGRPC{
		ctx:        ctx,
		grpcClient: client,
	}, nil
}

// Creates a new EventstoreServiceClient but first waits for health endpoint to become ready
func NewStoreClientWithWait(ctx context.Context, storeEndpoint string) (store.EventstoreServiceClient, error) {
	var conn *grpc.ClientConn
	var storeClient store.EventstoreServiceClient

	err := common2.RetryNormal(func() error {
		var err error
		conn, err = grpc.Dial(storeEndpoint, grpc.WithInsecure())
		if err != nil {
			return err
		}

		storeClient = store.NewEventstoreServiceClient(conn)
		_, err = storeClient.Healtz(ctx, &store.HealthRequest{})
		if err != nil {
			return err
		}

		return nil
	})
	return storeClient, err
}

type eventStoreClientWithNoNetworking struct {
	server store.EventstoreServiceServer
}

// NewEventStoreServiceClientWithNoNetworking is useful for unittesting where you can embed the server directly inside the
// client and can do all kinds of tests without having to worry about the networking bit and spinning up servers
func NewEventStoreServiceClientWithNoNetworking(server store.EventstoreServiceServer) store.EventstoreServiceClient {
	return &eventStoreClientWithNoNetworking{
		server: server,
	}
}

func (c *eventStoreClientWithNoNetworking) Healtz(ctx context.Context, in *store.HealthRequest, opts ...grpc.CallOption) (*store.HealthResponse, error) {
	return c.server.Healtz(ctx, in)
}

func (c *eventStoreClientWithNoNetworking) Append(ctx context.Context, in *store.AppendEventRequest, opts ...grpc.CallOption) (*store.AppendEventResponse, error) {
	return c.server.Append(ctx, in)
}

func (c *eventStoreClientWithNoNetworking) GetEvents(ctx context.Context, in *store.GetEventsRequest, opts ...grpc.CallOption) (*store.GetEventsResponse, error) {
	return c.server.GetEvents(ctx, in)
}

func (c *eventStoreClientWithNoNetworking) Logs(ctx context.Context, in *store.AppLogRequest, opts ...grpc.CallOption) (*store.AppLogResponse, error) {
	return c.server.Logs(ctx, in)
}

func (c *eventStoreClientWithNoNetworking) LogsPoll(ctx context.Context, in *store.AppLogRequest, opts ...grpc.CallOption) (store.EventstoreService_LogsPollClient, error) {
	return nil, fmt.Errorf("not implemented")
}
