package diagnostics

import (
	"sort"
	"sync"
	"time"
)

// Collector is the mutable accumulator used during a single engine.Run().
// It is goroutine-safe: the engine may call it concurrently from the parallel
// step scheduler.
//
// Credential safety is caller-enforced. All values passed to Collector methods
// (Detail in Emit, URLs in RequestRecord, Messages in AssertionRecord) MUST
// already have credential content masked per ADR-014 §6.
type Collector struct {
	mu        sync.Mutex
	traceID   string
	origin    time.Time // time.Now() at construction; monotonic component preserved
	events    []TraceEvent
	workflows map[string]*wfEntry // keyed by workflowID
	wfOrder   []string            // insertion-ordered workflow IDs
}

type wfEntry struct {
	record    WorkflowRecord
	stepOrder []string
	steps     map[string]*StepRecord
}

// NewCollector constructs a Collector with a fixed traceID and records the
// current time as the monotonic origin for all subsequent Emit calls.
func NewCollector(traceID string) *Collector {
	return &Collector{
		traceID:   traceID,
		origin:    time.Now(),
		workflows: make(map[string]*wfEntry),
	}
}

// Emit appends a TraceEvent, computing MonotonicMs automatically from the
// collector's origin time.
func (c *Collector) Emit(kind TraceEventKind, workflowID, stepID, spanID, detail string) {
	ms := time.Since(c.origin).Milliseconds()
	c.mu.Lock()
	c.events = append(c.events, TraceEvent{
		MonotonicMs: ms,
		Kind:        kind,
		WorkflowID:  workflowID,
		StepID:      stepID,
		SpanID:      spanID,
		Detail:      detail,
	})
	c.mu.Unlock()
}

// StartWorkflow records the start of a workflow. Safe to call concurrently.
func (c *Collector) StartWorkflow(workflowID string) {
	now := time.Now()
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, exists := c.workflows[workflowID]; !exists {
		c.workflows[workflowID] = &wfEntry{
			record: WorkflowRecord{
				WorkflowID: workflowID,
				StartedAt:  now,
			},
			steps: make(map[string]*StepRecord),
		}
		c.wfOrder = append(c.wfOrder, workflowID)
	}
}

// CompleteWorkflow records the outcome and duration for a workflow.
func (c *Collector) CompleteWorkflow(workflowID, outcome string, dur time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	entry, ok := c.workflows[workflowID]
	if !ok {
		return
	}
	entry.record.Outcome = outcome
	entry.record.Duration = dur
}

// AddStep appends a fully-populated StepRecord to a workflow. The engine
// constructs StepRecord after a step finishes and calls this once per step.
// Safe to call concurrently for different or the same workflowID.
func (c *Collector) AddStep(workflowID string, step StepRecord) {
	c.mu.Lock()
	defer c.mu.Unlock()
	entry, ok := c.workflows[workflowID]
	if !ok {
		return
	}
	entry.stepOrder = append(entry.stepOrder, step.StepID)
	s := step
	entry.steps[step.StepID] = &s
}

// Record assembles and returns the final ExecutionRecord.
//
// Workflows appear in StartWorkflow call order. Steps within each workflow
// appear in AddStep call order. TraceEvents are sorted ascending by
// MonotonicMs. Safe to call once after all goroutines have finished.
func (c *Collector) Record() ExecutionRecord {
	c.mu.Lock()
	defer c.mu.Unlock()

	wfRecords := make([]WorkflowRecord, 0, len(c.wfOrder))
	for _, wfID := range c.wfOrder {
		entry := c.workflows[wfID]
		steps := make([]StepRecord, 0, len(entry.stepOrder))
		for _, sid := range entry.stepOrder {
			if s, ok := entry.steps[sid]; ok {
				steps = append(steps, *s)
			}
		}
		rec := entry.record
		rec.Steps = steps
		wfRecords = append(wfRecords, rec)
	}

	events := make([]TraceEvent, len(c.events))
	copy(events, c.events)
	sort.Slice(events, func(i, j int) bool {
		return events[i].MonotonicMs < events[j].MonotonicMs
	})

	// Wall-clock elapsed since the collector was created. Summing individual
	// workflow durations would overcount for parallel workflows.
	total := time.Since(c.origin)

	return ExecutionRecord{
		TraceID:   c.traceID,
		StartedAt: c.origin,
		Duration:  total,
		Workflows: wfRecords,
		Events:    events,
	}
}
