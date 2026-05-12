package main

import (
	"fmt"
	// "io"
	"encoding/binary"
	"errors"
	"log"
	"math"
	"os"
)

func putKey(keyToByteOffset map[string]int, file *os.File, key, value string) error {
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

	return nil
}

func initFile(filePath string) {
	file, err := os.Create(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	file.WriteString("key,value\n")
}

func initMap() {
	fmt.Println("Initialization of map with existing memory.csv file...")
}

func main() {
	keyToByteOffset := make(map[string]int)

	filePath := "memory.csv"
	file, err := os.OpenFile(filePath, os.O_APPEND, 0644)
	if err == nil {
		initMap()
		file.Close()
	} else {
		initFile(filePath)
	}

	file, err = os.OpenFile(filePath, os.O_APPEND, 0644)
	putKey(keyToByteOffset, file, "10", "{ value1: 'value1', value2: 'value2}")
}
