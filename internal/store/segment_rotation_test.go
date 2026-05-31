package store_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mguzm4n/kv-str-store/internal/store"
)

func TestStore_RotatesActiveSegmentWhenFull(t *testing.T) {
	dir := t.TempDir()
	st, err := store.NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore() error: %v", err)
	}
	defer st.CloseSegments()

	// it creates the first active segment file
	if _, err := os.Stat(filepath.Join(dir, "segment-0.log")); err != nil {
		t.Fatalf("expected segment-0.log to exist: %v", err)
	}

	// chose a value size that makes the segment exceed capacity after 2 writes,
	// but not exceed capacity after the first write
	capacity := int64(store.SEGMENT_CAPACITY)
	recordTarget := capacity/2 + 1
	valueLen := int(recordTarget - int64(6+len("same")))
	if valueLen <= 0 {
		t.Fatalf("SEGMENT_CAPACITY too small for rotation test; capacity=%v", store.SEGMENT_CAPACITY)
	}
	big := strings.Repeat("x", valueLen)

	// fill the first segment past the soft capacity -> rotation only
	// happens on the *next* put
	if err := st.PutKey("first", big); err != nil {
		t.Fatalf("PutKey(first) error: %v", err)
	}
	if err := st.PutKey("filling", big); err != nil {
		t.Fatalf("PutKey(filling) error: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "segment-1.log")); err == nil {
		t.Fatalf("expected segment-1.log to not exist before rotation")
	} else if !os.IsNotExist(err) {
		t.Fatalf("stat segment-1.log: %v", err)
	}
	stat, err := os.Stat(filepath.Join(dir, "segment-0.log"))
	if err != nil {
		t.Fatalf("stat segment-0.log: %v", err)
	}
	if stat.Size() <= capacity {
		t.Fatalf("expected segment-0.log to be over capacity before rotation; size=%d capacity=%d", stat.Size(), capacity)
	}

	// now put should rotate
	if err := st.PutKey("trigger", big); err != nil {
		t.Fatalf("PutKey(trigger) error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, "segment-1.log")); err != nil {
		t.Fatalf("expected segment-1.log to exist after rotation: %v", err)
	}

	// New writes go to the new active segment and should shadow older values.
	if err := st.PutKey("first", "new"); err != nil {
		t.Fatalf("PutKey(first=new) error: %v", err)
	}

	got, err := st.GetKey("first")
	if err != nil {
		t.Fatalf("GetKey(first) error: %v", err)
	}
	if got != "new" {
		t.Fatalf("GetKey(first) got %q, want %q", got, "new")
	}

	got, err = st.GetKey("trigger")
	if err != nil {
		t.Fatalf("GetKey(trigger) error: %v", err)
	}
	if got != big {
		t.Fatalf("GetKey(trigger) got %q, want %q", got, big)
	}

	// Keys written to the sealed segment should still be readable.
	for _, k := range []string{"filling"} {
		got, err := st.GetKey(k)
		if err != nil {
			t.Fatalf("GetKey(%s) error: %v", k, err)
		}
		if got != big {
			t.Fatalf("GetKey(%s) got %q, want %q", k, got, big)
		}
	}
}

func TestStore_RotatesMultipleSegments(t *testing.T) {
	dir := t.TempDir()
	st, err := store.NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore() error: %v", err)
	}
	defer st.CloseSegments()

	const rotations = 4
	capacity := int64(store.SEGMENT_CAPACITY)
	recordTarget := capacity/2 + 1
	valueLen := int(recordTarget - int64(6+len("fill-0-a")))
	if valueLen <= 0 {
		t.Fatalf("SEGMENT_CAPACITY too small for rotation test; capacity=%v", store.SEGMENT_CAPACITY)
	}
	big := strings.Repeat("x", valueLen)

	sealedKeys := make([]string, 0, rotations)
	for rot := 0; rot < rotations; rot++ {
		keyA := fmt.Sprintf("fill-%d-a", rot)
		keyB := fmt.Sprintf("fill-%d-b", rot)

		if err := st.PutKey(keyA, big); err != nil {
			t.Fatalf("PutKey(%s) error: %v", keyA, err)
		}
		if err := st.PutKey(keyB, big); err != nil {
			t.Fatalf("PutKey(%s) error: %v", keyB, err)
		}

		nextSegPath := filepath.Join(dir, fmt.Sprintf("segment-%d.log", rot+1))
		if _, err := os.Stat(nextSegPath); err == nil {
			t.Fatalf("expected %s to not exist before rotation", filepath.Base(nextSegPath))
		} else if !os.IsNotExist(err) {
			t.Fatalf("stat %s: %v", filepath.Base(nextSegPath), err)
		}

		curSegPath := filepath.Join(dir, fmt.Sprintf("segment-%d.log", rot))
		stat, err := os.Stat(curSegPath)
		if err != nil {
			t.Fatalf("stat %s: %v", filepath.Base(curSegPath), err)
		}
		if stat.Size() <= capacity {
			t.Fatalf("expected %s to be over capacity before rotation; size=%d capacity=%d", filepath.Base(curSegPath), stat.Size(), capacity)
		}

		sealedKeys = append(sealedKeys, keyA)
		shadowVal := fmt.Sprintf("shadow-%d", rot)
		if err := st.PutKey("shadow", shadowVal); err != nil {
			t.Fatalf("PutKey(shadow) error: %v", err)
		}

		if _, err := os.Stat(nextSegPath); err != nil {
			t.Fatalf("expected %s to exist after rotation: %v", filepath.Base(nextSegPath), err)
		}
	}

	if got, err := st.GetKey("shadow"); err != nil {
		t.Fatalf("GetKey(shadow) error: %v", err)
	} else if got != fmt.Sprintf("shadow-%d", rotations-1) {
		t.Fatalf("GetKey(shadow) got %q, want %q", got, fmt.Sprintf("shadow-%d", rotations-1))
	}

	if got := len(st.Segments); got != rotations {
		t.Fatalf("Segments count got %d, want %d", got, rotations)
	}

	for _, k := range sealedKeys {
		got, err := st.GetKey(k)
		if err != nil {
			t.Fatalf("GetKey(%s) error: %v", k, err)
		}
		if got != big {
			t.Fatalf("GetKey(%s) got %q, want %q", k, got, big)
		}
	}
}
