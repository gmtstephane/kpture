package k8s

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

type Forwarder interface {
	ForwardPorts() error
}

// GetKubeForwarder returns a port forwarder for a given pod
func GetKubeForwarder(
	r *rest.Config,
	path string,
	rch, sch chan struct{},
	proxyport int32,
) (*portforward.PortForwarder, int, error) {
	port, err := getFreePort()
	if err != nil {
		return nil, 0, err
	}

	hostIP := strings.TrimLeft(r.Host, "htps:/")
	transport, upgrader, err := spdy.RoundTripperFor(r)
	if err != nil {
		return nil, 0, err
	}

	dialer := spdy.NewDialer(
		upgrader,
		&http.Client{Transport: transport},
		http.MethodPost,
		&url.URL{Scheme: "https", Path: path, Host: hostIP},
	)

	fw, err := portforward.New(
		dialer,
		[]string{fmt.Sprintf("%d:%d", port, proxyport)},
		sch,
		rch,
		io.Discard,
		os.Stderr,
	)
	if err != nil {
		log.Println(err)
		return nil, 0, err
	}

	return fw, port, nil
}

func PortForward(forwarder Forwarder, readych chan struct{}, timeout time.Duration) error {
	errchan := make(chan error, 1)

	go func() {
		err := forwarder.ForwardPorts()
		if err != nil {
			errchan <- err
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	for {
		select {
		case <-ctx.Done():
			return errors.New("timeout waiting for port forward")
		case errForward := <-errchan:
			return errForward
		case <-readych:
			return nil
		}
	}
}

// getFreePort asks the kernel for a free open port that is ready to use.
func getFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}
