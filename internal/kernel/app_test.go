package kernel

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Wi1low/go-web-template/internal/config"
)

// recorder is a scripted module that appends its lifecycle events to a shared
// log so ordering can be asserted.
type recorder struct {
	name    string
	enabled bool
	initErr error
	stopErr error
	log     *[]string
}

func (m *recorder) Name() string { return m.name }

func (m *recorder) Enabled(*config.Config) bool { return m.enabled }

func (m *recorder) Init(_ context.Context, a *App) error {
	*m.log = append(*m.log, "init:"+m.name)
	if m.name == "config" {
		a.Cfg = &config.Config{}
	}
	return m.initErr
}

func (m *recorder) Stop(context.Context) error {
	*m.log = append(*m.log, "stop:"+m.name)
	return m.stopErr
}

func newConfig(log *[]string) *recorder {
	return &recorder{name: "config", enabled: true, log: log}
}

func TestRunStopsInReverseAndHonoursSkipDisabled(t *testing.T) {
	var log []string
	app := New(
		newConfig(&log),
		&recorder{name: "a", enabled: true, log: &log},
		&recorder{name: "b", enabled: false, log: &log},
		&recorder{name: "c", enabled: true, log: &log},
	)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(20 * time.Millisecond)
		cancel()
	}()

	if err := app.Run(ctx); err != nil {
		t.Fatalf("Run() = %v", err)
	}
	want := []string{"init:config", "init:a", "init:c", "stop:c", "stop:a", "stop:config"}
	if len(log) != len(want) {
		t.Fatalf("log = %v, want %v", log, want)
	}
	for i := range want {
		if log[i] != want[i] {
			t.Fatalf("log = %v, want %v", log, want)
		}
	}
}

func TestInitFailureRollsBackStartedModules(t *testing.T) {
	var log []string
	boom := errors.New("boom")
	app := New(
		newConfig(&log),
		&recorder{name: "a", enabled: true, log: &log},
		&recorder{name: "bad", enabled: true, initErr: boom, log: &log},
	)

	err := app.Run(context.Background())
	if !errors.Is(err, boom) {
		t.Fatalf("Run() error = %v, want wrapped boom", err)
	}
	want := []string{"init:config", "init:a", "init:bad", "stop:a", "stop:config"}
	for i := range want {
		if i >= len(log) || log[i] != want[i] {
			t.Fatalf("log = %v, want %v", log, want)
		}
	}
}

func TestFailTriggersShutdown(t *testing.T) {
	var log []string
	app := New(newConfig(&log), &recorder{name: "a", enabled: true, log: &log})

	go func() {
		time.Sleep(20 * time.Millisecond)
		app.Fail(errors.New("listener died"))
	}()

	start := time.Now()
	err := app.Run(context.Background())
	if err == nil || err.Error() != "listener died" {
		t.Fatalf("Run() error = %v, want the reported failure", err)
	}
	if time.Since(start) > 2*time.Second {
		t.Fatal("Run did not react to Fail promptly")
	}
}

func TestStopErrorsAreJoined(t *testing.T) {
	var log []string
	nasty := errors.New("stop failed")
	app := New(newConfig(&log), &recorder{name: "a", enabled: true, stopErr: nasty, log: &log})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := app.Run(ctx)
	if !errors.Is(err, nasty) {
		t.Fatalf("Run() error = %v, want stop error surfaced", err)
	}
}

func TestContainerProvideRequire(t *testing.T) {
	var log []string
	app := New(newConfig(&log))

	type myService struct{ n int }
	Provide[myService](app, myService{n: 7})

	got, err := Require[myService](app)
	if err != nil || got.n != 7 {
		t.Fatalf("Require() = (%v, %v)", got, err)
	}
	if _, err := Require[*myService](app); err == nil {
		t.Fatal("Require of unregistered type must fail")
	}
	defer func() {
		if recover() == nil {
			t.Fatal("double Provide must panic")
		}
	}()
	Provide[myService](app, myService{n: 1})
}

func TestNewRejectsBadRegistration(t *testing.T) {
	var log []string
	defer func() {
		if recover() == nil {
			t.Fatal("New should panic when config is not first")
		}
	}()
	New(&recorder{name: "a", enabled: true, log: &log})
}
