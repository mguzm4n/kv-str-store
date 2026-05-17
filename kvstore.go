package main

type Segment struct {
	keyToOffsetMap map[string]uint64
	fd             uint64
}

type Store struct {
	Segments      []Segment
	ActiveSegment *Segment
}

func New() *Store {
	return &Store{}
}

func orchestrator() {

}

func writer() {

}

func compact() {

}
