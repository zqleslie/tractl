// Package diagnostics assembles all observability events collected across the
// tractl pipeline into a single, credential-safe ExecutionRecord artifact.
//
// Credential safety is caller-enforced: the Collector accepts pre-masked values
// only. TraceEvent.Detail, AssertionRecord.Message, and RequestRecord.URL MUST
// have credential values masked by the caller before being passed to any
// Collector method. See ADR-014 §6.
package diagnostics

import "time"

// ExecutionRecord is the complete credential-safe observability artifact for
// one tractl engine.Run() invocation. Produced by Collector.Record().
type ExecutionRecord struct {
	TraceID   string
	StartedAt time.Time
	Duration  time.Duration
	Workflows []WorkflowRecord
	Events    []TraceEvent // provenance trace (ADR-014 §5), chronological order
}

// WorkflowRecord holds the per-workflow observability data.
type WorkflowRecord struct {
	WorkflowID string
	StartedAt  time.Time
	Duration   time.Duration
	Outcome    string // "pass" | "fail" | "skipped"
	Steps      []StepRecord
}

// StepRecord holds the per-step observability data.
type StepRecord struct {
	StepID       string
	StartedAt    time.Time
	Duration     time.Duration
	DependsOn    []string      // step IDs this step waited for
	WaitDuration time.Duration // time spent waiting for dependencies
	Outcome      string        // "pass" | "fail" | "skipped" | "dependency-skipped"
	Requests     []RequestRecord
	Assertions   []AssertionRecord
	Extracts     []ExtractRecord
}

// RequestRecord represents one HTTP attempt (including retries).
// Attempt 1 is the initial request; Attempt 2+ are retries.
// URL must be pre-masked by the caller per ADR-014 §6.
type RequestRecord struct {
	Attempt     int
	Method      string
	URL         string // credential-safe (query params with credential patterns masked)
	Status      int
	Timeline    TimingRecord
	Redirects   []RedirectRecord
	RetryReason string // empty for attempt 1
	BackoffMs   int64  // 0 for attempt 1 or fixed-delay retries
}

// TimingRecord captures the seven mandatory segments from ADR-014 §2.
// All values are milliseconds for JSON/YAML serialisation friendliness.
type TimingRecord struct {
	DNSMs         int64
	TCPMs         int64
	TLSMs         int64
	RequestSentMs int64
	TTFBMs        int64
	TransferMs    int64
	TotalMs       int64
}

// DurationToTimingRecord converts individual duration fields (matching the
// shape of runtime.TimingData) into a TimingRecord without importing
// internal/runtime.
func DurationToTimingRecord(dns, tcp, tls, sent, ttfb, transfer, total time.Duration) TimingRecord {
	return TimingRecord{
		DNSMs:         dns.Milliseconds(),
		TCPMs:         tcp.Milliseconds(),
		TLSMs:         tls.Milliseconds(),
		RequestSentMs: sent.Milliseconds(),
		TTFBMs:        ttfb.Milliseconds(),
		TransferMs:    transfer.Milliseconds(),
		TotalMs:       total.Milliseconds(),
	}
}

// LatencyAttribution returns the percentage of TotalMs attributed to each
// segment (0–100 each). Returns a map with zero values if TotalMs is zero.
func (t TimingRecord) LatencyAttribution() map[string]float64 {
	result := map[string]float64{
		"dns":          0,
		"tcp":          0,
		"tls":          0,
		"request_sent": 0,
		"ttfb":         0,
		"transfer":     0,
	}
	if t.TotalMs == 0 {
		return result
	}
	total := float64(t.TotalMs)
	result["dns"] = float64(t.DNSMs) / total * 100
	result["tcp"] = float64(t.TCPMs) / total * 100
	result["tls"] = float64(t.TLSMs) / total * 100
	result["request_sent"] = float64(t.RequestSentMs) / total * 100
	result["ttfb"] = float64(t.TTFBMs) / total * 100
	result["transfer"] = float64(t.TransferMs) / total * 100
	return result
}

// DominantSegment returns the name of the segment with the highest Ms value.
// Returns "total" if all segments are zero.
func (t TimingRecord) DominantSegment() string {
	segments := []struct {
		name string
		val  int64
	}{
		{"dns", t.DNSMs},
		{"tcp", t.TCPMs},
		{"tls", t.TLSMs},
		{"request_sent", t.RequestSentMs},
		{"ttfb", t.TTFBMs},
		{"transfer", t.TransferMs},
	}
	best := ""
	var bestVal int64
	for _, s := range segments {
		if s.val > bestVal {
			bestVal = s.val
			best = s.name
		}
	}
	if best == "" {
		return "total"
	}
	return best
}

// RedirectRecord captures one hop in an HTTP redirect chain.
type RedirectRecord struct {
	Hop      int
	FromURL  string
	ToURL    string
	Status   int
	Timeline TimingRecord
}

// AssertionRecord captures the credential-safe result of one assertion evaluation.
// Message must be pre-masked by the caller per ADR-014 §6.
type AssertionRecord struct {
	ID            string
	Kind          string // status|header|body|schema|script|extension
	Outcome       string // "pass" | "fail" | "skipped"
	Severity      string // "error" | "warning"
	CausesFailure bool
	Message       string // credential-safe failure detail; empty on pass
}

// ExtractRecord captures the result of one extract evaluation.
type ExtractRecord struct {
	ID       string
	Source   string // status|header|body|metadata|timing|extension
	As       string // variable name it was bound to
	Resolved bool
}
