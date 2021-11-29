package dynconf

import (
	"testing"

	"github.com/google/go-cmp/cmp"
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

	c, err := New("configs/curiosity")
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

func TestConfigBool(t *testing.T) {
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

	c, err := New("configs/curiosity")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := c.Close(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("no key", func(t *testing.T) {
		got := c.Bool("is_camera_enabled", defaultIsCameraEnabled)
		want := defaultIsCameraEnabled
		if want != got {
			t.Errorf("expected %t got %t", want, got)
		}
	})

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			c.settings.Store("is_camera_enabled", tc.in)
			got := c.Bool("is_camera_enabled", defaultIsCameraEnabled)
			if tc.want != got {
				t.Errorf("expected %t got %t", tc.want, got)
			}
		})
	}
}

func TestConfigInt(t *testing.T) {
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

	c, err := New("configs/curiosity")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := c.Close(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("no key", func(t *testing.T) {
		got := c.Int("velocity", defaultVelocity)
		want := defaultVelocity
		if want != got {
			t.Errorf("expected %d got %d", want, got)
		}
	})

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			c.settings.Store("velocity", tc.in)
			got := c.Int("velocity", defaultVelocity)
			if tc.want != got {
				t.Errorf("expected %d got %d", tc.want, got)
			}
		})
	}
}

func TestConfigFloat(t *testing.T) {
	const defaultTemperature = 0.001

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

	c, err := New("configs/curiosity")
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

	c, err := New("configs/curiosity")
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
