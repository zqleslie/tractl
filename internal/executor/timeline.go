package executor

import (
	"crypto/tls"
	"net/http/httptrace"
	"time"

	"github.com/tractl/tractl/internal/runtime"
)

// TimingRecorder records httptrace timings for a single request.
type TimingRecorder struct {
	// monotonic timestamps
	wallClockStart time.Time
	dnsStart       time.Time
	dnsDone        time.Time
	connectStart   time.Time
	connectDone    time.Time
	tlsStart       time.Time
	tlsDone        time.Time
	wroteRequest   time.Time
	firstByte      time.Time
	bodyDone       time.Time

	// redirect hops recorded via CheckRedirect (stored on request context)
	Redirects []RedirectHop
}

// RedirectHop records a single redirect.
type RedirectHop struct {
	HopNumber  int
	SourceURL  string
	TargetURL  string
	StatusCode int
}

// NewTimingRecorder initializes a recorder with wallClockStart set.
func NewTimingRecorder() *TimingRecorder {
	return &TimingRecorder{wallClockStart: time.Now()}
}

// ClientTrace returns an httptrace.ClientTrace wired to record timings.
func (tr *TimingRecorder) ClientTrace() *httptrace.ClientTrace {
	return &httptrace.ClientTrace{
		DNSStart:             func(_ httptrace.DNSStartInfo) { tr.dnsStart = time.Now() },
		DNSDone:              func(_ httptrace.DNSDoneInfo) { tr.dnsDone = time.Now() },
		ConnectStart:         func(_, _ string) { tr.connectStart = time.Now() },
		ConnectDone:          func(_, _ string, _ error) { tr.connectDone = time.Now() },
		TLSHandshakeStart:    func() { tr.tlsStart = time.Now() },
		TLSHandshakeDone:     func(_ tls.ConnectionState, _ error) { tr.tlsDone = time.Now() },
		WroteRequest:         func(_ httptrace.WroteRequestInfo) { tr.wroteRequest = time.Now() },
		GotFirstResponseByte: func() { tr.firstByte = time.Now() },
	}
}

// ToRequestTimeline computes durations and returns a runtime.RequestTimeline.
func (tr *TimingRecorder) ToRequestTimeline(statusCode int, bodySize int64) *runtime.RequestTimeline {
	// ensure bodyDone set
	if tr.bodyDone.IsZero() {
		tr.bodyDone = time.Now()
	}

	dns := durationIfSet(tr.dnsStart, tr.dnsDone)
	connect := durationIfSet(tr.connectStart, tr.connectDone)
	tlsHandshake := durationIfSet(tr.tlsStart, tr.tlsDone)

	// RequestSent: connection ready → last byte of request sent.
	// Connection ready is TLS done for HTTPS, or TCP connect done for plain HTTP.
	connEnd := tr.connectDone
	if !tr.tlsDone.IsZero() {
		connEnd = tr.tlsDone
	}
	requestSent := durationIfSet(connEnd, tr.wroteRequest)

	// TTFB: request sent → first byte of response received.
	ttfb := durationIfSet(tr.wroteRequest, tr.firstByte)

	// ResponseTransfer: first byte received → last byte of response received.
	transfer := durationIfSet(tr.firstByte, tr.bodyDone)

	total := time.Since(tr.wallClockStart)

	return &runtime.RequestTimeline{
		DNSResolution:    dns,
		TCPConnect:       connect,
		TLSHandshake:     tlsHandshake,
		RequestSent:      requestSent,
		TimeToFirstByte:  ttfb,
		ResponseTransfer: transfer,
		TotalDuration:    total,
		WallClockStart:   tr.wallClockStart,
		StatusCode:       statusCode,
		ResponseSize:     bodySize,
	}
}

// helper
func durationIfSet(start, end time.Time) time.Duration {
	if start.IsZero() || end.IsZero() {
		return 0
	}
	return end.Sub(start)
}
