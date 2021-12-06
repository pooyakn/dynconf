# Dynamic service configuration

This Go package provides a dynamic service configuration backed by etcd,
so there should be no need to redeploy a service to change its settings.

For example, project Curiosity expects settings such as `velocity = 10` and `is_camera_enabled = true`.
Let's save them in etcd with a path prefix `/configs/curiosity/`.

```sh
$ brew install etcd
$ etcd
$ etcdctl put /configs/curiosity/velocity 10
OK
$ etcdctl put /configs/curiosity/is_camera_enabled true
OK
```

On the service side the settings are fetched from the same path.

```go
const defaultVelocity = 5

c, err := dynconf.New("/configs/curiosity/")
if err != nil {
	// No worries if etcd is down, the rover can still roll with the default settings.
}
defer c.Close()

rover.SetVelocity(
	c.Int("velocity", defaultVelocity),
)
```

## Testing

Run etcd (`127.0.0.1:2379` by default) and then launch the tests.

```sh
$ etcd
$ go test -count=1 .
```

### Toxiproxy

With [Toxiproxy](https://github.com/Shopify/toxiproxy) you can simulate network conditions.
All communication with the Toxiproxy daemon from the client happens via HTTP on port 8474.

```sh
$ brew install toxiproxy
$ toxiproxy-server
INFO[0000] API HTTP server starting                      host=localhost port=8474 version=2.2.0
```

Create a proxy called `dynconf_etcd` that listens on port `22379`
and proxies traffic to etcd running on `2379` port.

```sh
$ toxiproxy-cli create --listen localhost:22379 --upstream localhost:2379 dynconf_etcd
Created new proxy dynconf_etcd
$ toxiproxy-cli list
Name			Listen		Upstream		Enabled		Toxics
======================================================================================
dynconf_etcd		127.0.0.1:22379	localhost:2379		enabled		None
```

Check that etcd is reachable via Toxiproxy.

```sh
$ etcdctl --endpoints=127.0.0.1:22379 get /configs/curiosity/velocity
OK
```

Add 1500ms downstream latency.

```sh
$ toxiproxy-cli toxic add --type latency --attribute latency=1500 dynconf_etcd
$ toxiproxy-cli inspect dynconf_etcd
Name: dynconf_etcd	Listen: 127.0.0.1:22379	Upstream: localhost:2379
======================================================================
Upstream toxics:
Proxy has no Upstream toxics enabled.

Downstream toxics:
latency_downstream:	type=latency	stream=downstream	toxicity=1.00	attributes=[	jitter=0	latency=5000	]
```

There should be 1.5s latency introduced.
You can remove `latency_downstream` afterwards.

```sh
$ etcdctl --endpoints=127.0.0.1:22379 get /configs/curiosity/velocity
OK
$ toxiproxy-cli toxic remove -n latency_downstream dynconf_etcd
Removed toxic 'latency_downstream' on proxy 'dynconf_etcd'
```

There are more [toxics](https://github.com/Shopify/toxiproxy#toxics) available.
For example, stop all data from getting through, and close the connection after 10s.

```sh
$ toxiproxy-cli toxic add --type timeout --attribute timeout=10000 dynconf_etcd
Added downstream timeout toxic 'timeout_downstream' on proxy 'dynconf_etcd'
$ toxiproxy-cli toxic remove -n timeout_downstream dynconf_etcd
```

Simulate TCP RESET (Connection reset by peer) by closing connections after 10s.

```sh
$ toxiproxy-cli toxic add --type reset_peer --attribute timeout=10000 dynconf_etcd
Added downstream reset_peer toxic 'reset_peer_downstream' on proxy 'dynconf_etcd'
$ toxiproxy-cli toxic remove -n reset_peer_downstream dynconf_etcd
```
