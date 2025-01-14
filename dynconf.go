// Package dynconf provides a dynamic configuration backed by etcd.
// It can be used to access your project's settings without redeploying it every time a value changes.
package dynconf

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
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

// WithOnUpdate sets a function to be called when a setting is updated.
func WithOnUpdate(f func(settings map[string]string)) Option {
	return func(c *Config) {
		c.onUpdate = f
	}
}

// Config provides access to a project's settings stored in etcd.
type Config struct {
	// path (etcd key prefix) is the path to the project's config where settings are stored.
	path string
	// settings map holds the project's settings obtained from etcd.
	settings *sync.Map
	etcd     *clientv3.Client
	logger   log.Logger
	onUpdate func(settings map[string]string)
	ready    chan struct{}
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
		ready:    make(chan struct{}, 1),
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

// Ready waits until the Config is ready to use.
func (c *Config) Ready(ctx context.Context) error {
	select {
	case <-c.ready:
		close(c.ready)
		return nil
	case <-ctx.Done():
		return fmt.Errorf("dynconf not ready: %w", ctx.Err())
	}
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

	c.ready <- struct{}{}

	return nil
}

// watch watches for the settings' changes in etcd and
// updates the in-memory settings cache.
func (c *Config) watch() {
	if err := c.load(); err != nil {
		c.logger.Log("msg", "dynconf failed to load settings", "path", c.path, "err", err)
	}

	prefixLen := len(c.path)
	// As long as the context has not been canceled,
	// watch will retry on recoverable errors forever until reconnected.
	updates := c.etcd.Watch(context.Background(), c.path, clientv3.WithPrefix())
	for u := range updates {
		if err := u.Err(); err != nil {
			c.logger.Log("msg", "dynconf watch error", "path", c.path, "err", err)
		}

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

		if c.onUpdate != nil {
			c.onUpdate(c.Settings())
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

// StringRequired returns the string value of the given setting,
// or error if it wasn't found.
func (c *Config) StringRequired(setting string) (string, error) {
	v, ok := c.settings.Load(setting)
	if !ok {
		c.logger.Log("msg", "dynconf setting not found", "path", c.path, "setting", setting, "err", "not found")
		return "", fmt.Errorf("dynconf setting not found: %s", setting)
	}

	s, ok := v.(string)
	if !ok {
		c.logger.Log("msg", "dynconf invalid string value", "path", c.path, "setting", setting, "value", v)
		return "", fmt.Errorf("dynconf invalid string value: %s", setting)
	}

	return s, nil
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

// BooleanRequired returns the boolean value of the given setting,
// or error if it wasn't found or parsing failed.
func (c *Config) BooleanRequired(setting string) (bool, error) {
	v, ok := c.settings.Load(setting)
	if !ok {
		c.logger.Log("msg", "dynconf setting not found", "path", c.path, "setting", setting, "err", "not found")
		return false, fmt.Errorf("dynconf setting not found: %s", setting)
	}

	s, ok := v.(string)
	if !ok {
		c.logger.Log("msg", "dynconf invalid string value", "path", c.path, "setting", setting, "value", v)
		return false, fmt.Errorf("dynconf invalid string value: %s", setting)
	}

	b, err := strconv.ParseBool(s)
	if err != nil {
		c.logger.Log("msg", "dynconf invalid boolean setting", "path", c.path, "setting", setting, "value", s, "err", err)
		return false, fmt.Errorf("dynconf invalid boolean setting: %s", setting)
	}

	return b, nil
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

// IntegerRequired returns the integer value of the given setting,
// or error if it wasn't found or parsing failed.
func (c *Config) IntegerRequired(setting string) (int, error) {
	v, ok := c.settings.Load(setting)
	if !ok {
		c.logger.Log("msg", "dynconf setting not found", "path", c.path, "setting", setting, "err", "not found")
		return 0, fmt.Errorf("dynconf setting not found: %s", setting)
	}

	s, ok := v.(string)
	if !ok {
		c.logger.Log("msg", "dynconf invalid string value", "path", c.path, "setting", setting, "value", v)
		return 0, fmt.Errorf("dynconf invalid string value: %s", setting)
	}

	i, err := strconv.Atoi(s)
	if err != nil {
		c.logger.Log("msg", "dynconf invalid integer setting", "path", c.path, "setting", setting, "value", s, "err", err)
		return 0, fmt.Errorf("dynconf invalid integer setting: %s", setting)
	}

	return i, nil
}

// Int64 returns the int64 value of the given setting,
// or defaultValue if it wasn't found or parsing failed.
func (c *Config) Int64(setting string, defaultValue int64) int64 {
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

	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		c.logger.Log("msg", "dynconf invalid integer setting", "path", c.path, "setting", setting, "value", s, "err", err)
		return defaultValue
	}

	return i
}

// Int64Required returns the int64 value of the given setting,
// or error if it wasn't found or parsing failed.
func (c *Config) Int64Required(setting string) (int64, error) {
	v, ok := c.settings.Load(setting)
	if !ok {
		c.logger.Log("msg", "dynconf setting not found", "path", c.path, "setting", setting, "err", "not found")
		return 0, fmt.Errorf("dynconf setting not found: %s", setting)
	}

	s, ok := v.(string)
	if !ok {
		c.logger.Log("msg", "dynconf invalid string value", "path", c.path, "setting", setting, "value", v)
		return 0, fmt.Errorf("dynconf invalid string value: %s", setting)
	}

	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		c.logger.Log("msg", "dynconf invalid integer setting", "path", c.path, "setting", setting, "value", s, "err", err)
		return 0, fmt.Errorf("dynconf invalid integer setting: %s", setting)
	}

	return i, nil
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

// FloatRequired returns the float value of the given setting,
// or error if it wasn't found or parsing failed.
func (c *Config) FloatRequired(setting string) (float64, error) {
	v, ok := c.settings.Load(setting)
	if !ok {
		c.logger.Log("msg", "dynconf setting not found", "path", c.path, "setting", setting, "err", "not found")
		return 0, fmt.Errorf("dynconf setting not found: %s", setting)
	}

	s, ok := v.(string)
	if !ok {
		c.logger.Log("msg", "dynconf invalid string value", "path", c.path, "setting", setting, "value", v)
		return 0, fmt.Errorf("dynconf invalid string value: %s", setting)
	}

	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		c.logger.Log("msg", "dynconf invalid float setting", "path", c.path, "setting", setting, "value", s, "err", err)
		return 0, fmt.Errorf("dynconf invalid float setting: %s", setting)
	}

	return f, nil
}

// Date returns the date value of the given setting,
// or defaultValue if it wasn't found or RFC3339 parsing failed.
func (c *Config) Date(setting string, format string, defaultValue time.Time) time.Time {
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

	t, err := time.Parse(format, s)
	if err != nil {
		c.logger.Log("msg", "dynconf invalid RFC3339 date setting", "path", c.path, "setting", setting, "value", s, "err", err)
		return defaultValue
	}

	return t
}

// DateRequired returns the date value of the given setting,
// or error if it wasn't found or RFC3339 parsing failed.
func (c *Config) DateRequired(setting string, format string) (time.Time, error) {
	v, ok := c.settings.Load(setting)
	if !ok {
		c.logger.Log("msg", "dynconf setting not found", "path", c.path, "setting", setting, "err", "not found")
		return time.Time{}, fmt.Errorf("dynconf setting not found: %s", setting)
	}

	s, ok := v.(string)
	if !ok {
		c.logger.Log("msg", "dynconf invalid string value", "path", c.path, "setting", setting, "value", v)
		return time.Time{}, fmt.Errorf("dynconf invalid string value: %s", setting)
	}

	t, err := time.Parse(format, s)
	if err != nil {
		c.logger.Log("msg", "dynconf invalid RFC3339 date setting", "path", c.path, "setting", setting, "value", s, "err", err)
		return time.Time{}, fmt.Errorf("dynconf invalid RFC3339 date setting: %s", setting)
	}

	return t, nil
}

// Struct returns the struct value of the given setting,
func (c *Config) Struct(setting string, out interface{}) error {
	v, ok := c.settings.Load(setting)
	if !ok {
		c.logger.Log("msg", "dynconf setting not found", "path", c.path, "setting", setting, "err", "not found")
		return errors.New("setting not found")
	}

	s, ok := v.(string)
	if !ok {
		c.logger.Log("msg", "dynconf invalid string value", "path", c.path, "setting", setting, "value", v)
		return errors.New("invalid string value")
	}

	if unmarshaler, ok := out.(json.Unmarshaler); ok && unmarshaler != nil {
		return unmarshaler.UnmarshalJSON([]byte(s))
	}

	return json.Unmarshal([]byte(s), out)
}

// Duration returns the duration value of the given setting,
// or defaultValue if it wasn't found or parsing failed.
func (c *Config) Duration(setting string, defaultValue time.Duration) time.Duration {
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

	d, err := time.ParseDuration(s)
	if err != nil {
		c.logger.Log("msg", "dynconf invalid duration setting", "path", c.path, "setting", setting, "value", s, "err", err)
		return defaultValue
	}

	return d
}

// DurationRequired returns the duration value of the given setting,
// or error if it wasn't found or parsing failed.
func (c *Config) DurationRequired(setting string) (time.Duration, error) {
	v, ok := c.settings.Load(setting)
	if !ok {
		c.logger.Log("msg", "dynconf setting not found", "path", c.path, "setting", setting, "err", "not found")
		return 0, fmt.Errorf("dynconf setting not found: %s", setting)
	}

	s, ok := v.(string)
	if !ok {
		c.logger.Log("msg", "dynconf invalid string value", "path", c.path, "setting", setting, "value", v)
		return 0, fmt.Errorf("dynconf invalid string value: %s", setting)
	}

	d, err := time.ParseDuration(s)
	if err != nil {
		c.logger.Log("msg", "dynconf invalid duration setting", "path", c.path, "setting", setting, "value", s, "err", err)
		return 0, fmt.Errorf("dynconf invalid duration setting: %s", setting)
	}

	return d, nil
}

// StringArray returns the string array value of the given setting,
func (c *Config) StringArray(setting string, delimiter string) []string {
	v, ok := c.settings.Load(setting)
	if !ok {
		c.logger.Log("msg", "dynconf setting not found", "path", c.path, "setting", setting, "err", "not found")
		return nil
	}

	s, ok := v.(string)
	if !ok {
		c.logger.Log("msg", "dynconf invalid string value", "path", c.path, "setting", setting, "value", v)
		return nil
	}

	return strings.Split(s, delimiter)
}

// IntegerArray returns the integer array value of the given setting,
func (c *Config) IntegerArray(setting string, delimiter string) []int {
	v, ok := c.settings.Load(setting)
	if !ok {
		c.logger.Log("msg", "dynconf setting not found", "path", c.path, "setting", setting, "err", "not found")
		return nil
	}

	s, ok := v.(string)
	if !ok {
		c.logger.Log("msg", "dynconf invalid string value", "path", c.path, "setting", setting, "value", v)
		return nil
	}

	ss := strings.Split(s, delimiter)
	is := make([]int, len(ss))
	for i, s := range ss {
		is[i], _ = strconv.Atoi(s)
	}

	return is
}

// FloatArray returns the float array value of the given setting,
func (c *Config) FloatArray(setting string, delimiter string) []float64 {
	v, ok := c.settings.Load(setting)
	if !ok {
		c.logger.Log("msg", "dynconf setting not found", "path", c.path, "setting", setting, "err", "not found")
		return nil
	}

	s, ok := v.(string)
	if !ok {
		c.logger.Log("msg", "dynconf invalid string value", "path", c.path, "setting", setting, "value", v)
		return nil
	}

	ss := strings.Split(s, delimiter)
	fs := make([]float64, len(ss))
	for i, s := range ss {
		fs[i], _ = strconv.ParseFloat(s, 64)
	}

	return fs
}

// DateArray returns the date array value of the given setting,
func (c *Config) DateArray(setting string, format string, delimiter string) []time.Time {
	v, ok := c.settings.Load(setting)
	if !ok {
		c.logger.Log("msg", "dynconf setting not found", "path", c.path, "setting", setting, "err", "not found")
		return nil
	}

	s, ok := v.(string)
	if !ok {
		c.logger.Log("msg", "dynconf invalid string value", "path", c.path, "setting", setting, "value", v)
		return nil
	}

	ss := strings.Split(s, delimiter)
	ts := make([]time.Time, len(ss))
	for i, s := range ss {
		ts[i], _ = time.Parse(format, s)
	}

	return ts
}

// BooleanArray returns the boolean array value of the given setting,
func (c *Config) BooleanArray(setting string, delimiter string) []bool {
	v, ok := c.settings.Load(setting)
	if !ok {
		c.logger.Log("msg", "dynconf setting not found", "path", c.path, "setting", setting, "err", "not found")
		return nil
	}

	s, ok := v.(string)
	if !ok {
		c.logger.Log("msg", "dynconf invalid string value", "path", c.path, "setting", setting, "value", v)
		return nil
	}

	ss := strings.Split(s, delimiter)
	bs := make([]bool, len(ss))
	for i, s := range ss {
		bs[i], _ = strconv.ParseBool(s)
	}

	return bs
}
