package files

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
)

func Openfile(fileToOpen interface{}) (*os.File, error) {
	switch v := fileToOpen.(type) {
	case string:
		if !strings.HasSuffix(strings.ToLower(v), ".step") {
			return nil, fmt.Errorf("the file isn't a '.step' one")
		}
		openedFile, err := os.Open(v)
		if err != nil {
			return nil, fmt.Errorf("unable to read file: %w", err)
		}
		return openedFile, nil
	default:
		return nil, fmt.Errorf("it's not a file")
	}
}

func SplitFile(fileToBeChunked *os.File, numOfChunks int) chan string {

	//const fileChunk = 40 * (1 << 20) // 40 MB, change this to your requirement

	fileInfo, _ := fileToBeChunked.Stat()
	var fileSize int64 = fileInfo.Size()

	fileChunk := float64(fileSize) / float64(numOfChunks)

	// calculate total number of parts the file will be chunked into
	totalPartsNum := uint64(math.Ceil(float64(fileSize) / float64(fileChunk)))

	fmt.Printf("Splitting to %d pieces.\n", totalPartsNum)

	addPartBuffer := ""
	chunk := ""
	work := make(chan string, numOfChunks)
	for i := uint64(0); i < totalPartsNum; i++ {

		partSize := int(math.Min(fileChunk, float64(fileSize-int64(i*uint64(fileChunk)))))
		partBuffer := make([]byte, partSize)

		fileToBeChunked.Read(partBuffer)

		endLine := strings.LastIndex(string(partBuffer), ";")

		chunkName := "chunk_" + strconv.FormatUint(i, 10)
		chunk = addPartBuffer + string(partBuffer)[:endLine+1]
		work <- chunk

		fmt.Println("Split to : ", chunkName)

		chunk = ""
		addPartBuffer = string(partBuffer)[endLine+1 : len(string(partBuffer))]
	}
	defer close(work)
	return work
}
