// Package executor implements the HTTP step executor and request builder.
package executor

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/tractl/tractl/internal/compiler"
	"github.com/tractl/tractl/internal/runtime"
)

// RequestBuilder builds an *http.Request from a CompiledStep using expression evaluation.
type RequestBuilder struct {
	eval *runtime.ExpressionEvaluator
}

// NewRequestBuilder creates a builder bound to an evaluator.
func NewRequestBuilder(eval *runtime.ExpressionEvaluator) *RequestBuilder {
	return &RequestBuilder{eval: eval}
}

// Build constructs the http.Request for the step. ctx should carry step timeout.
func (b *RequestBuilder) Build(ctx context.Context, step *compiler.CompiledStep) (*http.Request, error) {
	if step == nil || step.Request == nil {
		return nil, errors.New("no request payload")
	}
	// resolve target
	target, err := b.eval.Evaluate(ctx, step.Request.Target)
	if err != nil {
		return nil, err
	}
	// resolve method
	method, err := b.eval.Evaluate(ctx, step.Request.Method)
	if err != nil {
		return nil, err
	}
	if method == "" {
		method = "GET"
	}

	var bodyReader io.Reader
	if step.Request.Body != nil {
		switch strings.ToLower(step.Request.Body.Encoding) {
		case "json":
			// marshal content
			bs, jerr := json.Marshal(step.Request.Body.Content)
			if jerr != nil {
				return nil, jerr
			}
			bodyReader = bytes.NewReader(bs)
		case "form":
			// expect map[string]string
			vals := url.Values{}
			if m, ok := step.Request.Body.Content.(map[string]string); ok {
				for k, v := range m {
					vals.Set(k, v)
				}
				bodyReader = strings.NewReader(vals.Encode())
			} else {
				// unsupported form content shape
				return nil, errors.New("unsupported form content")
			}
		default:
			// no body
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, target, bodyReader)
	if err != nil {
		return nil, err
	}
	// set content-type for json/form
	if step.Request.Body != nil {
		switch strings.ToLower(step.Request.Body.Encoding) {
		case "json":
			req.Header.Set("Content-Type", "application/json")
		case "form":
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		}
	}

	// set headers
	if len(step.Request.Headers) > 0 {
		hmap, herr := b.eval.EvaluateMap(ctx, step.Request.Headers)
		if herr != nil {
			return nil, herr
		}
		for k, v := range hmap {
			req.Header.Set(k, v)
		}
	}

	return req, nil
}
