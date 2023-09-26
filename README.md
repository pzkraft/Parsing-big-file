This program parses STEP file with several goroutines and upload data to ArangoDB Graph.

Algorithm:
- Input file separates to chunks
- One goroutine takes one chunk, catch certain elements in the chunk and upload data to ArangoDB Collection of nodes. Also gouroutine makes 2 slices to create a collection of edges later
- When all node elements are uploaded creates edge collection 

TODO:
- move creating of edge collection into gouroutines(as much as possible)
