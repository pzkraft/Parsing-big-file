package parser

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

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

	if variant == "NEXT_ASSEMBLY" {
		reName := regexp.MustCompile(`\'.*?\:`)
		reParentID := regexp.MustCompile(`\#.*?\,`)
		reID := regexp.MustCompile(`\#.*?\,`)

		Vertex.Name = strings.Trim(reName.FindString(line), "':")
		Edge.ParentID_to = "TestVertices" + "/" + strings.Trim(reParentID.FindString(line[3:]), "#,")
		i := strings.LastIndex(line, "#")
		Vertex.ID = strings.Trim(reID.FindString(line[i:]), "#,")
		Edge.ID_from = "TestVertices" + "/" + Vertex.ID

	} else if variant == "PRODUCT_" {
		reName := regexp.MustCompile(`(\,'|\n')(.*?)\'`)
		reID := regexp.MustCompile(`\#.*?\=`)

		Vertex.Name = strings.Trim(reName.FindString(line), "\n',")
		Vertex.ID = strings.Trim(reID.FindString(line), "#=")

	}
	return Vertex, Edge
}

// Parsing Tree of parts for a model in STEP file
func ParseFile(chunk string, vertexCollection driver.Collection) ([]Edge, []Vertex, error) {
	var edges []Edge
	var parts []Vertex

	for _, line := range strings.Split(chunk, ";") {

		// Parsing filelines with links between parts
		if strings.Contains(line, "NEXT_ASSEMBLY_USAGE_OCCURRENCE") {

			part, edge := ProcessVertexEdge(line, "NEXT_ASSEMBLY")

			_, err := vertexCollection.CreateDocument(context.TODO(), part)
			if driver.IsArangoError(err) {
				for i := 1; driver.IsArangoErrorWithCode(err, 409); i++ {
					part.ID = part.ID + "(" + strconv.Itoa(i) + ")"
					_, err = vertexCollection.CreateDocument(context.TODO(), part)
					edge.ID_from = "TestVertices" + "/" + part.ID
					part.ID = part.ID[:len(part.ID)-3]
					if i == 99 {
						return nil, nil, fmt.Errorf("counter of identical ID reached the limit, %w", err)
					}
				}
			} else if err != nil {
				return nil, nil, fmt.Errorf("fail to create vertex(part) document: %w", err)
			}
			edges = append(edges, edge)
		}

		// Parsing filelines without "parent"
		if strings.Contains(line, "PRODUCT_DEFINITION(") {

			part, _ := ProcessVertexEdge(line, "PRODUCT_")
			parts = append(parts, part)

		}
	}
	return edges, parts, nil
}

func CatchHeadParts(parts []Vertex, edges []Edge, vertexCollection driver.Collection) error {
	re := regexp.MustCompile("[0-9]+")
	for _, onePart := range parts {
		hasChild := false
		for i := 0; i < len(edges); i++ {
			if onePart.ID == re.FindString(edges[i].ID_from) {
				hasChild = true
				break
			}
		}
		if !hasChild {
			_, err := vertexCollection.CreateDocument(context.TODO(), onePart)
			if err != nil {
				return fmt.Errorf("fail to create vertex document from partsFull: %w", err)
			}
		}
	}
	return nil
}
