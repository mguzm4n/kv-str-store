package main

import (
	"errors"
	"fmt"
	"os"
	"sync"
)

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

func NewSegment(basename string, correlative int) (*Segment, error) {
	fname := fmt.Sprintf("%s-%d.log", basename, correlative)

	segment := &Segment{
		keyToOffsetMap: make(map[string]int64),
	}

	// read and write paths + flags using the same handler -> add |os.O_APPEND later
	f, err := os.OpenFile(fname, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, errors.New("Couldn't create new segment")
	}
	segment.file = f
	return segment, nil
}

func (s *Segment) Lookup(key string) (string, error) {
	s.mu.RLock()
	offset, ok := s.keyToOffsetMap[key]
	s.mu.RUnlock() // multiple reads for a file is safe at os-level
	if !ok {
		return "", errors.New("No value found")
	}
	return GetKey(s.file, key, offset)
}
