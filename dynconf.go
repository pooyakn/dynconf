// Package dynconf provides a dynamic configuration backed by etcd.
// It can be used to access your project's settings without redeploying it every time a value changes.
package dynconf

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/go-kit/log"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// Option sets up a Config.
type Option func(*Config)

// WithEtcdClient sets the underlying etcd client.
func WithEtcdClient(etcd *clientv3.Client) Option {
	return func(c *Config) {
		c.etcd = etcd
	}
}

// WithLogger sets a logger to monitor possible syntax errors in setting values.
func WithLogger(logger log.Logger) Option {
	return func(c *Config) {
		c.logger = logger
	}
}

// Config provides an access to a project's settings stored in etcd.
type Config struct {
	// path (etcd key prefix) is the path to the project's config where settings are stored.
	path string
	// settings map holds the project's settings obtained from etcd.
	settings *sync.Map
	etcd     *clientv3.Client
	logger   log.Logger
}

// New returns a Config which can be set up with Option functions.
// By default an etcd client connects to 127.0.0.1:2379 gRPC endpoint.
// Note, the path to a config in etcd should be set to isolate config settings of different projects.
//
// For example, project Curiosity might have settings such as velocity and is_camera_enabled.
// If the path is /configs/curiosity/, then the settings would be stored as the following etcd keys:
// /configs/curiosity/velocity and /configs/curiosity/is_camera_enabled.
func New(path string, options ...Option) (*Config, error) {
	c := Config{
		path:     path,
		settings: &sync.Map{},
		logger:   log.NewNopLogger(),
	}
	for _, opt := range options {
		opt(&c)
	}

	if c.etcd == nil {
		var err error
		c.etcd, err = clientv3.New(clientv3.Config{
			Endpoints: []string{"127.0.0.1:2379"},
		})
		if err != nil {
			return nil, err
		}
	}
	go c.watch()

	return &c, nil
}

// Close closes the underlying etcd client.
func (c *Config) Close() error {
	return c.etcd.Close()
}

// load fetches all the settings from etcd for the configured path.
func (c *Config) load() error {
	r, err := c.etcd.Get(context.Background(), c.path, clientv3.WithPrefix())
	if err != nil {
		return err
	}

	// prefixLen is the length of the key prefix (path) in etcd to extract a setting name.
	prefixLen := len(c.path)
	for i := 0; i < len(r.Kvs); i++ {
		setting := string(r.Kvs[i].Key)
		setting = setting[prefixLen:]

		c.settings.Store(
			setting,
			string(r.Kvs[i].Value),
		)
	}

	return nil
}

// watch watches for the settings' changes in etcd and
// updates the in-memory settings cache.
func (c *Config) watch() {
	if err := c.load(); err != nil {
		c.logger.Log("msg", "dynconf failed to load settings", "path", c.path, "err", err)
	}

	updates := c.etcd.Watch(context.Background(), c.path, clientv3.WithPrefix())
	prefixLen := len(c.path)
	for u := range updates {
		for _, e := range u.Events {
			setting := string(e.Kv.Key)
			setting = setting[prefixLen:]

			switch e.Type {
			case clientv3.EventTypePut:
				c.settings.Store(setting, string(e.Kv.Value))
			case clientv3.EventTypeDelete:
				c.settings.Delete(setting)
			}
		}
	}
}

// Settings returns all the settings.
func (c *Config) Settings() map[string]string {
	ss := make(map[string]string)

	var k, v string
	c.settings.Range(func(key interface{}, value interface{}) bool {
		k, _ = key.(string)
		v, _ = value.(string)
		ss[k] = v
		return true
	})
	if len(ss) == 0 {
		return nil
	}

	return ss
}

// String returns the string value of the given setting,
// or defaultValue if it wasn't found.
func (c *Config) String(setting, defaultValue string) string {
	v, ok := c.settings.Load(setting)
	if !ok {
		c.logger.Log("msg", "dynconf setting not found", "path", c.path, "setting", setting, "err", "not found")
		return defaultValue
	}

	s, ok := v.(string)
	if !ok {
		c.logger.Log("msg", "dynconf invalid string value", "path", c.path, "setting", setting, "value", v)
		return defaultValue
	}

	return s
}

// Boolean returns the boolean value of the given setting,
// or defaultValue if it wasn't found or parsing failed.
func (c *Config) Boolean(setting string, defaultValue bool) bool {
	v, ok := c.settings.Load(setting)
	if !ok {
		c.logger.Log("msg", "dynconf setting not found", "path", c.path, "setting", setting, "err", "not found")
		return defaultValue
	}

	s, ok := v.(string)
	if !ok {
		c.logger.Log("msg", "dynconf invalid string value", "path", c.path, "setting", setting, "value", v)
		return defaultValue
	}

	b, err := strconv.ParseBool(s)
	if err != nil {
		c.logger.Log("msg", "dynconf invalid boolean setting", "path", c.path, "setting", setting, "value", s, "err", err)
		return defaultValue
	}

	return b
}

// Integer returns the integer value of the given setting,
// or defaultValue if it wasn't found or parsing failed.
func (c *Config) Integer(setting string, defaultValue int) int {
	v, ok := c.settings.Load(setting)
	if !ok {
		c.logger.Log("msg", "dynconf setting not found", "path", c.path, "setting", setting, "err", "not found")
		return defaultValue
	}

	s, ok := v.(string)
	if !ok {
		c.logger.Log("msg", "dynconf invalid string value", "path", c.path, "setting", setting, "value", v)
		return defaultValue
	}

	i, err := strconv.Atoi(s)
	if err != nil {
		c.logger.Log("msg", "dynconf invalid integer setting", "path", c.path, "setting", setting, "value", s, "err", err)
		return defaultValue
	}

	return i
}

// Float returns the float value of the given setting,
// or defaultValue if it wasn't found or parsing failed.
func (c *Config) Float(setting string, defaultValue float64) float64 {
	v, ok := c.settings.Load(setting)
	if !ok {
		c.logger.Log("msg", "dynconf setting not found", "path", c.path, "setting", setting, "err", "not found")
		return defaultValue
	}

	s, ok := v.(string)
	if !ok {
		c.logger.Log("msg", "dynconf invalid string value", "path", c.path, "setting", setting, "value", v)
		return defaultValue
	}

	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		c.logger.Log("msg", "dynconf invalid float setting", "path", c.path, "setting", setting, "value", s, "err", err)
		return defaultValue
	}

	return f
}

// Date returns the date value of the given setting,
// or defaultValue if it wasn't found or RFC3339 parsing failed.
func (c *Config) Date(setting string, defaultValue time.Time) time.Time {
	v, ok := c.settings.Load(setting)
	if !ok {
		c.logger.Log("msg", "dynconf setting not found", "path", c.path, "setting", setting, "err", "not found")
		return defaultValue
	}

	s, ok := v.(string)
	if !ok {
		c.logger.Log("msg", "dynconf invalid string value", "path", c.path, "setting", setting, "value", v)
		return defaultValue
	}

	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		c.logger.Log("msg", "dynconf invalid RFC3339 date setting", "path", c.path, "setting", setting, "value", s, "err", err)
		return defaultValue
	}

	return t
}
