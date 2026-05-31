# kv-str-store: key-value string store

Basic, non-production ready, mixed key-value store (**in memory** keys, reading values from **disk**). The concurrency model is basically a producer-consumer architecture, since we maintain a single writer head and multiple non-blocking reads (abstracted as segments: only one segment is active).

TODO: Compaction goroutine in the background.
Based off: Designing Data-Intensive Apps (Kleppman, 2017, pp. 70-75).


## Run

1. `go mod tidy`
2. `go run ./cmd/repl`


## Test Suite

1. Sync reads and writes
2. Segment sizes
3. Concurrent reads and writes
