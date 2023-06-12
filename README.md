notes for me:

F10 - Bios
start penguin WSL from start menu:
    >sudo service docker start
    >docker run hello-world
    >docker run -p 8529:8529 -e ARANGO_ROOT_PASSWORD=openSesame arangodb/arangodb:3.10.2
then VSCode go run .