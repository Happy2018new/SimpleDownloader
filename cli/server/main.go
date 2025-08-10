package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"simple_downloader/protocol"
	"simple_downloader/utils"
	"time"

	"github.com/pterm/pterm"
	"github.com/sandertv/go-raknet"

	customRaknet "simple_downloader/raknet"
)

func main() {
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
			var needBreak bool

			for _, pk := range conn.ReadPackets() {
				switch p := pk.(type) {
				case *protocol.FileSeekRequest:
					_ = file.Close()
					file, _ = os.OpenFile(filePath, os.O_RDONLY, 0600)
					_, err = file.Seek(2, int(p.Offset))
					conn.WriteSinglePacket(&protocol.FileSeekResponse{
						Success:   err == nil,
						ErrorInfo: fmt.Sprintf("%v", err),
					})
				case *protocol.FileChunkRequest:
					result := make([]byte, 524288)
					validLength, err := file.Read(result)
					isFinalChunk := errors.Is(err, io.EOF)
					conn.WriteSinglePacket(&protocol.FileChunkResponse{
						IsFinalChunk: isFinalChunk,
						ChunkData:    result[:validLength],
					})
					if isFinalChunk {
						conn.CloseConnection()
						pterm.Success.Println("File download finished")
						return
					}
				case *protocol.FileSizeRequest:
					conn.WriteSinglePacket(&protocol.FileSizeResponse{
						Size: fileSize,
					})
				}
			}

			select {
			case <-conn.GetContext().Done():
				needBreak = true
			default:
			}
			if needBreak {
				break
			}
		}
	}
}

// ConnectToClient ..
func ConnectToClient(address string) *customRaknet.Raknet {
	var conn *raknet.Conn
	var err error

	for {
		ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*5)
		conn, err = raknet.DialContext(ctx, address)
		if err != nil {
			pterm.Warning.Printfln("Faild to connect to %s due to %v, retry.", address, err)
			cancelFunc()
			continue
		}
		cancelFunc()
		break
	}

	wrapper := customRaknet.NewRaknet(customRaknet.DecodePacket, customRaknet.EncodePacket)
	wrapper.SetConnection(conn)
	go wrapper.ProcessIncomingPackets()

	return wrapper
}
