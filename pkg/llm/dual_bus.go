package llm

import (
	"context"
	"sync"
)

// StreamEventKind indicates the type of event on the dual stream bus.
type StreamEventKind string

const (
	EventDraftEmitted     StreamEventKind = "draft_emitted"
	EventRevisionProposed StreamEventKind = "revision_proposed"
	EventRevisionApplied  StreamEventKind = "revision_applied"
	EventSentenceApproved StreamEventKind = "sentence_approved"
	EventStreamCompleted  StreamEventKind = "stream_completed"
	EventStreamError      StreamEventKind = "stream_error"
)

// StreamEvent carries sentence-level translation data and criticism events.
type StreamEvent struct {
	Index       int             `json:"index"`
	Source      string          `json:"source"`
	Draft       string          `json:"draft"`
	Revised     string          `json:"revised"`
	Reason      string          `json:"reason"`
	Kind        StreamEventKind `json:"kind"`
	Error       error           `json:"-"`
}

// DualStreamBus coordinates streaming events and lock-step synchronization between agents.
type DualStreamBus struct {
	eventsChan chan StreamEvent
	buffer     []StreamEvent
	mu         sync.RWMutex
	closed     bool
}

// NewDualStreamBus creates a new DualStreamBus with a buffered event channel.
func NewDualStreamBus(bufferSize int) *DualStreamBus {
	if bufferSize <= 0 {
		bufferSize = 64
	}
	return &DualStreamBus{
		eventsChan: make(chan StreamEvent, bufferSize),
		buffer:     make([]StreamEvent, 0),
	}
}

// Events returns the receive-only channel of stream events.
func (b *DualStreamBus) Events() <-chan StreamEvent {
	return b.eventsChan
}

// Publish publishes an event to subscribers and records it in the history buffer.
func (b *DualStreamBus) Publish(ctx context.Context, evt StreamEvent) bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return false
	}
	b.buffer = append(b.buffer, evt)

	select {
	case <-ctx.Done():
		return false
	case b.eventsChan <- evt:
		return true
	default:
		return true
	}
}

// HotPatch updates the final approved text for a specific sentence index in the buffer.
func (b *DualStreamBus) HotPatch(index int, revisedText, reason string) bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	for i := len(b.buffer) - 1; i >= 0; i-- {
		if b.buffer[i].Index == index {
			b.buffer[i].Revised = revisedText
			b.buffer[i].Reason = reason
			b.buffer[i].Kind = EventRevisionApplied
			return true
		}
	}
	return false
}

// GetApprovedSentences aggregates the latest approved or revised text for all sentences.
func (b *DualStreamBus) GetApprovedSentences() []string {
	b.mu.RLock()
	defer b.mu.RUnlock()

	latestByIndex := make(map[int]string)
	maxIdx := -1

	for _, evt := range b.buffer {
		if evt.Index > maxIdx {
			maxIdx = evt.Index
		}
		if evt.Revised != "" {
			latestByIndex[evt.Index] = evt.Revised
		} else if evt.Draft != "" {
			latestByIndex[evt.Index] = evt.Draft
		}
	}

	result := make([]string, 0, maxIdx+1)
	for i := 0; i <= maxIdx; i++ {
		if text, exists := latestByIndex[i]; exists {
			result = append(result, text)
		}
	}
	return result
}

// Close closes the event channel safely.
func (b *DualStreamBus) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if !b.closed {
		b.closed = true
		close(b.eventsChan)
	}
}
