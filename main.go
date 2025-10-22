package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	// Define command-line flags
	testType := flag.String("test", "inmemory", "Type of test to run: 'inmemory' or 'mongodb'")
	flag.Parse()

	fmt.Println("Scheduler Test Runner")
	fmt.Println("=====================")
	fmt.Printf("Running test: %s\n\n", *testType)

	switch *testType {
	case "inmemory", "memory", "mem":
		runInMemoryTest()
	case "mongodb", "mongo", "db":
		runMongoDBTest()
	default:
		fmt.Printf("Unknown test type: %s\n\n", *testType)
		fmt.Println("Usage:")
		fmt.Println("  go run . -test=inmemory    # Run in-memory store test")
		fmt.Println("  go run . -test=mongodb     # Run MongoDB store test")
		fmt.Println()
		fmt.Println("Aliases:")
		fmt.Println("  inmemory: memory, mem")
		fmt.Println("  mongodb:  mongo, db")
		os.Exit(1)
	}
}
