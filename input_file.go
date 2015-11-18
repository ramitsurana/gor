package main

import (
	"bufio"
	"log"
	"os"
	"strconv"
	"time"
)

// FileInput can read requests generated by FileOutput
type FileInput struct {
	data        chan []byte
	path        string
	file        *os.File
	speedFactor float64
}

// NewFileInput constructor for FileInput. Accepts file path as argument.
func NewFileInput(path string) (i *FileInput) {
	i = new(FileInput)
	i.data = make(chan []byte)
	i.path = path
	i.speedFactor = 1
	i.init(path)

	go i.emit()

	return
}

func (i *FileInput) init(path string) {
	file, err := os.Open(path)

	if err != nil {
		log.Fatal(i, "Cannot open file %q. Error: %s", path, err)
	}

	i.file = file
}

func (i *FileInput) Read(data []byte) (int, error) {
	buf := <-i.data
	copy(data, buf)

	return len(buf), nil
}

func (i *FileInput) String() string {
	return "File input: " + i.path
}

func (i *FileInput) emit() {
	var lastTime int64

	// reader := bufio.NewReader(conn)
	scanner := bufio.NewScanner(i.file)
	scanner.Split(payloadScanner)

	for scanner.Scan() {
		buf := scanner.Bytes()
		meta := payloadMeta(buf)

		if len(meta) > 2 && meta[0][0] == RequestPayload {
			ts, _ := strconv.ParseInt(string(meta[2]), 10, 64)

			if lastTime != 0 {
				timeDiff := ts - lastTime

				if i.speedFactor != 1 {
					timeDiff = int64(float64(timeDiff) / i.speedFactor)
				}

				time.Sleep(time.Duration(timeDiff))
			}

			lastTime = ts
		}

		// scanner returs only pointer, so to remove data-race we have to allocate new array
		newBuf := make([]byte, len(buf))
		copy(newBuf, buf)

		i.data <- newBuf
	}

	log.Println("FileInput: end of file")
}
