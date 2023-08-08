package dynconf

import (
	"context"
	"os"
	"reflect"
	"strconv"
	"strings"
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
	c, err := New("/configs/curiosity/", WithLogger(logger))
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

func TestConfigStringRequired(t *testing.T) {
	tests := map[string]struct {
		in      interface{}
		want    string
		wantErr bool
	}{
		"string": {
			in:   "alice",
			want: "alice",
		},
		"bytes": {
			in:      []byte("alice"),
			wantErr: true,
		},
		"nil": {
			in:      nil,
			wantErr: true,
		},
		"int": {
			in:      100,
			wantErr: true,
		},
		"float": {
			in:      0.001,
			wantErr: true,
		},
	}

	logger := log.NewJSONLogger(log.NewSyncWriter(os.Stderr))
	c, err := New("/configs/curiosity/", WithLogger(logger))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := c.Close(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("no key", func(t *testing.T) {
		_, err := c.StringRequired("name")
		if err == nil {
			t.Errorf("expected error")
		}
	})

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			c.settings.Store("name", tc.in)
			got, err := c.StringRequired("name")
			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error")
				}
				return
			}

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
	c, err := New("/configs/curiosity/", WithLogger(logger))
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
	c, err := New("/configs/curiosity/", WithLogger(logger))
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

func TestConfigInt64(t *testing.T) {
	const defaultVelocity int64 = 10

	tests := map[string]struct {
		in   interface{}
		want int64
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
	c, err := New("/configs/curiosity/", WithLogger(logger))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := c.Close(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("no key", func(t *testing.T) {
		got := c.Int64("velocity", defaultVelocity)
		want := defaultVelocity
		if want != got {
			t.Errorf("expected %d got %d", want, got)
		}
	})

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			c.settings.Store("velocity", tc.in)
			got := c.Int64("velocity", defaultVelocity)
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
	c, err := New("/configs/curiosity/", WithLogger(logger))
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
	c, err := New("/configs/curiosity/", WithLogger(logger))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := c.Close(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("no key", func(t *testing.T) {
		got := c.Date("launched_at", time.RFC3339, defaultLaunchedDate)
		want := defaultLaunchedDate
		if !want.Equal(got) {
			t.Errorf("expected %s got %s", want, got)
		}
	})

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			c.settings.Store("launched_at", tc.in)
			got := c.Date("launched_at", time.RFC3339, defaultLaunchedDate)
			if !tc.want.Equal(got) {
				t.Errorf("expected %s got %s", tc.want, got)
			}
		})
	}
}

func TestConfigStruct(t *testing.T) {
	type config struct {
		Name   string  `json:"name"`
		Age    int     `json:"age"`
		Weight float64 `json:"weight"`
	}

	tests := map[string]struct {
		in   interface{}
		want config
	}{
		"string int": {
			in: "{\"age\":10}",
			want: config{
				Age: 10,
			},
		},
		"string float": {
			in: "{\"weight\":10.1}",
			want: config{
				Weight: 10.1,
			},
		},
		"string name": {
			in:   "{\"name\":\"alice\"}",
			want: config{Name: "alice"},
		},
	}

	logger := log.NewJSONLogger(log.NewSyncWriter(os.Stderr))
	c, err := New("/configs/curiosity/", WithLogger(logger))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := c.Close(); err != nil {
			t.Fatal(err)
		}
	})

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			c.settings.Store("config", tc.in)
			var got config
			if err := c.Struct("config", &got); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(tc.want, got) {
				t.Errorf("expected %v got %v", tc.want, got)
			}
		})
	}
}

type configWithUnmarshaler struct {
	Name   string
	Age    int
	Weight float64
}

func (c *configWithUnmarshaler) UnmarshalJSON(data []byte) error {
	split := strings.Split(string(data), ",")
	conf := configWithUnmarshaler{}
	for _, d := range split {
		parts := strings.Split(d, ":")
		switch strings.Trim(parts[0], "\"") {
		case "name":
			conf.Name = strings.Trim(parts[1], "\"")
		case "age":
			age, err := strconv.Atoi(strings.Trim(parts[1], "\""))
			if err != nil {
				return err
			}
			conf.Age = age
		case "weight":
			weight, err := strconv.ParseFloat(strings.Trim(parts[1], "\""), 64)
			if err != nil {
				return err
			}
			conf.Weight = weight
		}
	}

	*c = conf
	return nil
}

