package soul_test

import (
	"context"
	"testing"

	"novelclaw/pkg/soul"
)

func TestSoulController_JobLifecycle(t *testing.T) {
	ctrl := soul.NewControllerWithDir(t.TempDir())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	job := ctrl.RegisterJob("job_101", "proj_alpha", 1, cancel)
	if job == nil {
		t.Fatal("expected job to be registered")
	}

	if ctrl.GetSoul().CurrentMood != soul.StateTranslating {
		t.Errorf("expected mood Translating, got %s", ctrl.GetSoul().CurrentMood)
	}

	ok := ctrl.ApplyStylePatch("job_101", "increase wuxia tone")
	if !ok {
		t.Fatal("failed to apply style patch")
	}

	patch := ctrl.GetJobStylePatch("job_101")
	if patch != "increase wuxia tone" {
		t.Errorf("expected patch 'increase wuxia tone', got %s", patch)
	}

	if ctrl.IsSoftStopRequested("job_101") {
		t.Error("soft stop should initially be false")
	}

	if !ctrl.RequestSoftStop("job_101") {
		t.Fatal("expected RequestSoftStop to succeed")
	}

	if !ctrl.IsSoftStopRequested("job_101") {
		t.Error("soft stop should be true after request")
	}

	if !ctrl.RequestHardAbort("job_101") {
		t.Fatal("expected RequestHardAbort to succeed")
	}

	select {
	case <-ctx.Done():
	default:
		t.Error("expected context to be cancelled by hard abort")
	}

	if ctrl.GetSoul().CurrentMood != soul.StateIdle {
		t.Errorf("expected mood Idle after abort, got %s", ctrl.GetSoul().CurrentMood)
	}
}
