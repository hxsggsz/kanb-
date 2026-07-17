package update

import (
	"context"
	"errors"
	"testing"
	"time"
)

func withFastBackoff(t *testing.T) {
	t.Helper()
	orig := backoffSchedule
	backoffSchedule = []time.Duration{0, 0, 0, 0, 0}
	t.Cleanup(func() { backoffSchedule = orig })
}

func withFetch(t *testing.T, fn func(ctx context.Context) (string, error)) {
	t.Helper()
	orig := fetchLatestTagFn
	fetchLatestTagFn = fn
	t.Cleanup(func() { fetchLatestTagFn = orig })
}

func TestCheckLatest_NewerVersionAvailable(t *testing.T) {
	withFetch(t, func(ctx context.Context) (string, error) { return "v2.0.0", nil })

	latest, available, err := CheckLatest(context.Background(), "v1.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !available || latest != "v2.0.0" {
		t.Fatalf("got latest=%q available=%v, want v2.0.0/true", latest, available)
	}
}

func TestCheckLatest_AlreadyUpToDate(t *testing.T) {
	withFetch(t, func(ctx context.Context) (string, error) { return "v1.0.0", nil })

	_, available, err := CheckLatest(context.Background(), "v1.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if available {
		t.Fatalf("expected available=false when versions match")
	}
}

func TestCheckLatest_RetriesThenGivesUpSilently(t *testing.T) {
	withFastBackoff(t)

	calls := 0
	withFetch(t, func(ctx context.Context) (string, error) {
		calls++
		return "", errors.New("network error")
	})

	_, available, err := CheckLatest(context.Background(), "v1.0.0")
	if err != nil {
		t.Fatalf("expected nil error on exhausted retries, got %v", err)
	}
	if available {
		t.Fatalf("expected available=false on exhausted retries")
	}
	if calls != maxAttempts {
		t.Fatalf("got %d attempts, want %d", calls, maxAttempts)
	}
}

func TestCheckLatest_SucceedsAfterTransientFailures(t *testing.T) {
	withFastBackoff(t)

	calls := 0
	withFetch(t, func(ctx context.Context) (string, error) {
		calls++
		if calls < 3 {
			return "", errors.New("network error")
		}
		return "v1.5.0", nil
	})

	latest, available, err := CheckLatest(context.Background(), "v1.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !available || latest != "v1.5.0" {
		t.Fatalf("got latest=%q available=%v, want v1.5.0/true", latest, available)
	}
}

func TestCheckLatest_MalformedTagIsSilent(t *testing.T) {
	withFetch(t, func(ctx context.Context) (string, error) { return "not-a-version", nil })

	_, available, err := CheckLatest(context.Background(), "v1.0.0")
	if err != nil {
		t.Fatalf("expected nil error for malformed tag, got %v", err)
	}
	if available {
		t.Fatalf("expected available=false for malformed tag")
	}
}
