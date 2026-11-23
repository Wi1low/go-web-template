package kernel

import (
	"fmt"
	"reflect"
)

// Provide publishes v as the sole service of type T. When publishing a
// concrete implementation under an interface type, instantiate the type
// parameter explicitly:
//
//	Provide[auth.TokenParser](app, parser)
func Provide[T any](a *App, v T) {
	t := reflect.TypeOf((*T)(nil)).Elem()
	a.mu.Lock()
	defer a.mu.Unlock()
	if _, dup := a.services[t]; dup {
		panic(fmt.Sprintf("kernel: service %s already provided", t))
	}
	a.services[t] = v
}

// Require returns the service of type T published by an earlier Init.
func Require[T any](a *App) (T, error) {
	var zero T
	t := reflect.TypeOf((*T)(nil)).Elem()
	a.mu.RLock()
	defer a.mu.RUnlock()
	v, ok := a.services[t]
	if !ok {
		return zero, fmt.Errorf("kernel: service %s not provided; register the module that owns it before the one that requires it", t)
	}
	svc, ok := v.(T)
	if !ok {
		return zero, fmt.Errorf("kernel: service %s has unexpected dynamic type %T", t, v)
	}
	return svc, nil
}

// MustRequire is Require for modules that consider the dependency fatal.
func MustRequire[T any](a *App) T {
	v, err := Require[T](a)
	if err != nil {
		panic(err.Error())
	}
	return v
}
