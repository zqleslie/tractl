package executor

import (
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

// CredentialMasker masks sensitive header and URL parameter values for observation.
// All masked values use typed placeholders per ADR-014 §6 (never empty strings).
type CredentialMasker struct{}

// NewCredentialMasker returns a new masker.
func NewCredentialMasker() *CredentialMasker { return &CredentialMasker{} }

// exactHeaderMasks maps canonical (http.CanonicalHeaderKey) header names to
// their typed mask. O(1) lookup replaces the previous O(n) regex scan for
// known exact-match headers.
var exactHeaderMasks = map[string]string{
	"Authorization":       "[masked:bearer_token]",
	"Cookie":              "[masked:cookie]",
	"Set-Cookie":          "[masked:cookie]",
	"X-Api-Key":           "[masked:api_key]",
	"Api-Key":             "[masked:api_key]",
	"Proxy-Authorization": "[masked:bearer_token]",
}

// credentialSuffixRE matches headers whose name ends with a credential-like
// suffix after a hyphen or underscore (e.g. "X-My-Secret", "Service-Token").
var credentialSuffixRE = regexp.MustCompile(`(?i)[-_](key|secret|token|password)$`)

// MaskHeaders returns a copy of headers with credentials masked using typed
// placeholders. The input http.Header is never mutated.
func (c *CredentialMasker) MaskHeaders(headers http.Header) http.Header {
	out := make(http.Header, len(headers))
	for k, vals := range headers {
		canonical := http.CanonicalHeaderKey(k)
		if mask, ok := exactHeaderMasks[canonical]; ok {
			out.Set(k, mask)
			continue
		}
		if credentialSuffixRE.MatchString(k) {
			out.Set(k, "[masked:credential]")
			continue
		}
		for _, v := range vals {
			out.Add(k, v)
		}
	}
	return out
}

// credentialQuery maps a query parameter name pattern to its typed mask.
// Exact-match names use a map; no regex needed for the common cases.
var exactQueryMasks = map[string]string{
	"token":    "[masked:token]",
	"auth":     "[masked:auth]",
	"apikey":   "[masked:api_key]",
	"api_key":  "[masked:api_key]",
	"api-key":  "[masked:api_key]",
	"key":      "[masked:key]",
	"secret":   "[masked:secret]",
	"password": "[masked:password]",
}

// MaskURL masks credential-like query parameter values with typed placeholders.
// Returns the masked URL string; returns the input unchanged if it does not parse.
// The query string is reassembled by hand to preserve the readability of the
// typed placeholder (`[masked:token]`) in display output.
func (c *CredentialMasker) MaskURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	if u.RawQuery == "" {
		return rawURL
	}

	pairs := strings.Split(u.RawQuery, "&")
	changed := false
	for i, p := range pairs {
		eq := strings.IndexByte(p, '=')
		var name string
		if eq < 0 {
			name = p
		} else {
			name = p[:eq]
		}
		decodedName, _ := url.QueryUnescape(name)
		mask, matched := matchQueryMask(decodedName)
		if !matched {
			continue
		}
		pairs[i] = name + "=" + mask
		changed = true
	}
	if !changed {
		return rawURL
	}
	u.RawQuery = strings.Join(pairs, "&")
	// Manual reassembly: url.URL.String() would re-encode the typed placeholder brackets.
	out := ""
	if u.Scheme != "" {
		out += u.Scheme + "://"
	}
	if u.Host != "" {
		out += u.Host
	}
	out += u.Path
	out += "?" + u.RawQuery
	if u.Fragment != "" {
		out += "#" + u.Fragment
	}
	return out
}

func matchQueryMask(param string) (string, bool) {
	low := strings.ToLower(param)
	if mask, ok := exactQueryMasks[low]; ok {
		return mask, true
	}
	return "", false
}

// exactHeaderMasksLower mirrors exactHeaderMasks with lowercased keys, matching
// how response headers are stored in runtime.StepResult (lowercased by executor).
var exactHeaderMasksLower = func() map[string]string {
	m := make(map[string]string, len(exactHeaderMasks))
	for k, v := range exactHeaderMasks {
		m[strings.ToLower(k)] = v
	}
	return m
}()

// MaskResponseHeaders returns a copy of a lowercased response-header map with
// credential values replaced by typed placeholders. The input map is never mutated.
func (c *CredentialMasker) MaskResponseHeaders(headers map[string]string) map[string]string {
	out := make(map[string]string, len(headers))
	for k, v := range headers {
		if mask, ok := exactHeaderMasksLower[k]; ok {
			out[k] = mask
			continue
		}
		if credentialSuffixRE.MatchString(k) {
			out[k] = "[masked:credential]"
			continue
		}
		out[k] = v
	}
	return out
}
