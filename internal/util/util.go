package util

import (
	"context"
	"fmt"
	"hash/fnv"
	"k8s.io/klog/v2"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/nats-io/nats.go"
)

const (
	natsConnectionTimeout       = 350 * time.Millisecond
	natsConnectionRetryInterval = 100 * time.Millisecond
	NatsRequestTimeout          = 10 * time.Second
	natsScanRequestTimeout      = 100 * time.Millisecond
	dockerHubRateLimitDelay     = 1 * time.Hour
)

// NewConnection creates a new NATS connection
func NewConnection(addr, credFile string) (nc *nats.Conn, err error) {
	hostname, _ := os.Hostname()
	opts := []nats.Option{
		nats.Name(fmt.Sprintf("scanner-backend.%s", hostname)),
		nats.MaxReconnects(-1),
		nats.ErrorHandler(errorHandler),
		nats.ReconnectHandler(reconnectHandler),
		nats.DisconnectErrHandler(disconnectHandler),
		// nats.UseOldRequestStyle(),
	}

	if _, err := os.Stat(credFile); os.IsNotExist(err) {
		var username, password string
		if v, ok := os.LookupEnv("NATS_USERNAME"); ok {
			username = v
		} else {
			username = os.Getenv("THIS_IS_NATS_USERNAME")
		}
		if v, ok := os.LookupEnv("NATS_PASSWORD"); ok {
			password = v
		} else {
			password = os.Getenv("THIS_IS_NATS_PASSWORD")
		}
		opts = append(opts, nats.UserInfo(username, password))
	} else {
		opts = append(opts, nats.UserCredentials(credFile))
	}

	//if os.Getenv("NATS_CERTIFICATE") != "" && os.Getenv("NATS_KEY") != "" {
	//	opts = append(opts, nats.ClientCert(os.Getenv("NATS_CERTIFICATE"), os.Getenv("NATS_KEY")))
	//}
	//
	//if os.Getenv("NATS_CA") != "" {
	//	opts = append(opts, nats.RootCAs(os.Getenv("NATS_CA")))
	//}

	// initial connections can error due to DNS lookups etc, just retry, eventually with backoff
	ctx, cancel := context.WithTimeout(context.Background(), natsConnectionTimeout)
	defer cancel()

	ticker := time.NewTicker(natsConnectionRetryInterval)
	for {
		select {
		case <-ticker.C:
			nc, err := nats.Connect(addr, opts...)
			if err == nil {
				return nc, nil
			}
			klog.V(5).InfoS("failed to connect to event receiver", "error", err)
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}

// called during errors subscriptions etc
func errorHandler(nc *nats.Conn, s *nats.Subscription, err error) {
	if s != nil {
		klog.V(5).Infof("error in event receiver connection: %s: subscription: %s: %s", nc.ConnectedUrl(), s.Subject, err)
		return
	}
	klog.V(5).Infof("Error in event receiver connection: %s: %s", nc.ConnectedUrl(), err)
}

// called after reconnection
func reconnectHandler(nc *nats.Conn) {
	klog.V(5).Infof("Reconnected to %s", nc.ConnectedUrl())
}

// called after disconnection
func disconnectHandler(nc *nats.Conn, err error) {
	if err != nil {
		klog.V(5).Infof("Disconnected from event receiver due to error: %v", err)
	} else {
		klog.V(5).Infof("Disconnected from event receiver")
	}
}

// SigHandler sets up interrupt signal handlers
func SigHandler() chan os.Signal {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT)

	return sigs
}

// Partition calculates a partition to publish data to based on the hostname
// always yields the same partition for the same name. Uses SHIPPER_PARTITIONS
// to determine the amount of desired partitions. SHIPPER_PARTITIONS of 0 will
// always return partition 0
func Partition() (part uint32, name string, err error) {
	name, err = hostname()
	if err != nil {
		return 0, "", fmt.Errorf("could not determine the hostname, set HOSTNAME to override: %s", err)
	}

	partitions := 0

	partString := os.Getenv("SHIPPER_PARTITIONS")
	if partString != "" {
		partitions, err = strconv.Atoi(partString)
		if err != nil {
			return 0, name, fmt.Errorf("could not process partitions as an integer: %s", err)
		}
	}

	if partitions == 0 {
		return 0, name, nil
	}

	if partitions < 0 {
		return 0, name, fmt.Errorf("partitions has to be >= 0")
	}

	h := fnv.New32a()
	h.Write([]byte(name))

	return h.Sum32() % uint32(partitions), name, nil
}

func hostname() (string, error) {
	if os.Getenv("HOSTNAME") != "" {
		return os.Getenv("HOSTNAME"), nil
	}

	return os.Hostname()
}
