package executor

import (
	"net/http"
	"strings"
	"testing"
)

func TestMasker_TypedPlaceholders(t *testing.T) {
	m := NewCredentialMasker()
	cases := []struct {
		header string
		value  string
		mask   string
	}{
		{"Authorization", "Bearer xyz", "[masked:bearer_token]"},
		{"X-API-Key", "abc", "[masked:api_key]"},
		{"X-Api-Key", "abc", "[masked:api_key]"},
		{"Api-Key", "abc", "[masked:api_key]"},
		{"Cookie", "session=abc", "[masked:cookie]"},
	}
	for _, c := range cases {
		h := http.Header{}
		h.Set(c.header, c.value)
		out := m.MaskHeaders(h)
		got := out.Get(c.header)
		if got != c.mask {
			t.Fatalf("header %q: expected %q, got %q", c.header, c.mask, got)
		}
		if strings.Contains(got, c.value) {
			t.Fatalf("header %q masked output leaked value: %q", c.header, got)
		}
	}
}

func TestMasker_NonCredentialHeaderUnchanged(t *testing.T) {
	m := NewCredentialMasker()
	h := http.Header{}
	h.Set("Content-Type", "application/json")
	h.Set("Accept", "*/*")
	out := m.MaskHeaders(h)
	if out.Get("Content-Type") != "application/json" {
		t.Fatalf("Content-Type changed: %s", out.Get("Content-Type"))
	}
	if out.Get("Accept") != "*/*" {
		t.Fatalf("Accept changed: %s", out.Get("Accept"))
	}
}

func TestMasker_InputHeaderNotMutated(t *testing.T) {
	m := NewCredentialMasker()
	h := http.Header{}
	h.Set("Authorization", "Bearer original")
	_ = m.MaskHeaders(h)
	if h.Get("Authorization") != "Bearer original" {
		t.Fatalf("input header was mutated: %s", h.Get("Authorization"))
	}
}

func TestMaskResponseHeaders_CredentialsRedacted(t *testing.T) {
	m := NewCredentialMasker()
	headers := map[string]string{
		"authorization":   "Bearer secret-token",
		"set-cookie":      "session=abc123",
		"content-type":    "application/json",
		"x-service-token": "super-secret",
	}
	out := m.MaskResponseHeaders(headers)

	if out["authorization"] != "[masked:bearer_token]" {
		t.Fatalf("authorization not masked: %q", out["authorization"])
	}
	if out["set-cookie"] != "[masked:cookie]" {
		t.Fatalf("set-cookie not masked: %q", out["set-cookie"])
	}
	if out["x-service-token"] != "[masked:credential]" {
		t.Fatalf("x-service-token not masked: %q", out["x-service-token"])
	}
	if out["content-type"] != "application/json" {
		t.Fatalf("content-type should be unchanged: %q", out["content-type"])
	}
	// original map must not be mutated
	if headers["authorization"] != "Bearer secret-token" {
		t.Fatal("input map was mutated")
	}
}

func TestMaskURL_CredentialParamsTyped(t *testing.T) {
	m := NewCredentialMasker()
	cases := []struct {
		name     string
		in       string
		notHave  string
		mustHave []string
	}{
		{"token", "http://x?token=abc", "abc", []string{"[masked:token]"}},
		{"key", "http://x?key=abc", "abc", []string{"[masked:key]"}},
		{"secret", "http://x?secret=abc", "abc", []string{"[masked:secret]"}},
		{"api_key", "http://x?api_key=abc", "abc", []string{"[masked:api_key]"}},
		{"apikey", "http://x?apikey=abc", "abc", []string{"[masked:api_key]"}},
		{"password", "http://x?password=abc", "abc", []string{"[masked:password]"}},
		{"auth", "http://x?auth=abc", "abc", []string{"[masked:auth]"}},
	}
	for _, c := range cases {
		out := m.MaskURL(c.in)
		if strings.Contains(out, c.notHave) {
			t.Fatalf("%s: out %q contains leaked %q", c.name, out, c.notHave)
		}
		for _, want := range c.mustHave {
			if !strings.Contains(out, want) {
				t.Fatalf("%s: out %q missing %q", c.name, out, want)
			}
		}
	}
}
