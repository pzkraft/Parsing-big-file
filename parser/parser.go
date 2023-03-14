package parser

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/arangodb/go-driver"
)

// Edge and Vertex structures consist of mandatory information about each part.
type Edge struct {
	ID_from     string `json:"_from"`
	ParentID_to string `json:"_to"`
}

type Vertex struct {
	Name string `json:"part"`
	ID   string `json:"_key"`
}

// Forms Vertex and Edge structs from the fileline using regular expressions(regexp).
// Input param is "line" - fileline, "variant" == VE(Vertex+Edge) if it has lines with edges or V(Vertex) if without edges
// Output param is a filled structures
func ProcessVertexEdge(line string, variant string) (Vertex, Edge) {
	Vertex := Vertex{}
	Edge := Edge{}

	if variant == "VE" {
		reName := regexp.MustCompile(`\'.*?\:`)
		reParentID := regexp.MustCompile(`\#.*?\,`)
		reID := regexp.MustCompile(`\#.*?\,`)

		Vertex.Name = strings.Trim(reName.FindString(line), "':")
		Edge.ParentID_to = "TestVertices" + "/" + strings.Trim(reParentID.FindString(line[3:]), "#,")
		i := strings.LastIndex(line, "#")
		Vertex.ID = strings.Trim(reID.FindString(line[i:]), "#,")
		Edge.ID_from = "TestVertices" + "/" + Vertex.ID

	} else if variant == "V" {
		reName := regexp.MustCompile(`(\,'|\n')(.*?)\'`)
		reID := regexp.MustCompile(`\#.*?\=`)

		Vertex.Name = strings.Trim(reName.FindString(line), "\n',")
		Vertex.ID = strings.Trim(reID.FindString(line), "#=")

	}
	return Vertex, Edge
}

// Parsing Tree of parts for a model in STEP file
func ParseSTEP(chunk string, vertexCollection driver.Collection) ([]Edge, error) {
	flag := 0
	var edges []Edge

	for _, line := range strings.Split(chunk, ";") {

		// Parsing filelines with links between parts
		if strings.Contains(line, "NEXT_ASSEMBLY_USAGE_OCCURRENCE") {

			vertex, edge := ProcessVertexEdge(line, "VE")

			_, err := vertexCollection.CreateDocument(context.TODO(), vertex)
			if driver.IsArangoError(err) {
				for i := 1; driver.IsArangoErrorWithCode(err, 409) && i < 100; i++ {
					vertex.ID = vertex.ID + "(" + strconv.Itoa(i) + ")"
					_, err = vertexCollection.CreateDocument(context.TODO(), vertex)
					edge.ID_from = "TestVertices" + "/" + vertex.ID
					vertex.ID = vertex.ID[:len(vertex.ID)-3]
					if i == 99 {
						return nil, fmt.Errorf("counter of identical ID  reached the limit, %w", err)
					}
				}
			} else if err != nil {
				return nil, fmt.Errorf("fail to create vertex documents: %w", err)
			}
			edges = append(edges, edge)
			continue
		}

		// Parsing filelines without "parent"
		if strings.Contains(line, "PRODUCT_DEFINITION(") {

			if flag == 0 {
				time.Sleep(time.Second)
				flag = 1
			}

			vertex, _ := ProcessVertexEdge(line, "V")

			_, err := vertexCollection.CreateDocument(context.TODO(), vertex)
			if driver.IsArangoErrorWithCode(err, 409) {
				continue
			} else if err != nil {
				return nil, fmt.Errorf("fail to create vertex documents: %w", err)
			}
		}
	}
	return edges, nil
}
