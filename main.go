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

var KEY_SIZE_BYTES = 2
var VALUE_SIZE_BYTES = 4

func PutKey(keyToByteOffset map[string]int64, file *os.File, key, value string) error {
	if len(key) > math.MaxUint16 {
		return errors.New("key size exceeds maximum for 2^16 (65535 bytes)")
	}
	if len(value) > math.MaxUint32 {
		return errors.New("value size exceeds maximum 2^32 (4GB)")
	}

	buffer := make([]byte, 6+len(key)+len(value)) // 2 (key) + 4 (value) + key/value contents' size
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
	return nil
}

func GetKey(keyToByteOffset map[string]int64, file *os.File, key string) (string, error) {
	offset, ok := keyToByteOffset[key]
	if !ok {
		return "", errors.New("Value not in memory")
	}

	// whence -> from where do i start reading/writing bytes?
	_, err := file.Seek(offset, io.SeekStart)
	if err != nil {
		log.Fatalf("Failed to seek to offset: %v", err)
	}

	keySizeBuffer := make([]byte, KEY_SIZE_BYTES) // 2 bytes -> uint16
	_, err = file.Read(keySizeBuffer)
	if err != nil {
		log.Fatalf("Failed to read key size: %v", err)
	}

	keySize := binary.BigEndian.Uint16(keySizeBuffer) // bytes -> big endian encoding

	fmt.Printf("keySize=%d\n", keySize)

	valueSizeBuffer := make([]byte, VALUE_SIZE_BYTES)
	_, err = file.Read(valueSizeBuffer)
	if err != nil {
		log.Fatalf("Failed to read content size: %v", err)
	}

	valueSize := binary.BigEndian.Uint32(valueSizeBuffer)

	fmt.Printf("valueSize=%d\n", valueSize)

	keyReadBuffer := make([]byte, keySize)
	_, err = file.Read(keyReadBuffer)
	if err != nil {
		log.Fatalf("Failed to read key: %v", err)
	}
	fmt.Printf("key=%s\n", keyReadBuffer)

	valueReadBuffer := make([]byte, valueSize)
	_, err = file.Read(valueReadBuffer)
	if err != nil {
		log.Fatalf("Failed to read value: %v", err)
	}
	fmt.Printf("key=%s\n", valueReadBuffer)

	return "", nil
}

func initFile(filePath string) {
	file, err := os.Create(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
}

func initMap() {
	fmt.Println("Initialization of map with existing memory.log file...")
}

func main() {
	keyToByteOffset := make(map[string]int64)

	filePath := "memory.log"
	file, err := os.OpenFile(filePath, os.O_APPEND, 0644)
	if err == nil {
		initMap()
		file.Close()
	} else {
		initFile(filePath)
	}

	file, err = os.OpenFile(filePath, os.O_APPEND, 0644)
	PutKey(keyToByteOffset, file, "100", "{ value1: 'value1', value2: 'value2' }")
	PutKey(keyToByteOffset, file, "abcd-efgh-jklm-pqrs", "{ userId: 100 }")

	val, err := GetKey(keyToByteOffset, file, "abcd-efgh-jklm-pqrs")
	fmt.Printf("%s\n", val)

	val, err = GetKey(keyToByteOffset, file, "100")
	fmt.Printf("%s\n", val)
}