func TestConfigStructCustomUnmarshaler(t *testing.T) {
	tests := map[string]struct {
		in   interface{}
		want configWithUnmarshaler
	}{
		"string int": {
			in: "age:10",
			want: configWithUnmarshaler{
				Age: 10,
			},
		},
		"string float": {
			in: "weight:10.1",
			want: configWithUnmarshaler{
				Weight: 10.1,
			},
		},
		"string name": {
			in:   "name:alice",
			want: configWithUnmarshaler{Name: "alice"},
		},
		"multi key": {
			in: "name:alice,age:10,weight:10.1",
			want: configWithUnmarshaler{
				Name:   "alice",
				Age:    10,
				Weight: 10.1,
			},
		},
	}

	logger := log.NewJSONLogger(log.NewSyncWriter(os.Stderr))
	c, err := New("/configs/curiosity/", WithLogger(logger))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := c.Close(); err != nil {
			t.Fatal(err)
		}
	})

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			c.settings.Store("config", tc.in)
			var got configWithUnmarshaler
			if err := c.Struct("config", &got); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(tc.want, got) {
				t.Errorf("expected %v got %v", tc.want, got)
			}
		})
	}
}

func TestConfigDuration(t *testing.T) {
	tests := map[string]struct {
		in   interface{}
		want time.Duration
	}{
		"duration second": {
			in:   "10s",
			want: 10 * time.Second,
		},
		"duration minute": {
			in:   "10m",
			want: 10 * time.Minute,
		},
		"duration hour": {
			in:   "10h",
			want: 10 * time.Hour,
		},
		"no duration": {
			in:   "10",
			want: 5 * time.Second,
		},
	}

	logger := log.NewJSONLogger(log.NewSyncWriter(os.Stderr))
	c, err := New("/configs/curiosity/", WithLogger(logger))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := c.Close(); err != nil {
			t.Fatal(err)
		}
	})

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			c.settings.Store("duration", tc.in)
			got := c.Duration("duration", 5*time.Second)
			if tc.want != got {
				t.Errorf("expected %s got %s", tc.want, got)
			}
		})
	}
}

func TestConfigStringArray(t *testing.T) {
	tests := map[string]struct {
		in   interface{}
		del  string
		want []string
	}{
		"string array": {
			in:   "alice,bob",
			del:  ",",
			want: []string{"alice", "bob"},
		},
		"string array with different separator": {
			in:   "alice|bob",
			del:  "|",
			want: []string{"alice", "bob"},
		},
	}

	logger := log.NewJSONLogger(log.NewSyncWriter(os.Stderr))
	c, err := New("/configs/curiosity/", WithLogger(logger))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := c.Close(); err != nil {
			t.Fatal(err)
		}
	})

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			c.settings.Store("names", tc.in)
			got := c.StringArray("names", tc.del)
			if !reflect.DeepEqual(tc.want, got) {
				t.Errorf("expected %v got %v", tc.want, got)
			}
		})
	}
}

func TestConfigIntegerArray(t *testing.T) {
	tests := map[string]struct {
		in   interface{}
		del  string
		want []int
	}{
		"string array": {
			in:   "10,20",
			del:  ",",
			want: []int{10, 20},
		},
		"string array with different separator": {
			in:   "10|20",
			del:  "|",
			want: []int{10, 20},
		},
	}

	logger := log.NewJSONLogger(log.NewSyncWriter(os.Stderr))
	c, err := New("/configs/curiosity/", WithLogger(logger))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := c.Close(); err != nil {
			t.Fatal(err)
		}
	})

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			c.settings.Store("numbers", tc.in)
			got := c.IntegerArray("numbers", tc.del)
			if !reflect.DeepEqual(tc.want, got) {
				t.Errorf("expected %v got %v", tc.want, got)
			}
		})
	}
}

func TestConfigFloatArray(t *testing.T) {
	tests := map[string]struct {
		in   interface{}
		del  string
		want []float64
	}{
		"string array": {
			in:   "10.1,20.2",
			del:  ",",
			want: []float64{10.1, 20.2},
		},
		"string array with different separator": {
			in:   "10.1|20.2",
			del:  "|",
			want: []float64{10.1, 20.2},
		},
	}

	logger := log.NewJSONLogger(log.NewSyncWriter(os.Stderr))
	c, err := New("/configs/curiosity/", WithLogger(logger))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := c.Close(); err != nil {
			t.Fatal(err)
		}
	})

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			c.settings.Store("numbers", tc.in)
			got := c.FloatArray("numbers", tc.del)
			if !reflect.DeepEqual(tc.want, got) {
				t.Errorf("expected %v got %v", tc.want, got)
			}
		})
	}
}

