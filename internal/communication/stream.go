package communication

import (
	"bytes"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
	"github.com/libp2p/go-libp2p/core/network"
)

const (
	ChunkSize    = 1024
	Timeout      = 5 * time.Second
	ReadDeadline = 10 * time.Second
)

type ResponseData struct {
	Id   string
	Data []byte
	Err  error
}

func WriteData(stream network.Stream, data []byte) error {

	chunks := splitIntoChunks(data, ChunkSize)

	for _, chunk := range chunks {
		_, err := stream.Write(chunk)
		if err != nil {
			return fmt.Errorf("error writing data to stream: %w", err)
		}
	}

	_, err := stream.Write([]byte("END_OF_TRANSMISSION"))
	if err != nil {
		return fmt.Errorf("error writing end signal to stream: %w", err)
	}

	return nil
}

func splitIntoChunks(data []byte, chunkSize int) [][]byte {
	var chunks [][]byte

	for i := 0; i < len(data); i += chunkSize {
		end := i + chunkSize
		if end > len(data) {
			end = len(data)
		}
		chunks = append(chunks, data[i:end])
	}

	return chunks
}

func HearStream(stream network.Stream, channelResponse chan<- ResponseData) {
	var responseData ResponseData

	var buf bytes.Buffer
	buffer := make([]byte, ChunkSize)
	for {
		bytesRead, err := stream.Read(buffer)
		if err != nil && err != io.EOF {
			responseData.Err = fmt.Errorf("sendRequest: failed to read response data: %w", err)
			fmt.Println("sendRequest: failed to read response data: %w", err)
			return
		}

		buf.Write(buffer[:bytesRead])

		data := buf.Bytes()

		endIndex := bytes.Index(data, []byte("END_OF_TRANSMISSION"))
		if endIndex != -1 {
			data = data[:endIndex]

			responseData.Data = data
			responseData.Id = uuid.New().String()

			channelResponse <- responseData

			buf.Reset()
			if endIndex+18 < len(data) {
				buf.Write(data[endIndex+18:])
			}
		}
	}
}
