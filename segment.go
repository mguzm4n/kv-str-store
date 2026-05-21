package main

import (
	"errors"
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

func (s *Segment) Lookup(key string) (string, error) {
	s.mu.RLock()
	offset, ok := s.keyToOffsetMap[key]
	s.mu.RUnlock() // multiple reads for a file is safe at os-level
	if !ok {
		return "", errors.New("No value found")
	}
	return GetKey(s.file, key, offset)
}
