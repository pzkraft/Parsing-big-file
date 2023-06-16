package main

import (
	"STEPparse_ver_beta/arangodb"
	"STEPparse_ver_beta/files"
	"STEPparse_ver_beta/parser"
	"context"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"time"
)

func main() {
	// Open file
	CADModel, err := files.Openfile("00_RAPTOR_2 v49.step")
	if err != nil {
		fmt.Printf("fail to open the file: %v \n", err)
	}
	defer CADModel.Close()

	// Make DB
	vertexCollection, edgeCollection, err := arangodb.MakeDB("Test", "TestGraph", "TestVertices", "TestEdges")
	if err != nil {
		fmt.Printf("fail to make DB: %v \n", err)
	}

	// Make chunks
	work := make(chan string)
	files.SplitFileToChunks(work, CADModel, 5)

	// Parallel parsing and making of vertexCollection and slices of edges and parts
	var wg sync.WaitGroup
	var edgesFull []parser.Edge
	var partsFull []parser.Vertex

	start := time.Now()

	for i := 0; i < 5; i++ {
		wg.Add(1)
		fmt.Println("add Goroutine number:", i+1)
		go func() {
			defer wg.Done()
			for chunk := range work {
				fmt.Println("+++start:", strings.Trim(chunk[:15], "\r\n"))
				edges, parts, err := parser.ParseFile(chunk, vertexCollection)
				if err != nil {
					fmt.Printf("fail to parse or write to DB: %v \n", err)
				}
				edgesFull = append(edgesFull, edges...)
				partsFull = append(partsFull, parts...)
				fmt.Println("---end:", strings.Trim(chunk[:15], "\r\n"))

			}
		}()
	}
	fmt.Println("The number of active Goroutines:", runtime.NumGoroutine())
	wg.Wait()
	duration := time.Since(start)

	//catching head parts without 'parent'
	err = parser.CatchHeadParts(partsFull, edgesFull, vertexCollection)
	if err != nil {
		fmt.Printf("fail to catch part without child: %v", err)
	}

	// making of edgeCollection
	for _, oneEdge := range edgesFull {
		_, err := edgeCollection.CreateDocument(context.TODO(), oneEdge)
		if err != nil {
			fmt.Printf("fail to create edge documents: %v", err)
		}
	}
	fmt.Println("The number of active Goroutines before the end:", runtime.NumGoroutine())
	fmt.Println("Execution time of goroutine's loop:", duration)
}
