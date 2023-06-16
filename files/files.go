package files

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"sync"
)

// TODO: interface delete
func Openfile(fileToOpen string) (*os.File, error) {
	if !strings.HasSuffix(strings.ToLower(fileToOpen), ".step") {
		return nil, fmt.Errorf("the file isn't a '.step' one")
	}

	openedFile, err := os.Open(fileToOpen)
	if err != nil {
		return nil, fmt.Errorf("unable to read file: %w", err)
	}
	return openedFile, nil

}

func SplitFileToChunks(workChan chan string, fileToBeChunked *os.File, numOfChunks int) {

	fileInfo, _ := fileToBeChunked.Stat()
	var fileSize int64 = fileInfo.Size()

	fileChunk := float64(fileSize) / float64(numOfChunks)
	//fileChunk := 40 * (1 << 20) // 40 MB, change this to your requirement

	// calculate total number of parts the file will be chunked into
	totalPartsNum := uint64(math.Ceil(float64(fileSize) / float64(fileChunk)))
	fmt.Printf("add Goroutine number: 0 --- Splitting to %d pieces.\n", totalPartsNum)

	//TODO: for in goroutine
	var wg sync.WaitGroup
	wg.Add(1)
	go func(workChan chan string) {
		defer wg.Done()
		addPartBuffer := ""
		chunk := ""
		for i := uint64(0); i < totalPartsNum; i++ {

			partSize := int(math.Min(fileChunk, float64(fileSize-int64(i*uint64(fileChunk)))))
			partBuffer := make([]byte, partSize)

			fileToBeChunked.Read(partBuffer)

			endLine := strings.LastIndex(string(partBuffer), ";")

			chunkName := "chunk_" + strconv.FormatUint(i, 10)
			chunk = addPartBuffer + string(partBuffer)[:endLine+1]
			workChan <- chunk

			fmt.Println("Split to : ", chunkName, "starts at:", strings.Trim(chunk[:15], "\r\n"))

			chunk = ""
			addPartBuffer = string(partBuffer)[endLine+1 : len(string(partBuffer))]
		}
		defer close(workChan)
	}(workChan)
}
