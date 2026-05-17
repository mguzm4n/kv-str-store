package main

import (
	"errors"
	"os"
	"sync"
)

type Segment struct {
	mu             sync.RWMutex
	keyToOffsetMap map[string]uint64
	file           *os.File
}

type Store struct {
	mu            sync.Mutex
	Segments      []*Segment
	ActiveSegment *Segment
}

func newSegment(filename string) (*Segment, error) {
	segment := &Segment{
		keyToOffsetMap: make(map[string]uint64),
	}
	f, err := os.OpenFile(filename, os.O_CREATE, 0644)
	if err != nil {
		return nil, errors.New("Couldn't create new segment")
	}
	segment.file = f
	return segment, nil
}

func New() (*Store, error) {
	segments := make([]*Segment, 0)
	// TODO: recover state from disk

	active, err := newSegment("active")
	if err != nil {
		return nil, errors.New("Couldn't bootstrap the store")
	}

	store := &Store{
		Segments:      segments,
		ActiveSegment: active,
	}
	go store.compact()
	return store, nil
}

func (s *Store) PutKey(key, value string) {
	activeSegment := s.ActiveSegment
	activeSegment.mu.Lock()
	defer activeSegment.mu.Unlock()
}

func (s *Store) GetKey(key string) (string, error) {
	activeSegment := s.ActiveSegment
	activeSegment.mu.Lock()
	defer activeSegment.mu.Unlock()
	return "", nil
}

func (s *Store) compact() {

}
