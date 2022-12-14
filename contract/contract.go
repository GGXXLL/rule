package contract

import (
	"context"
	"time"
)

// Dispatcher is the event registry that is able to send payload to each listener.
type Dispatcher interface {
	Dispatch(ctx context.Context, topic interface{}, payload interface{}) error
	Subscribe(listener Listener)
}

// Listener is the handler for event.
type Listener interface {
	Listen() (topic interface{})
	Process(ctx context.Context, payload interface{}) error
}

// ConfigUnmarshaler is a minimum config interface that can be used to retrieve
// configuration from external system. If the configuration is hot reloaded,
// ConfigUnmarshaler should fetch the latest info.
type ConfigUnmarshaler interface {
	Unmarshal(path string, o interface{}) error
}

// ConfigAccessor builds upon the ConfigUnmarshaler and provides a richer set of
// API.
// Note: it is recommended to inject ConfigUnmarshaler as the dependency
// and call config.Upgrade to get the ConfigAccessor. The interface area of
// ConfigUnmarshaler is much smaller and thus much easier to customize.
type ConfigAccessor interface {
	ConfigUnmarshaler
	String(string) string
	Int(string) int
	Strings(string) []string
	Bool(string) bool
	Get(string) interface{}
	Float64(string) float64
	Duration(string) time.Duration
}
