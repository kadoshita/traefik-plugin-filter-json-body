package filterjsonbody_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	filterjsonbody "github.com/kadoshita/traefik-plugin-filter-json-body"
)

func TestFilterJsonBody(t *testing.T) {
	nextHandler := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)
		rw.Write([]byte("Next handler called"))
	})

	ctx := context.Background()
	config := filterjsonbody.CreateConfig()
	pluginHandler, err := filterjsonbody.New(ctx, nextHandler, config, "test-plugin")
	if err != nil {
		t.Fatalf("Error creating plugin handler: %v", err)
	}

	req := httptest.NewRequestWithContext(ctx, "GET", "http://example.com/foo", nil)
	rr := httptest.NewRecorder()

	pluginHandler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	expectedBody := "Next handler called"
	if rr.Body.String() != expectedBody {
		t.Errorf("Handler returned unexpected body: got %v want %v", rr.Body.String(), expectedBody)
	}
}
