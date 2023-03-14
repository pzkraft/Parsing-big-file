package main

import (
	"STEPparse_ver_beta/arangodb"
	"STEPparse_ver_beta/files"
	"STEPparse_ver_beta/parser"
	"context"
	"fmt"
	"sync"
)

func main() {
	//Open file
	CADModel, err := files.Openfile("00_RAPTOR_2 v49.step")
	if err != nil {
		fmt.Printf("fail to open the file: %v \n", err)
	}
	defer CADModel.Close()

	// Make chunks
	work := files.SplitFile(CADModel, 5)

	// Make DB
	vertexCollection, edgeCollection, err := arangodb.MakeDB("Test", "TestGraph", "TestVertices", "TestEdges")
	if err != nil {
		fmt.Printf("fail to make DB: %v \n", err)
	}

	// Parallel parse and write to DB
	var wg sync.WaitGroup
	var edgesFull []parser.Edge

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for chunk := range work {
				edges, err := parser.ParseSTEP(chunk, vertexCollection)
				if err != nil {
					fmt.Printf("fail to parse or write to DB: %v \n", err)
				}
				edgesFull = append(edgesFull, edges...)

			}
		}()
	}
	wg.Wait()
	for _, v := range edgesFull {
		_, err := edgeCollection.CreateDocument(context.TODO(), v)
		if err != nil {
			fmt.Printf("fail to create edge documents: %v", err)
		}
	}
}
