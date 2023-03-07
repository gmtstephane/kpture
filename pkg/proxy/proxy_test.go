package proxy

import (
	"context"
	"log"
	"net"
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
	err = agentServiceStream.Send(&capture.PacketDescriptor{})
	assert.NoError(t, err)
	err = agentServiceStream.Send(&capture.PacketDescriptor{})
	assert.NoError(t, err)

	<-testdone // wait for the packet to be received
	cancelClientctx()
	// cancelAgentctx()
	time.Sleep(1 * time.Second)
	err = agentServiceStream.Send(&capture.PacketDescriptor{})
	assert.NoError(t, err)
	// time.Sleep(10 * time.Second)
}
