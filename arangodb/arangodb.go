package arangodb

import (
	"context"
	"fmt"

	driver "github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/http"
)

func MakeDB(DBname string, nameOfGraph string, nameOfVerticeCollection string, nameOfEdgeCollection string) (driver.Collection, driver.Collection, error) {
	// Create an HTTP connection to the database
	conn, err := http.NewConnection(http.ConnectionConfig{
		Endpoints: []string{"http://localhost:8529"},
	})
	if err != nil {
		return nil, nil, fmt.Errorf("fail to create HTTP connection: %w", err)
	}

	// Create a client
	c, err := driver.NewClient(driver.ClientConfig{
		Connection:     conn,
		Authentication: driver.BasicAuthentication("root", "openSesame"),
	})
	if err != nil {
		return nil, nil, fmt.Errorf("fail to authentication: %w", err)
	}

	// Create database
	db, err := c.CreateDatabase(context.TODO(), DBname, nil)
	if driver.IsArangoErrorWithCode(err, 409) {
		for i := 1; err != nil && i < 50; i++ {
			DBname = DBname + "1"
			db, err = c.CreateDatabase(context.TODO(), DBname, nil)
			if i == 49 {
				return nil, nil, fmt.Errorf("DBname reached the limit, %w", err)
			}
		}
	} else if err != nil {
		return nil, nil, fmt.Errorf("fail to create database %w", err)
	}

	// define the edgeCollection to store the edges
	var edgeDefinition driver.EdgeDefinition
	edgeDefinition.Collection = nameOfEdgeCollection
	// define a set of collections where an edge is going out...
	edgeDefinition.From = []string{nameOfVerticeCollection}

	// repeat this for the collections where an edge is going into
	edgeDefinition.To = []string{nameOfVerticeCollection}

	// A graph can contain additional vertex collections, defined in the set of orphan collections
	var options driver.CreateGraphOptions
	// options.OrphanVertexCollections = []string{"myCollection4", "myCollection5"}
	options.EdgeDefinitions = []driver.EdgeDefinition{edgeDefinition}

	// now it's possible to create a graph
	graph, err := db.CreateGraphV2(context.TODO(), nameOfGraph, &options)
	if err != nil {
		return nil, nil, fmt.Errorf("fail to create graph: %w", err)
	}

	vertexCollection, err := graph.VertexCollection(context.TODO(), nameOfVerticeCollection)
	if err != nil {
		return nil, nil, fmt.Errorf("fail to get vertex collection: %w", err)
	}

	edgeCollection, _, err := graph.EdgeCollection(context.TODO(), nameOfEdgeCollection)
	if err != nil {
		return nil, nil, fmt.Errorf("fail to get Edges collection: %w", err)
	}
	return vertexCollection, edgeCollection, err
}
