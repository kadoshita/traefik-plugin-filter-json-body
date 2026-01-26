package traefik_plugin_filter_json_body

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/antchfx/jsonquery"
)

type Rule struct {
	Path               string `json:"path,omitempty"`               // only exact match
	Method             string `json:"method,omitempty"`             // only exact match
	BodyPath           string `json:"bodyPath,omitempty"`           // JSON XPath
	BodyValueCondition string `json:"bodyValueCondition,omitempty"` // regex match
}

// Config the plugin configuration.
type Config struct {
	Rules []Rule `json:"rules,omitempty"`
}

// CreateConfig creates the default plugin configuration.
func CreateConfig() *Config {
	return &Config{
		Rules: []Rule{},
	}
}

type compiledRule struct {
	Path               string
	Method             string
	BodyPath           string
	BodyValueCondition *regexp.Regexp
}

// FilterJsonBody a FilterJsonBody plugin.
type FilterJsonBody struct {
	next                   http.Handler
	name                   string
	rulesAfterRegexCompile []compiledRule
}

const MAX_BODY_SIZE = 10 * 1024 * 1024 // 10 MB

var contentTypeRegex = regexp.MustCompile(`^application/(json|[a-z0-9.-]+\+json)`)

func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	if len(config.Rules) < 1 {
		return nil, fmt.Errorf("at least one rule must be defined")
	}
	rulesAfterRegexCompile := make([]compiledRule, len(config.Rules))
	for i, rule := range config.Rules {
		if rule.Path == "" {
			continue
		}

		if rule.Method == "" {
			continue
		}

		if rule.BodyPath == "" {
			continue
		}

		if rule.BodyValueCondition == "" {
			continue
		}

		compiledBodyValueConditionRegex, err := regexp.Compile(rule.BodyValueCondition)
		if err != nil {
			return nil, fmt.Errorf("failed to compile BodyValueCondition regex for rule %d: %v", i, err)
		}
		rulesAfterRegexCompile[i] = compiledRule{
			Path:               rule.Path,
			Method:             rule.Method,
			BodyPath:           rule.BodyPath,
			BodyValueCondition: compiledBodyValueConditionRegex,
		}
	}

	return &FilterJsonBody{
		next:                   next,
		name:                   name,
		rulesAfterRegexCompile: rulesAfterRegexCompile,
	}, nil
}

func (a *FilterJsonBody) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	contentType := strings.ToLower(req.Header.Get("Content-Type"))
	if contentType == "" || !contentTypeRegex.MatchString(contentType) {
		a.next.ServeHTTP(rw, req)
		return
	}

	var matchedRules []*compiledRule

	for i := range a.rulesAfterRegexCompile {
		rule := &a.rulesAfterRegexCompile[i]

		if req.URL.Path != rule.Path {
			continue
		}
		if req.Method != rule.Method {
			continue
		}

		matchedRules = append(matchedRules, rule)
	}

	if len(matchedRules) == 0 {
		a.next.ServeHTTP(rw, req)
		return
	}

	limitedReader := io.LimitReader(req.Body, MAX_BODY_SIZE)
	bodyBytes, err := io.ReadAll(limitedReader)
	req.Body.Close()

	if len(bodyBytes) >= MAX_BODY_SIZE {
		req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		a.next.ServeHTTP(rw, req)
		return
	}

	if err != nil {
		req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		a.next.ServeHTTP(rw, req)
		return
	}

	doc, err := jsonquery.Parse(bytes.NewReader(bodyBytes))
	if err != nil {
		req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		a.next.ServeHTTP(rw, req)
		return
	}

	for _, matchedRule := range matchedRules {
		target := jsonquery.FindOne(doc, matchedRule.BodyPath)
		if target != nil {
			value := fmt.Sprintf("%v", target.Value())
			if matchedRule.BodyValueCondition.MatchString(value) {
				fmt.Printf("Request blocked path=%s method=%s bodyPath=%s bodyValue=%s\n", req.URL.Path, req.Method, matchedRule.BodyPath, value)
				rw.WriteHeader(http.StatusForbidden)
				rw.Write([]byte("Forbidden"))
				return
			}
		}
	}

	req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	a.next.ServeHTTP(rw, req)
}
