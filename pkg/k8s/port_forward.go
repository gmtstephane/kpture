package k8s

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"

	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

func (k KubeClient) PortForward(id string) (chan struct{}, int, error) {
	port, err := getFreePort()
	if err != nil {
		return nil, 0, err
	}
	name := "kpture-proxy-" + id

	path := fmt.Sprintf("/api/v1/namespaces/%s/pods/%s/portforward", k.Namespace, name)
	hostIP := strings.TrimLeft(k.RestConf.Host, "htps:/")
	transport, upgrader, err := spdy.RoundTripperFor(k.RestConf)
	if err != nil {
		return nil, port, err
	}

	dialer := spdy.NewDialer(
		upgrader,
		&http.Client{Transport: transport},
		http.MethodPost,
		&url.URL{Scheme: "https", Path: path, Host: hostIP},
	)
	readychan, stopchan, errchan := make(chan struct{}), make(chan struct{}), make(chan error)
	fw, err := portforward.New(
		dialer,
		[]string{fmt.Sprintf("%d:%d", port, 10000)},
		stopchan,
		readychan,
		io.Discard,
		os.Stderr,
	)
	if err != nil {
		return nil, port, err
	}

	go func() {
		err = fw.ForwardPorts()
		if err != nil {
			errchan <- err
		}
	}()

	for {
		select {
		case errForward := <-errchan:
			return nil, port, errForward
		case <-readychan:
			return stopchan, port, nil
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
