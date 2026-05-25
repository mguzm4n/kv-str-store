package main

import (
	"errors"
	"sync"
	"sync/atomic"
)

var KEY_SIZE_BYTES = 2
var VALUE_SIZE_BYTES = 4

type Store struct {
	mu            sync.RWMutex
	Segments      []*Segment
	ActiveSegment *Segment
}

func (s *Store) newSegment() (*Segment, error) {
	size := len(s.Segments) // !!! assume we hold the store lock
	return NewSegment("segment", size)
}

func NewStore() (*Store, error) {
	store := &Store{}
	store.Segments = make([]*Segment, 0)
	// TODO: recover state from disk

	active, err := store.newSegment()
	if err != nil {
		return nil, errors.New("Couldn't bootstrap the store")
	}
	store.ActiveSegment = active

	go store.compact()
	return store, nil
}

func (s *Store) PutKey(key, value string) error {
	s.mu.Lock()

	if atomic.LoadInt64(&s.ActiveSegment.size) > SEGMENT_CAPACITY { // soft limit - can be an outdated check immediately after this instruction on multiple putKeys
		// !!! file "closed" for writing
		s.Segments = append(s.Segments, s.ActiveSegment)
		activeSegment, _ := s.newSegment()
		s.ActiveSegment = activeSegment
	}

	currentActiveSeg := s.ActiveSegment
	currentActiveSeg.mu.Lock()
	defer currentActiveSeg.mu.Unlock()
	s.mu.Unlock() // !!! we can only unlock once we secured the active segment

	nBytes, err := PutKey(currentActiveSeg.keyToOffsetMap, currentActiveSeg.file, key, value)
	if err != nil {
		return err
	}
	atomic.AddInt64(&currentActiveSeg.size, int64(nBytes))
	return nil
}

func (s *Store) GetKey(key string) (string, error) {
	s.mu.RLock()
	activeSegment := s.ActiveSegment
	s.mu.RUnlock() // *can* drop the lock once i copied the pointer

	val, err := activeSegment.Lookup(key)
	if err == nil {
		return val, nil
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	for i := len(s.Segments) - 1; i >= 0; i-- {
		segment := s.Segments[i]
		val, err = segment.Lookup(key)
		if err == nil {
			return val, nil
		}
	}

	return "", err
}

func (s *Store) compact() {

}
