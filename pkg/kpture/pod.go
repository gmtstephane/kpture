package kpture

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gmtstephane/kpture/api/capture"
	"github.com/gmtstephane/kpture/pkg/pcap"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

const (
	defaultPullTimeout = 60
)

type Pod struct {
	name           string
	debugContainer string
	namespace      string
	pcapOptions    pcap.Options
	localPort      int
	fw             *portforward.PortForwarder
	stopCh         chan struct{}
	errCh          chan<- error
	readyCh        chan struct{}
	capture        capture.Kpture_PacketsStreamClient
	log            *logrus.Entry
}

type PodInterface interface {
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*v1.Pod, error)
	List(ctx context.Context, opts metav1.ListOptions) (*v1.PodList, error)
	UpdateEphemeralContainers(ctx context.Context, podName string, pod *v1.Pod, opts metav1.UpdateOptions) (*v1.Pod, error)
}

func NewKpturePod(name string, ns string, id string, pcapOptions pcap.Options, errchan chan error) (*Pod, error) {
	localPort, err := getFreePort()
	if err != nil {
		return nil, err
	}

	k := &Pod{
		name:           name,
		namespace:      ns,
		debugContainer: DebugContainerName + "-" + id,
		pcapOptions:    pcapOptions,
		localPort:      localPort,
		readyCh:        make(chan struct{}),
		stopCh:         make(chan struct{}),
		errCh:          errchan,
		log:            logrus.WithField("Name", name).WithField("Namespace", ns),
	}
	return k, nil
}

func (k *Pod) CreateDebugContainer(client PodInterface) error {
	k.log.Info("Creating debug container")
	err := k.InjectContainer(client, k.debugContainer)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(defaultPullTimeout)*time.Second)

	defer cancel()
	for {
		select {
		case <-ctx.Done():
			return errors.New("timeout waiting for debug container to be ready")
		default:
			pod, errGetPod := client.Get(context.Background(), k.name, metav1.GetOptions{})
			if errGetPod != nil {
				return errGetPod
			}
			for _, eph := range pod.Status.EphemeralContainerStatuses {
				if eph.Name == k.debugContainer && eph.State.Running != nil {
					return nil
				}
			}
			time.Sleep(1 * time.Second)
			k.log.Info("Waiting for debug container to be ready...")
		}
	}
}

func (k *Pod) InjectContainer(client PodInterface, name string) error {
	pod, err := client.Get(context.Background(), k.name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	debugPod := generateDebugContainer(pod, name, k.pcapOptions)

	_, err = client.UpdateEphemeralContainers(context.Background(), pod.Name, debugPod, metav1.UpdateOptions{})
	if err != nil {
		logrus.Error(err)
		return err
	}

	return nil
}

func (k *Pod) PortForwardAPod(restConf *rest.Config) error {
	k.log.Info("Port forwarding to pod ", k.name)
	devnull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0o755)
	if err != nil {
		k.log.Error(err)
		return err
	}

	path := fmt.Sprintf("/api/v1/namespaces/%s/pods/%s/portforward", k.namespace, k.name)
	hostIP := strings.TrimLeft(restConf.Host, "htps:/")
	transport, upgrader, err := spdy.RoundTripperFor(restConf)
	if err != nil {
		k.log.Error(err)
		return err
	}

	dialer := spdy.NewDialer(
		upgrader,
		&http.Client{Transport: transport},
		http.MethodPost,
		&url.URL{Scheme: "https", Path: path, Host: hostIP},
	)
	k.fw, err = portforward.New(
		dialer,
		[]string{fmt.Sprintf("%d:%d", k.localPort, k.pcapOptions.Port)},
		k.stopCh,
		k.readyCh,
		devnull,
		os.Stderr,
	)

	if err != nil {
		k.log.Error(err)
		return err
	}

	go func() {
		err = k.fw.ForwardPorts()
		if err != nil {
			k.log.Error(err)
			k.errCh <- err
		}
	}()

	<-k.readyCh

	return err
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

func (k *Pod) Close() {
	k.log.Info("Closing port forward  and grpc client for pod ", k.name)
	if k.capture != nil {
		err := k.capture.CloseSend()
		if err != nil {
			k.log.Error(err)
			k.errCh <- err
		}
	}
	if k.fw != nil {
		k.fw.Close()
	}
}

func (k *Pod) ReadPackets(packetCh chan *PacketCapture) {
	k.log.Info("Reading packets from pod ", k.name)
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	target := "localhost:" + fmt.Sprintf("%d", k.localPort)
	conn, err := grpc.Dial(target, opts...)
	if err != nil {
		k.log.Error(err)
		k.errCh <- err
	}

	defer conn.Close()

	client := capture.NewKptureClient(conn)

	k.capture, err = client.PacketsStream(context.Background(), &capture.Empty{})
	if err != nil {
		k.log.Error(err)
		k.errCh <- err
		return
	}

	for {
		packet, errRecv := k.capture.Recv()
		if errors.Is(errRecv, io.EOF) {
			break
		}
		if errRecv != nil {
			k.log.Error(err)
			k.errCh <- err
			break
		}

		select {
		case packetCh <- &PacketCapture{
			Packet: packet,
			Pod:    k.name,
		}:
		default:
			k.errCh <- errors.New("packet channel is full")
		}
	}
}
