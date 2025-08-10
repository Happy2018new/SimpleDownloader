package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"simple_downloader/protocol"
	"simple_downloader/utils"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pterm/pterm"

	customWebsocket "simple_downloader/websocket"
)

func main() {
	var currentOffset int64

	// Get file path
	filePath := utils.ReadStringFromPanel("Type the file path: ")
	file, err := os.OpenFile(filePath, os.O_RDONLY, 0600)
	if err != nil {
		panic(fmt.Sprintf("Open file %s error: %v", filePath, err))
	}

	// Get file info
	fileInfo, err := file.Stat()
	if err != nil {
		panic(fmt.Sprintf("Get file info %s error: %v", filePath, err))
	}
	fileSize := fileInfo.Size()

	// Get client address
	clientAddress := utils.ReadStringFromPanel("Type client address: ")
	if len(clientAddress) == 0 {
		panic("Given client address is not valid (the length of it must bigger than 0)")
	}

	// Connect to client so we can passing data
	for {
		conn := ConnectToClient(clientAddress)

		for {
			pk, connClosed := conn.ReadPacket()
			if connClosed {
				break
			}
			switch p := pk.(type) {
			case *protocol.FileSeekRequest:
				currentOffset = p.Offset
				conn.WritePacket(&protocol.FileSeekResponse{})
			case *protocol.FileChunkRequest:
				// Read data (3MB per request)
				result := make([]byte, 3145728)
				validLength, err := file.ReadAt(result, currentOffset)
				// Change pointer
				isFinalChunk := errors.Is(err, io.EOF)
				currentOffset += int64(validLength)
				// Write packet
				conn.WritePacket(&protocol.FileChunkResponse{
					IsFinalChunk: isFinalChunk,
					ChunkData:    result[:validLength],
				})
				// If EOF, then return
				if isFinalChunk {
					conn.CloseConnection()
					pterm.Success.Println("File download finished")
					return
				}
			case *protocol.FileSizeRequest:
				conn.WritePacket(&protocol.FileSizeResponse{
					Size: fileSize,
				})
			}
		}
	}
}

// ConnectToClient ..
func ConnectToClient(address string) *customWebsocket.Websocket {
	var conn *websocket.Conn
	var err error

	for {
		ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*5)
		conn, _, err = websocket.DefaultDialer.DialContext(ctx, address, http.Header{})
		if err != nil {
			pterm.Warning.Printfln("Faild to connect to %s due to %v, retry.", address, err)
			cancelFunc()
			continue
		}
		cancelFunc()
		break
	}

	wrapper := customWebsocket.NewWebsocket(customWebsocket.DecodePacket, customWebsocket.EncodePacket)
	wrapper.SetConnection(conn)
	go wrapper.ProcessIncomingPackets()

	return wrapper
}
