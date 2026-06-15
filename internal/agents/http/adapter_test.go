package http

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ropeixoto/harnessx/internal/agents"
	"github.com/ropeixoto/harnessx/internal/agents/yaml"
)

func TestAdapter_RoundTrip(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if r.Header.Get("x-api-key") != "test-key" {
			t.Errorf("missing auth header, got %v", r.Header)
		}
		var payload map[string]any
		_ = json.Unmarshal(body, &payload)
		if payload["model"] != "default-model" {
			t.Errorf("missing model in request: %v", payload)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"content": []any{map[string]any{"text": "hello world"}},
			"usage":   map[string]any{"input_tokens": 12, "output_tokens": 7},
		})
	}))
	defer srv.Close()

	t.Setenv("HARNESS_SECRET_PROVIDER_KEY", "test-key")

	var spec yaml.Spec
	spec.ID = "provider-api"
	spec.Name = "Provider"
	spec.Type = "api"
	spec.Capabilities.Models = map[string]string{"default": "default-model"}
	spec.API.Endpoint = srv.URL
	spec.API.Method = "POST"
	spec.API.Auth.Header = "x-api-key"
	spec.API.Auth.SecretRef = "secret://provider_key"
	spec.API.RequestTemplate = `{"model":"{{model}}","prompt":"{{prompt}}"}`
	spec.API.Response.FinalMessage = "$.content.0.text"
	spec.API.Response.Usage = "$.usage"

	a := New(spec)
	res := a.Run(context.Background(), agents.AgentRequest{Prompt: "ping"})
	if res.Err != nil {
		t.Fatalf("unexpected err: %v", res.Err)
	}
	if res.Output.FinalMessage != "hello world" {
		t.Fatalf("final message: %q", res.Output.FinalMessage)
	}
	if res.Usage.InputTokens != 12 || res.Usage.OutputTokens != 7 {
		t.Fatalf("usage: %+v", res.Usage)
	}
}

func TestAdapter_AuthFailureClassified(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"invalid_api_key"}`))
	}))
	defer srv.Close()
	t.Setenv("HARNESS_SECRET_PROVIDER_KEY", "bad")

	var spec yaml.Spec
	spec.ID = "x"
	spec.Name = "x"
	spec.Type = "api"
	spec.API.Endpoint = srv.URL
	spec.API.Auth.Header = "x-api-key"
	spec.API.Auth.SecretRef = "secret://provider_key"
	spec.API.RequestTemplate = `{}`

	a := New(spec)
	res := a.Run(context.Background(), agents.AgentRequest{})
	if res.Failure != agents.FailureAuth {
		t.Fatalf("expected FailureAuth, got %v", res.Failure)
	}
}
