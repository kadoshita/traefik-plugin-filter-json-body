package filterjsonbody

import (
	"context"
	"net/http"
)

// Config the plugin configuration.
type Config struct {
}

// CreateConfig creates the default plugin configuration.
func CreateConfig() *Config {
	return &Config{}
}

// FilterJsonBody a FilterJsonBody plugin.
type FilterJsonBody struct {
	next http.Handler
	name string
}

func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	return &FilterJsonBody{
		next: next,
		name: name,
	}, nil
}

func (a *FilterJsonBody) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	a.next.ServeHTTP(rw, req)
}
