/*
Program watcher periodically prints all the settings stored at the given etcd path
so you can observe the changes made via etcdctl.
For example, run the watcher and set the following keys.

	etcdctl put /configs/curiosity/velocity 10
	etcdctl put /configs/curiosity/is_camera_enabled true
	etcdctl put /configs/curiosity/velocity 20
	etcdctl del /configs/curiosity/velocity

You should see that the updated settings are printed.
*/
package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-kit/log"
	"github.com/marselester/dynconf"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func main() {
	// By default an exit code is set to indicate a failure since
	// there are more failure scenarios to begin with.
	exitCode := 1
	defer func() { os.Exit(exitCode) }()

	endpoints := flag.String("endpoints", "127.0.0.1:2379", "etcd endpoints")
	path := flag.String("path", "/configs/curiosity/", "path (etcd key prefix) in etcd where settings are stored")
	interval := flag.Duration("interval", 5*time.Second, "how often the settings shall be printed")
	flag.Parse()

	logger := log.NewJSONLogger(log.NewSyncWriter(os.Stderr))
	logger = log.With(logger, "ts", log.DefaultTimestampUTC)
	logger = log.With(logger, "caller", log.DefaultCaller)

	c, err := clientv3.New(clientv3.Config{
		Endpoints: strings.Split(*endpoints, ","),
	})
	if err != nil {
		logger.Log("msg", "failed to create etcd client", "err", err)
	}

	conf, err := dynconf.New(
		*path,
		dynconf.WithLogger(logger),
		dynconf.WithEtcdClient(c),
	)
	if err != nil {
		// No worries if etcd is down, the rover can still roll with the default settings.
		logger.Log("msg", "dynconf failed to connect to etcd", "err", err)
	}
	defer func() {
		if err := conf.Close(); err != nil {
			logger.Log("msg", "dynconf failed to close etcd connection", "err", err)
		}
	}()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

Loop:
	for {
		select {
		case <-ctx.Done():
			break Loop
		case <-time.After(*interval):
			logger.Log("settings", conf.Settings())
		}
	}

	// The program terminates successfully if it received INT/TERM signal.
	exitCode = 0
}
