package llm

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// SyncEventType identifies the type of knowledge discovered by a swarm worker.
type SyncEventType string

const (
	SyncEventEntityAdded     SyncEventType = "entity_added"
	SyncEventRelationChanged SyncEventType = "relation_changed"
	SyncEventGlossaryUpdated SyncEventType = "glossary_updated"
)

// GlobalSyncEvent carries shared story knowledge discovered across parallel workers.
type GlobalSyncEvent struct {
	Type         SyncEventType `json:"type"`
	ChapterIndex int64         `json:"chapter_index"`
	Key          string        `json:"key"`
	Value        string        `json:"value"`
}

// GlobalSyncBus broadcasts entity and terminology discoveries across parallel workers.
type GlobalSyncBus struct {
	subscribers []chan GlobalSyncEvent
	mu          sync.RWMutex
}

// NewGlobalSyncBus creates a new GlobalSyncBus instance.
func NewGlobalSyncBus() *GlobalSyncBus {
	return &GlobalSyncBus{
		subscribers: make([]chan GlobalSyncEvent, 0),
	}
}

// Subscribe registers a new subscriber channel for receiving global sync events.
func (b *GlobalSyncBus) Subscribe(bufferSize int) chan GlobalSyncEvent {
	b.mu.Lock()
	defer b.mu.Unlock()

	ch := make(chan GlobalSyncEvent, bufferSize)
	b.subscribers = append(b.subscribers, ch)
	return ch
}

// Unsubscribe removes a subscriber channel from the active broadcast list.
func (b *GlobalSyncBus) Unsubscribe(ch chan GlobalSyncEvent) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for i, sub := range b.subscribers {
		if sub == ch {
			b.subscribers = append(b.subscribers[:i], b.subscribers[i+1:]...)
			break
		}
	}
}

// Broadcast distributes a sync event to all active subscriber channels non-blockingly.
func (b *GlobalSyncBus) Broadcast(evt GlobalSyncEvent) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for _, ch := range b.subscribers {
		select {
		case ch <- evt:
		default:
			log.Warn().
				Str("event_type", string(evt.Type)).
				Int64("chapter", evt.ChapterIndex).
				Msg("subscriber channel full, dropping global sync event")
		}
	}
}

// SwarmChapterResult captures the translation result and telemetry for a single chapter in the swarm.
type SwarmChapterResult struct {
	ChapterIndex int64               `json:"chapter_index"`
	Translated   string              `json:"translated"`
	Telemetry    *ExecutionTelemetry `json:"telemetry"`
	Error        error               `json:"error,omitempty"`
}

// SwarmArcOrchestrator manages parallel worker pools translating multiple chapters in an arc.
type SwarmArcOrchestrator struct {
	orch    *TranslationOrchestrator
	syncBus *GlobalSyncBus
}

// NewSwarmArcOrchestrator creates a new SwarmArcOrchestrator.
func NewSwarmArcOrchestrator(orch *TranslationOrchestrator) *SwarmArcOrchestrator {
	return &SwarmArcOrchestrator{
		orch:    orch,
		syncBus: NewGlobalSyncBus(),
	}
}

// SyncBus returns the underlying GlobalSyncBus.
func (s *SwarmArcOrchestrator) SyncBus() *GlobalSyncBus {
	return s.syncBus
}

// TranslateArcChapters translates multiple chapters in an arc concurrently using a worker pool.
func (s *SwarmArcOrchestrator) TranslateArcChapters(
	ctx context.Context,
	projectID string,
	chapterIndices []int64,
	opts TranslationOptions,
	statusHandler func(chapterIndex int64, status string),
) ([]SwarmChapterResult, error) {
	concurrency := opts.Concurrency
	if concurrency <= 0 {
		concurrency = 2
	}

	sem := make(chan struct{}, concurrency)
	results := make([]SwarmChapterResult, len(chapterIndices))
	var wg sync.WaitGroup

	startTime := time.Now()
	log.Info().
		Str("project_id", projectID).
		Int("total_chapters", len(chapterIndices)).
		Int("concurrency", concurrency).
		Msg("starting swarm arc-parallel translation")

	for i, cIdx := range chapterIndices {
		wg.Add(1)
		go func(slot int, chapterIndex int64) {
			defer wg.Done()

			sem <- struct{}{}
			defer func() { <-sem }()

			chHandler := &TranslationEventHandler{
				OnStatus: func(status string) {
					if statusHandler != nil {
						statusHandler(chapterIndex, status)
					}
				},
			}

			translated, telemetry, err := s.orch.TranslateChapter(ctx, projectID, chapterIndex, opts, chHandler)
			results[slot] = SwarmChapterResult{
				ChapterIndex: chapterIndex,
				Translated:   translated,
				Telemetry:    telemetry,
				Error:        err,
			}

			if err == nil {
				s.syncBus.Broadcast(GlobalSyncEvent{
					Type:         SyncEventEntityAdded,
					ChapterIndex: chapterIndex,
					Key:          fmt.Sprintf("chapter_%d_completed", chapterIndex),
				})
			}
		}(i, cIdx)
	}

	wg.Wait()
	log.Info().
		Str("project_id", projectID).
		Int("completed_chapters", len(chapterIndices)).
		Dur("total_duration", time.Since(startTime)).
		Msg("swarm arc-parallel translation completed")

	return results, nil
}
