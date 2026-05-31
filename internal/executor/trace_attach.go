//go:build !js

package executor

import (
	"context"
	"net/http/httptrace"
)

func attachHTTPTrace(ctx context.Context, recorder *TimingRecorder) context.Context {
	return httptrace.WithClientTrace(ctx, recorder.ClientTrace())
}
