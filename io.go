package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"os"
)

func PutKey(keyToByteOffset map[string]int64, file *os.File, key, value string) (uint64, error) {
	if len(key) > math.MaxUint16 {
		return 0, errors.New("key size exceeds maximum for 2^16 (65535 bytes)")
	}
	if len(value) > math.MaxUint32 {
		return 0, errors.New("value size exceeds maximum 2^32 (4GB)")
	}

	totalSize := 6 + len(key) + len(value)
	buffer := make([]byte, totalSize) // 2 (key) + 4 (value) + key/value contents' size
	binary.BigEndian.PutUint16(buffer[0:2], uint16(len(key)))
	binary.BigEndian.PutUint32(buffer[2:6], uint32(len(value)))

	keyEnd := 6 + len(key)
	valueEnd := keyEnd + len(value)
	copy(buffer[6:keyEnd], key)
	copy(buffer[keyEnd:valueEnd], value)

	startPos, err := file.Seek(0, io.SeekEnd)
	if err != nil {
		log.Fatalf("Failed to seek to end of file: %v", err)
	}

	bytesWritten, err := file.Write(buffer)
	if err != nil {
		log.Fatal("Couldn't write to disk")
	}
	keyToByteOffset[key] = startPos
	fmt.Printf("written: %d with startingPos: %d\n", bytesWritten, startPos)
	return uint64(totalSize), nil
}

// TODO: add logging indicator (if debug or test, log)
func GetKey(file *os.File, key string, offset int64) (string, error) {
	var err error
	// whence -> from where do i start reading/writing bytes?

	r := io.NewSectionReader(file, offset, math.MaxInt64)
	var keySize int16   // 2 bytes buffer
	var valueSize int32 // 4 bytes buffer

	err = binary.Read(r, binary.BigEndian, &keySize)
	if err != nil {
		log.Fatalf("Failed to read key size: %v", err)
	}
	fmt.Printf("keySize=%d\n", keySize)

	err = binary.Read(r, binary.BigEndian, &valueSize)
	if err != nil {
		log.Fatalf("Failed to read value size: %v", err)
	}
	fmt.Printf("valueSize=%d\n", valueSize)

	keyReadBuffer := make([]byte, keySize)
	_, err = io.ReadFull(r, keyReadBuffer)
	if err != nil {
		log.Fatalf("Failed to read key: %v", err)
	}
	fmt.Printf("key=%s\n", keyReadBuffer)

	valueReadBuffer := make([]byte, valueSize)
	_, err = io.ReadFull(r, valueReadBuffer)
	if err != nil {
		log.Fatalf("Failed to read value: %v", err)
	}
	fmt.Printf("value=%s\n", valueReadBuffer)

	return string(valueReadBuffer), nil
}
