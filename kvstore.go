package main

import (
	"errors"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
)

var KEY_SIZE_BYTES = 2
var VALUE_SIZE_BYTES = 4

const (
	KB               = 1 << 10
	MB               = 1 << 20
	SEGMENT_CAPACITY = 64 * MB
)

type Segment struct {
	mu             sync.RWMutex
	keyToOffsetMap map[string]int64
	file           *os.File
	size           int64
}

type Store struct {
	mu            sync.RWMutex
	Segments      []*Segment
	ActiveSegment *Segment
}

func (s *Store) newSegment(filename string) (*Segment, error) {
	segment := &Segment{
		keyToOffsetMap: make(map[string]int64),
	}
	s.mu.RLock()
	fname := fmt.Sprintf("%s-%d", filename, len(s.Segments))
	s.mu.RUnlock()
	f, err := os.OpenFile(fname, os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return nil, errors.New("Couldn't create new segment")
	}
	segment.file = f
	return segment, nil
}

func New() (*Store, error) {
	store := &Store{}
	store.Segments = make([]*Segment, 0)
	// TODO: recover state from disk

	active, err := store.newSegment("segment")
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
		s.ActiveSegment.file.Close()
		s.Segments = append(s.Segments, s.ActiveSegment)
		activeSegment, _ := s.newSegment("segment")
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

func (s *Segment) Lookup(key string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	offset, ok := s.keyToOffsetMap[key]
	if !ok {
		return "", errors.New("No value found")
	}
	return GetKey(s.keyToOffsetMap, s.file, key, offset)
}

func (s *Store) GetKey(key string) (string, error) {
	s.mu.RLock()
	activeSegment := s.ActiveSegment
	s.mu.RUnlock()

	val, err := activeSegment.Lookup(key)
	if err == nil {
		return val, nil
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	for i := len(s.Segments) - 1; i >= 0; i-- {
		segment := s.Segments[i]
		val, err := segment.Lookup(key)
		if err == nil {
			return val, nil
		}
	}

	return "", nil
}

func (s *Store) compact() {

}
