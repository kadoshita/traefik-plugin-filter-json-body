package filterjsonbody_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	filterjsonbody "github.com/kadoshita/traefik-plugin-filter-json-body"
)

func TestFilterJsonBody(t *testing.T) {
	nextHandler := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)
		rw.Write([]byte("Next handler called"))
	})

	ctx := context.Background()
	config := &filterjsonbody.Config{
		Rules: []filterjsonbody.Rule{
			{
				Path:               "/api/test",
				Method:             "POST",
				BodyPath:           "key",
				BodyValueCondition: "^value$",
			},
		},
	}
	pluginHandler, err := filterjsonbody.New(ctx, nextHandler, config, "test-plugin")
	if err != nil {
		t.Fatalf("Error creating plugin handler: %v", err)
	}

	test_cases := []struct {
		title                 string
		method                string
		path                  string
		contentType           string
		body                  string
		expectStatusCode      int
		expectCallNextHandler bool
	}{
		{
			title:                 "POST json",
			method:                "POST",
			path:                  "/api/test",
			contentType:           "application/json",
			body:                  `{"key":"value"}`,
			expectStatusCode:      http.StatusForbidden,
			expectCallNextHandler: false,
		}, {
			title:                 "POST activity json",
			method:                "POST",
			path:                  "/api/test",
			contentType:           "application/activity+json",
			body:                  `{"key":"value"}`,
			expectStatusCode:      http.StatusForbidden,
			expectCallNextHandler: false,
		}, {
			title:                 "method not matching",
			method:                "GET",
			path:                  "/api/test",
			contentType:           "application/json",
			body:                  `{"key":"value"}`,
			expectStatusCode:      http.StatusOK,
			expectCallNextHandler: true,
		}, {
			title:                 "path not matching",
			method:                "POST",
			path:                  "/api/other",
			contentType:           "application/json",
			body:                  `{"key":"value"}`,
			expectStatusCode:      http.StatusOK,
			expectCallNextHandler: true,
		}, {
			title:                 "content type not matching",
			method:                "POST",
			path:                  "/api/test",
			contentType:           "text/plain",
			body:                  `{"key":"value"}`,
			expectStatusCode:      http.StatusOK,
			expectCallNextHandler: true,
		}, {
			title:                 "body value not matching",
			method:                "POST",
			path:                  "/api/test",
			contentType:           "application/json",
			body:                  `{"key":"value2"}`,
			expectStatusCode:      http.StatusOK,
			expectCallNextHandler: true,
		}, {
			title:                 "invalid json body",
			method:                "POST",
			path:                  "/api/test",
			contentType:           "application/json",
			body:                  `invalid json`,
			expectStatusCode:      http.StatusOK,
			expectCallNextHandler: true,
		}, {
			title:                 "empty body",
			method:                "POST",
			path:                  "/api/test",
			contentType:           "application/json",
			body:                  ``,
			expectStatusCode:      http.StatusOK,
			expectCallNextHandler: true,
		}, {
			title:                 "non matching body path",
			method:                "POST",
			path:                  "/api/test",
			contentType:           "application/json",
			body:                  `{"otherKey":"value"}`,
			expectStatusCode:      http.StatusOK,
			expectCallNextHandler: true,
		},
	}

	for _, test_case := range test_cases {
		t.Run(test_case.title, func(t *testing.T) {
			req := httptest.NewRequestWithContext(ctx, test_case.method, "http://example.com"+test_case.path, strings.NewReader(test_case.body))
			req.Header.Set("Content-Type", test_case.contentType)
			rr := httptest.NewRecorder()

			pluginHandler.ServeHTTP(rr, req)

			if status := rr.Code; status != test_case.expectStatusCode {
				t.Errorf("Handler returned wrong status code: got %v want %v", status, test_case.expectStatusCode)
			}

			if test_case.expectCallNextHandler {
				expectedBody := "Next handler called"
				if rr.Body.String() != expectedBody {
					t.Errorf("Handler returned unexpected body: got %v want %v", rr.Body.String(), expectedBody)
				}
			}
		})
	}
}

func TestNew(t *testing.T) {
	ctx := context.Background()
	nextHandler := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {})

	test_cases := []struct {
		title       string
		config      *filterjsonbody.Config
		expectError bool
	}{
		{
			title: "valid config",
			config: &filterjsonbody.Config{
				Rules: []filterjsonbody.Rule{
					{
						Path:               "/api/test",
						Method:             "POST",
						BodyPath:           "key",
						BodyValueCondition: "^value$",
					},
				},
			},
			expectError: false,
		}, {
			title: "no rules",
			config: &filterjsonbody.Config{
				Rules: []filterjsonbody.Rule{},
			},
			expectError: true,
		}, {
			title: "invalid regex",
			config: &filterjsonbody.Config{
				Rules: []filterjsonbody.Rule{
					{
						Path:               "/api/test",
						Method:             "POST",
						BodyPath:           "key",
						BodyValueCondition: "[invalid",
					},
				},
			},
			expectError: true,
		}, {
			title: "empty path ignore",
			config: &filterjsonbody.Config{
				Rules: []filterjsonbody.Rule{
					{
						Path:               "",
						Method:             "POST",
						BodyPath:           "key",
						BodyValueCondition: "^value$",
					},
				},
			},
			expectError: false,
		}, {
			title: "empty method ignore",
			config: &filterjsonbody.Config{
				Rules: []filterjsonbody.Rule{
					{
						Path:               "/api/test",
						Method:             "",
						BodyPath:           "key",
						BodyValueCondition: "^value$",
					},
				},
			},
			expectError: false,
		}, {
			title: "empty body path ignore",
			config: &filterjsonbody.Config{
				Rules: []filterjsonbody.Rule{
					{
						Path:               "/api/test",
						Method:             "POST",
						BodyPath:           "",
						BodyValueCondition: "^value$",
					},
				},
			},
			expectError: false,
		}, {
			title: "empty body value condition ignore",
			config: &filterjsonbody.Config{
				Rules: []filterjsonbody.Rule{
					{
						Path:               "/api/test",
						Method:             "POST",
						BodyPath:           "key",
						BodyValueCondition: "",
					},
				},
			},
			expectError: false,
		},
	}

	for _, test_case := range test_cases {
		t.Run(test_case.title, func(t *testing.T) {
			_, err := filterjsonbody.New(ctx, nextHandler, test_case.config, "test-plugin")
			if test_case.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !test_case.expectError && err != nil {
				t.Errorf("Did not expect error but got: %v", err)
			}
		})
	}
}
