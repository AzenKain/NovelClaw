package llm

import (
	"context"
	"testing"
	"time"
)

// TestDualStreamBus_PublishAndHotPatch verifies publishing events and hot-patching sentence buffers.
func TestDualStreamBus_PublishAndHotPatch(t *testing.T) {
	bus := NewDualStreamBus(10)
	defer bus.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	bus.Publish(ctx, StreamEvent{
		Index:  0,
		Source: "伊月は机に向かっていた。",
		Draft:  "Itsuki ngồi vào bàn làm việc.",
		Kind:   EventDraftEmitted,
	})

	bus.Publish(ctx, StreamEvent{
		Index:  1,
		Source: "那由多は「先輩」と呼んだ。",
		Draft:  "Nayuta gọi Itsuki là anh yêu.",
		Kind:   EventDraftEmitted,
	})

	patched := bus.HotPatch(1, "Nayuta gọi Itsuki là tiền bối.", "Nayuta uses Senpai in Chapter 1, not Darling")
	if !patched {
		t.Fatalf("expected hot-patch to succeed")
	}

	sentences := bus.GetApprovedSentences()
	if len(sentences) != 2 {
		t.Fatalf("expected 2 sentences, got %d", len(sentences))
	}

	if sentences[0] != "Itsuki ngồi vào bàn làm việc." {
		t.Errorf("sentence 0 mismatch: %s", sentences[0])
	}
	if sentences[1] != "Nayuta gọi Itsuki là tiền bối." {
		t.Errorf("sentence 1 was not hot-patched: %s", sentences[1])
	}
}

// TestDualStreamBus_PublishWhenFull verifies that Publish does not block when the channel buffer is full.
func TestDualStreamBus_PublishWhenFull(t *testing.T) {
	bus := NewDualStreamBus(2)
	defer bus.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	bus.Publish(ctx, StreamEvent{Index: 0, Draft: "one"})
	bus.Publish(ctx, StreamEvent{Index: 1, Draft: "two"})
	published := bus.Publish(ctx, StreamEvent{Index: 2, Draft: "three"})
	if !published {
		t.Errorf("expected publish to succeed via fallback buffer")
	}

	sentences := bus.GetApprovedSentences()
	if len(sentences) != 3 {
		t.Errorf("expected 3 sentences in buffer, got %d", len(sentences))
	}
}

// TestGlobalSyncBus_SubscribeUnsubscribe verifies subscribing, broadcasting, and cleanly unsubscribing.
func TestGlobalSyncBus_SubscribeUnsubscribe(t *testing.T) {
	bus := NewGlobalSyncBus()
	sub1 := bus.Subscribe(5)
	sub2 := bus.Subscribe(5)

	bus.Broadcast(GlobalSyncEvent{Type: SyncEventEntityAdded, ChapterIndex: 1, Key: "Itsuki"})

	select {
	case evt := <-sub1:
		if evt.Key != "Itsuki" {
			t.Errorf("unexpected event key: %s", evt.Key)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatalf("sub1 timed out waiting for event")
	}

	bus.Unsubscribe(sub1)
	close(sub1)

	bus.Broadcast(GlobalSyncEvent{Type: SyncEventEntityAdded, ChapterIndex: 2, Key: "Nayuta"})

	select {
	case evt := <-sub2:
		if evt.Key != "Nayuta" && evt.Key != "Itsuki" {
			t.Errorf("unexpected event key on sub2: %s", evt.Key)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatalf("sub2 timed out waiting for event")
	}
}
