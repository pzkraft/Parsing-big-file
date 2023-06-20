This program parses STEP file with several goroutines and upload data to ArangoDB Graph.
Algorithm:
- Input file separates to chunks
- One goroutine takes one chunk, catch certain elements in the chunk and upload data to ArangoDB Collection of nodes. Also gouroutine makes 2 slices to create a collection of edges later
- When all node elements are uploaded creates edge collection 

TODO:
- move creating of edge collection into gouroutines(as much as possible)


notes for me:

F10 - Bios
start penguin WSL from start menu:
    >sudo service docker start
    >docker run hello-world
    >docker run -p 8529:8529 -e ARANGO_ROOT_PASSWORD=openSesame arangodb/arangodb:3.10.2
then VSCode go run .