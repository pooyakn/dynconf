package dynconf

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/go-kit/log"
	"github.com/google/go-cmp/cmp"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func TestConfigString(t *testing.T) {
	const defaultName = "bob"

	tests := map[string]struct {
		in   interface{}
		want string
	}{
		"string": {
			in:   "alice",
			want: "alice",
		},
		"bytes": {
			in:   []byte("alice"),
			want: defaultName,
		},
		"nil": {
			in:   nil,
			want: defaultName,
		},
		"int": {
			in:   100,
			want: defaultName,
		},
		"float": {
			in:   0.001,
			want: defaultName,
		},
	}

	logger := log.NewJSONLogger(log.NewSyncWriter(os.Stderr))
	c, err := New("configs/curiosity", WithLogger(logger))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := c.Close(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("no key", func(t *testing.T) {
		got := c.String("name", defaultName)
		want := defaultName
		if want != got {
			t.Errorf("expected %q got %q", want, got)
		}
	})

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			c.settings.Store("name", tc.in)
			got := c.String("name", defaultName)
			if tc.want != got {
				t.Errorf("expected %q got %q", tc.want, got)
			}
		})
	}
}

func TestConfigBoolean(t *testing.T) {
	const defaultIsCameraEnabled = false

	tests := map[string]struct {
		in   interface{}
		want bool
	}{
		"string bool": {
			in:   "false",
			want: false,
		},
		"string int": {
			in:   "10",
			want: defaultIsCameraEnabled,
		},
		"string float": {
			in:   "0.001",
			want: defaultIsCameraEnabled,
		},
		"string name": {
			in:   "alice",
			want: defaultIsCameraEnabled,
		},
		"bytes": {
			in:   []byte("alice"),
			want: defaultIsCameraEnabled,
		},
		"nil": {
			in:   nil,
			want: defaultIsCameraEnabled,
		},
		"int": {
			in:   100,
			want: defaultIsCameraEnabled,
		},
		"float": {
			in:   0.001,
			want: defaultIsCameraEnabled,
		},
	}

	logger := log.NewJSONLogger(log.NewSyncWriter(os.Stderr))
	c, err := New("configs/curiosity", WithLogger(logger))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := c.Close(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("no key", func(t *testing.T) {
		got := c.Boolean("is_camera_enabled", defaultIsCameraEnabled)
		want := defaultIsCameraEnabled
		if want != got {
			t.Errorf("expected %t got %t", want, got)
		}
	})

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			c.settings.Store("is_camera_enabled", tc.in)
			got := c.Boolean("is_camera_enabled", defaultIsCameraEnabled)
			if tc.want != got {
				t.Errorf("expected %t got %t", tc.want, got)
			}
		})
	}
}

func TestConfigInteger(t *testing.T) {
	const defaultVelocity = 10

	tests := map[string]struct {
		in   interface{}
		want int
	}{
		"string int": {
			in:   "10",
			want: 10,
		},
		"string name": {
			in:   "alice",
			want: defaultVelocity,
		},
		"bytes": {
			in:   []byte("alice"),
			want: defaultVelocity,
		},
		"nil": {
			in:   nil,
			want: defaultVelocity,
		},
		"int": {
			in:   100,
			want: defaultVelocity,
		},
		"float": {
			in:   0.001,
			want: defaultVelocity,
		},
	}

	logger := log.NewJSONLogger(log.NewSyncWriter(os.Stderr))
	c, err := New("configs/curiosity", WithLogger(logger))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := c.Close(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("no key", func(t *testing.T) {
		got := c.Integer("velocity", defaultVelocity)
		want := defaultVelocity
		if want != got {
			t.Errorf("expected %d got %d", want, got)
		}
	})

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			c.settings.Store("velocity", tc.in)
			got := c.Integer("velocity", defaultVelocity)
			if tc.want != got {
				t.Errorf("expected %d got %d", tc.want, got)
			}
		})
	}
}

