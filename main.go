package main

import (
	"fmt"
	"log"
	"os"
)

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
	// keyToByteOffset := make(map[string]int64)

	filePath := "memory.log"
	file, err := os.OpenFile(filePath, os.O_APPEND, 0644)
	if err == nil {
		initMap()
		file.Close()
	} else {
		initFile(filePath)
	}

	// file, err = os.OpenFile(filePath, os.O_APPEND, 0644)
	// PutKey(keyToByteOffset, file, "100", "{ value1: 'value1', value2: 'value2' }")
	// PutKey(keyToByteOffset, file, "abcd-efgh-jklm-pqrs", "{ userId: 100 }")

	// val, err := GetKey(keyToByteOffset, file, "abcd-efgh-jklm-pqrs")
	// fmt.Printf("%s\n", val)

	// val, err = GetKey(keyToByteOffset, file, "100")
	// fmt.Printf("%s\n", val)
}
