/*
Program watcher periodically prints all the settings stored at the given etcd path
so you can observe the changes made via etcdctl.
For example, run the watcher and set the following keys.

	etcdctl put configs/curiosity/velocity 10
	etcdctl put configs/curiosity/is_camera_enabled true
	etcdctl put configs/curiosity/velocity 20
	etcdctl del configs/curiosity/velocity

You should see that the updated settings are printed.
*/
package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/marselester/dynconf"
)

func main() {
	// By default an exit code is set to indicate a failure since
	// there are more failure scenarios to begin with.
	exitCode := 1
	defer func() { os.Exit(exitCode) }()

	path := flag.String("path", "configs/curiosity/", "path (etcd key prefix) in etcd where settings are stored")
	interval := flag.Duration("interval", 5*time.Second, "how often the settings shall be printed")
	flag.Parse()

	conf, err := dynconf.New(*path)
	if err != nil {
		// No worries if etcd is down, the rover can still roll with the default settings.
		log.Printf("dynconf failed to connect to etcd: %v", err)
	}
	defer conf.Close()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

Loop:
	for {
		select {
		case <-ctx.Done():
			break Loop
		case <-time.After(*interval):
			log.Println(conf.Settings())
		}
	}

	// The program terminates successfully if it received INT/TERM signal.
	exitCode = 0
}