func TestConfigFloat(t *testing.T) {
	const defaultTemperature = 36.6

	tests := map[string]struct {
		in   interface{}
		want float64
	}{
		"string int": {
			in:   "10",
			want: 10,
		},
		"string float": {
			in:   "10.1",
			want: 10.1,
		},
		"string name": {
			in:   "alice",
			want: defaultTemperature,
		},
		"bytes": {
			in:   []byte("alice"),
			want: defaultTemperature,
		},
		"nil": {
			in:   nil,
			want: defaultTemperature,
		},
		"int": {
			in:   100,
			want: defaultTemperature,
		},
		"float": {
			in:   0.001,
			want: defaultTemperature,
		},
	}

	logger := log.NewJSONLogger(log.NewSyncWriter(os.Stderr))
	c, err := New("configs/curiosity", WithLogger(logger))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := c.Close(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("no key", func(t *testing.T) {
		got := c.Float("temperature", defaultTemperature)
		want := defaultTemperature
		if want != got {
			t.Errorf("expected %f got %f", want, got)
		}
	})

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			c.settings.Store("temperature", tc.in)
			got := c.Float("temperature", defaultTemperature)
			if tc.want != got {
				t.Errorf("expected %f got %f", tc.want, got)
			}
		})
	}
}

func TestConfigDate(t *testing.T) {
	defaultLaunchedDate, _ := time.Parse(time.RFC3339, "2021-11-30T20:14:05.134115+00:00")

	tests := map[string]struct {
		in   interface{}
		want time.Time
	}{
		"string date": {
			in:   "2021-11-30T21:14:05.134115+00:00",
			want: defaultLaunchedDate.Add(time.Hour),
		},
		"string int": {
			in:   "10",
			want: defaultLaunchedDate,
		},
		"string float": {
			in:   "10.1",
			want: defaultLaunchedDate,
		},
		"string name": {
			in:   "alice",
			want: defaultLaunchedDate,
		},
		"bytes": {
			in:   []byte("alice"),
			want: defaultLaunchedDate,
		},
		"nil": {
			in:   nil,
			want: defaultLaunchedDate,
		},
		"int": {
			in:   100,
			want: defaultLaunchedDate,
		},
		"float": {
			in:   0.001,
			want: defaultLaunchedDate,
		},
	}

	logger := log.NewJSONLogger(log.NewSyncWriter(os.Stderr))
	c, err := New("configs/curiosity", WithLogger(logger))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := c.Close(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("no key", func(t *testing.T) {
		got := c.Date("launched_at", defaultLaunchedDate)
		want := defaultLaunchedDate
		if !want.Equal(got) {
			t.Errorf("expected %s got %s", want, got)
		}
	})

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			c.settings.Store("launched_at", tc.in)
			got := c.Date("launched_at", defaultLaunchedDate)
			if !tc.want.Equal(got) {
				t.Errorf("expected %s got %s", tc.want, got)
			}
		})
	}
}

func TestConfigSettings(t *testing.T) {
	tests := map[string]struct {
		in   interface{}
		want map[string]string
	}{
		"string int": {
			in:   "10",
			want: map[string]string{"name": "10"},
		},
		"string float": {
			in:   "10.1",
			want: map[string]string{"name": "10.1"},
		},
		"string name": {
			in:   "alice",
			want: map[string]string{"name": "alice"},
		},
		"bytes": {
			in:   []byte("alice"),
			want: map[string]string{"name": ""},
		},
		"nil": {
			in:   nil,
			want: map[string]string{"name": ""},
		},
		"int": {
			in:   100,
			want: map[string]string{"name": ""},
		},
		"float": {
			in:   0.001,
			want: map[string]string{"name": ""},
		},
	}

	logger := log.NewJSONLogger(log.NewSyncWriter(os.Stderr))
	c, err := New("configs/curiosity", WithLogger(logger))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := c.Close(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("no keys", func(t *testing.T) {
		got := c.Settings()
		if got != nil {
			t.Errorf("expected nil got %v", got)
		}
	})

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			c.settings.Store("name", tc.in)
			got := c.Settings()
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func TestNew(t *testing.T) {
	etcd, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"127.0.0.1:2379"},
	})
	if err != nil {
		t.Fatal(err)
	}

	logger := log.NewJSONLogger(log.NewSyncWriter(os.Stderr))
	c, err := New("configs/curiosity", WithEtcdClient(etcd), WithLogger(logger))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := c.Close(); err != nil {
			t.Fatal(err)
		}
	})

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if r, err := etcd.Put(ctx, "configs/curiosity/velocity", "5"); err != nil {
		t.Fatalf("failed to put velocity=5 setting: %v %v", err, r)
	}
	// Wait for the watcher to see the changes in etcd.
	time.Sleep(time.Second)

	got := c.Integer("velocity", 10)
	want := 5
	if want != got {
		t.Errorf("expected velocity %d got %d", want, got)
	}
}
