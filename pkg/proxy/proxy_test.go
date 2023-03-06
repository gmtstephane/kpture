package proxy

import (
	"context"
	"log"
	"net"
	"reflect"
	"sync"
	"testing"
	"time"

	capture "github.com/gmtstephane/kpture/api/kpture"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

func newServer(t *testing.T, register func(srv *grpc.Server)) *grpc.ClientConn {
	lis := bufconn.Listen(1024 * 1024)
	t.Cleanup(func() {
		lis.Close()
	})

	srv := grpc.NewServer()
	t.Cleanup(func() {
		srv.Stop()
	})
	register(srv)
	go func() {
		if err := srv.Serve(lis); err != nil {
			log.Fatalf("srv.Serve %v", err)
		}
	}()

	dialer := func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	t.Cleanup(func() {
		cancel()
	})

	conn, err := grpc.DialContext(
		ctx,
		"",
		grpc.WithContextDialer(dialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()))

	t.Cleanup(func() {
		conn.Close()
	})
	if err != nil {
		t.Fatalf("grpc.DialContext %v", err)
	}

	return conn
}

const defaultbufferSize = 1500

func TestProxy_AddPacket(t *testing.T) {
	p := NewProxyServer(defaultbufferSize, func(wg *sync.WaitGroup, cancel context.CancelFunc) {
		cancel()
		wg.Wait()
	})
	assert.NotNil(t, p)

	conn := newServer(t, func(srv *grpc.Server) {
		capture.RegisterAgentServiceServer(srv, p)
		capture.RegisterClientServiceServer(srv, p)
	})

	t.Cleanup(func() {
		conn.Close()
	})
	assert.NotNil(t, conn)

	agentService := capture.NewAgentServiceClient(conn)
	assert.NotNil(t, agentService)

	clientService := capture.NewClientServiceClient(conn)
	assert.NotNil(t, clientService)

	agentCtx, cancelAgentCtx := context.WithCancel(context.Background())
	clientCtx, cancelClientctx := context.WithCancel(context.Background())

	defer cancelAgentCtx()
	defer cancelClientctx()

	clientServiceStream, err := clientService.GetPackets(clientCtx, &capture.Empty{})
	assert.NoError(t, err)
	assert.NotNil(t, clientServiceStream)

	resp, err := agentService.Ready(agentCtx, &capture.Pod{})
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	agentServiceStream, err := agentService.AddPacket(agentCtx)
	assert.NoError(t, err)
	assert.NotNil(t, agentServiceStream)

	testdone := make(chan bool, 1)
	go func() {
		for {
			_, errRecv := clientServiceStream.Recv()
			if errRecv != nil {
				break
			}
			t.Log("Received packet")
			testdone <- true
		}
	}()
	time.Sleep(1 * time.Second)
	err = agentServiceStream.Send(&capture.Packet{})
	assert.NoError(t, err)
	err = agentServiceStream.Send(&capture.Packet{})
	assert.NoError(t, err)

	<-testdone // wait for the packet to be received
	cancelClientctx()
	// cancelAgentctx()
	time.Sleep(1 * time.Second)
	err = agentServiceStream.Send(&capture.Packet{})
	assert.NoError(t, err)
	// time.Sleep(10 * time.Second)
}

func TestProxy_Ready(t *testing.T) {
	type fields struct {
		packets                          chan *capture.Packet
		started                          bool
		readypods                        []*capture.Pod
		wg                               *sync.WaitGroup
		ctx                              context.Context
		cancel                           context.CancelFunc
		UnimplementedAgentServiceServer  capture.UnimplementedAgentServiceServer
		UnimplementedClientServiceServer capture.UnimplementedClientServiceServer
	}
	type args struct {
		ctx context.Context
		pod *capture.Pod
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *capture.Empty
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Proxy{
				packets:                          tt.fields.packets,
				started:                          tt.fields.started,
				readypods:                        tt.fields.readypods,
				wg:                               tt.fields.wg,
				ctx:                              tt.fields.ctx,
				cancel:                           tt.fields.cancel,
				UnimplementedAgentServiceServer:  tt.fields.UnimplementedAgentServiceServer,
				UnimplementedClientServiceServer: tt.fields.UnimplementedClientServiceServer,
			}
			got, err := s.Ready(tt.args.ctx, tt.args.pod)
			if (err != nil) != tt.wantErr {
				t.Errorf("Proxy.Ready() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Proxy.Ready() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProxy_GetPackets(t *testing.T) {
	type fields struct {
		packets                          chan *capture.Packet
		started                          bool
		readypods                        []*capture.Pod
		wg                               *sync.WaitGroup
		ctx                              context.Context
		cancel                           context.CancelFunc
		UnimplementedAgentServiceServer  capture.UnimplementedAgentServiceServer
		UnimplementedClientServiceServer capture.UnimplementedClientServiceServer
	}
	type args struct {
		in     *capture.Empty
		stream capture.ClientService_GetPacketsServer
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Proxy{
				packets:                          tt.fields.packets,
				started:                          tt.fields.started,
				readypods:                        tt.fields.readypods,
				wg:                               tt.fields.wg,
				ctx:                              tt.fields.ctx,
				cancel:                           tt.fields.cancel,
				UnimplementedAgentServiceServer:  tt.fields.UnimplementedAgentServiceServer,
				UnimplementedClientServiceServer: tt.fields.UnimplementedClientServiceServer,
			}
			if err := s.GetPackets(tt.args.in, tt.args.stream); (err != nil) != tt.wantErr {
				t.Errorf("Proxy.GetPackets() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
