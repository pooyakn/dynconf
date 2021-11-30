# Dynamic service configuration

This Go package provides a dynamic service configuration backed by etcd,
so there should be no need to redeploy a service to change its settings.

For example, project Curiosity expects settings such as `velocity = 10` and `is_camera_enabled = true`.
Let's save them in etcd with a path prefix `configs/curiosity/`.

```sh
$ brew install etcd
$ etcd
$ etcdctl put configs/curiosity/velocity 10
OK
$ etcdctl put configs/curiosity/is_camera_enabled true
OK
```

On the service side the settings are fetched from the same path.

```go
const defaultVelocity = 5

c, err := dynconf.New("configs/curiosity")
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