func TestConfigDateArray(t *testing.T) {
	tests := map[string]struct {
		in     interface{}
		del    string
		format string
		want   []time.Time
	}{
		"string array": {
			in:     "2020-01-01,2020-02-02",
			del:    ",",
			format: "2006-01-02",
			want:   []time.Time{time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2020, 2, 2, 0, 0, 0, 0, time.UTC)},
		},
		"string array with different separator": {
			in:     "2020-01-01|2020-02-02",
			del:    "|",
			format: "2006-01-02",
			want:   []time.Time{time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2020, 2, 2, 0, 0, 0, 0, time.UTC)},
		},
	}

	logger := log.NewJSONLogger(log.NewSyncWriter(os.Stderr))
	c, err := New("/configs/curiosity/", WithLogger(logger))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := c.Close(); err != nil {
			t.Fatal(err)
		}
	})

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			c.settings.Store("dates", tc.in)
			got := c.DateArray("dates", tc.format, tc.del)
			if !reflect.DeepEqual(tc.want, got) {
				t.Errorf("expected %v got %v", tc.want, got)
			}
		})
	}
}

func TestConfigConfigSettings(t *testing.T) {
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
	c, err := New("/configs/curiosity/", WithLogger(logger))
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

func TestConfigBooleanArray(t *testing.T) {
	tests := map[string]struct {
		in   interface{}
		del  string
		want []bool
	}{
		"string array": {
			in:  "true,false",
			del: ",",
			want: []bool{
				true,
				false,
			},
		},
		"string array with different separator": {
			in:  "true|false",
			del: "|",
			want: []bool{
				true,
				false,
			},
		},
	}

	logger := log.NewJSONLogger(log.NewSyncWriter(os.Stderr))
	c, err := New("/configs/curiosity/", WithLogger(logger))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := c.Close(); err != nil {
			t.Fatal(err)
		}
	})

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			c.settings.Store("bools", tc.in)
			got := c.BooleanArray("bools", tc.del)
			if !reflect.DeepEqual(tc.want, got) {
				t.Errorf("expected %v got %v", tc.want, got)
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
	c, err := New("/configs/curiosity/", WithEtcdClient(etcd), WithLogger(logger))
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
	if r, err := etcd.Put(ctx, "/configs/curiosity/velocity", "5"); err != nil {
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

func TestOnUpdate(t *testing.T) {
	etcd, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"127.0.0.1:2379"},
	})
	if err != nil {
		t.Fatal(err)
	}

	received := "0"
	onUpdate := func(s map[string]string) {
		t.Logf("updated: %v", s)
		received = s["velocity"]
	}

	logger := log.NewJSONLogger(log.NewSyncWriter(os.Stderr))
	c, err := New("/configs/curiosity/", WithEtcdClient(etcd), WithLogger(logger), WithOnUpdate(onUpdate))
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
	if r, err := etcd.Put(ctx, "/configs/curiosity/velocity", "5"); err != nil {
		t.Fatalf("failed to put velocity=5 setting: %v %v", err, r)
	}
	// Wait for the watcher to see the changes in etcd.
	time.Sleep(time.Second)

	got := c.Integer("velocity", 10)
	want := 5
	if want != got {
		t.Errorf("expected velocity %d got %d", want, got)
	}

	if received != "5" {
		t.Errorf("expected received %s got %s", "5", received)
	}
}

func TestReady(t *testing.T) {
	etcd, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"127.0.0.1:2379"},
	})
	if err != nil {
		t.Fatal(err)
	}

	logger := log.NewJSONLogger(log.NewSyncWriter(os.Stderr))
	c, err := New("/configs/curiosity/", WithEtcdClient(etcd), WithLogger(logger))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := c.Close(); err != nil {
			t.Fatal(err)
		}
	})

	ctx, cancelCtx := context.WithTimeout(context.Background(), time.Second)
	defer cancelCtx()
	if r, err := etcd.Put(ctx, "/configs/curiosity/velocity", "5"); err != nil {
		t.Errorf("failed to put velocity=5 setting: %v %v", err, r)
		return
	}
	// Wait for the watcher to see the changes in etcd.

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err = c.Ready(ctx)
	if err != nil {
		t.Fatal(err)
	}

	got := c.Integer("velocity", 10)
	want := 5
	if want != got {
		t.Errorf("expected velocity %d got %d", want, got)
	}
}
