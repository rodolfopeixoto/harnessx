package containers

import (
	"runtime"
	"testing"
	"time"
)

func TestKnownRuntimeIDs(t *testing.T) {
	ids := KnownRuntimeIDs()
	if len(ids) < 4 {
		t.Fatalf("expected several runtimes, got %v", ids)
	}
	want := map[string]bool{"docker": true, "podman": true, "orbstack": true, "apple_container": true, "colima": true}
	for _, id := range ids {
		if !want[id] {
			t.Errorf("unexpected id %q", id)
		}
	}
}

func TestByID_Unknown(t *testing.T) {
	if _, err := ByID("nope"); err == nil {
		t.Fatal("expected error for unknown id")
	}
}

func TestByID_Docker(t *testing.T) {
	r, err := ByID("docker")
	if err != nil {
		t.Fatal(err)
	}
	if r.ID() != "docker" || r.Binary() != "docker" {
		t.Fatalf("unexpected: %+v", r)
	}
}

func TestPlatformPreferenceContainsExpected(t *testing.T) {
	pref := platformPreference()
	if runtime.GOOS == "darwin" && pref[0] != "apple_container" {
		t.Fatalf("darwin should prefer apple_container, got %v", pref)
	}
	if runtime.GOOS != "darwin" && pref[0] != "docker" {
		t.Fatalf("non-darwin should prefer docker, got %v", pref)
	}
}

func TestShouldPrune_RespectsStoppedFlag(t *testing.T) {
	stopped := Container{ID: "a", State: "exited", Status: "Exited (0)"}
	running := Container{ID: "b", State: "running", Status: "Up 5m"}
	cutoff := time.Time{}
	if !shouldPrune(stopped, PruneOptions{Stopped: true}, cutoff) {
		t.Fatal("stopped container should be pruned with Stopped=true")
	}
	if shouldPrune(running, PruneOptions{Stopped: true}, cutoff) {
		t.Fatal("running container should be skipped with Stopped=true")
	}
}

func TestShouldPrune_OlderThan(t *testing.T) {
	old := Container{ID: "a", State: "exited", CreatedAt: time.Now().Add(-48 * time.Hour)}
	fresh := Container{ID: "b", State: "exited", CreatedAt: time.Now().Add(-30 * time.Minute)}
	cutoff := time.Now().Add(-2 * time.Hour)
	if !shouldPrune(old, PruneOptions{Stopped: true, OlderThan: 2 * time.Hour}, cutoff) {
		t.Fatal("old stopped container should be pruned")
	}
	if shouldPrune(fresh, PruneOptions{Stopped: true, OlderThan: 2 * time.Hour}, cutoff) {
		t.Fatal("fresh stopped container should be skipped under OlderThan filter")
	}
}

func TestShouldPrune_AllIncludesRunning(t *testing.T) {
	running := Container{ID: "b", State: "running", Status: "Up 5m"}
	if !shouldPrune(running, PruneOptions{All: true}, time.Time{}) {
		t.Fatal("All=true should include running containers")
	}
}

func TestParseDockerJSON(t *testing.T) {
	raw := []byte(`{"ID":"abc","Names":"web","Image":"nginx","Status":"Up 5m","State":"running"}
{"ID":"def","Names":"db","Image":"postgres","Status":"Exited","State":"exited"}`)
	got, err := parseDockerJSON(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(got))
	}
	if got[0].ID != "abc" || got[0].Name != "web" {
		t.Fatalf("unexpected row: %+v", got[0])
	}
}
