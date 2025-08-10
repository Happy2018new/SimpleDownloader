package main

import (
	"fmt"
	"os"

	"simple_downloader/protocol"
	customRaknet "simple_downloader/raknet"
	"simple_downloader/utils"

	"github.com/pterm/pterm"
	"github.com/sandertv/go-raknet"
)

func main() {
	var currentOffset int

	// Get file path
	filePath := utils.ReadStringFromPanel("Type the file path: ")
	file, err := os.OpenFile(filePath, os.O_CREATE, 0600)
	if err != nil {
		panic(fmt.Sprintf("Open file %s error: %v", filePath, err))
	}
	defer file.Close()

	listener, err := raknet.Listen(":19132")
	if err != nil {
		panic(fmt.Sprintf("Failed to create listener due to %v", err))
	}
	pterm.Success.Println("Client is listen on 127.0.0.1:19132")

	for {
		incomingConn, err := listener.Accept()
		if err != nil {
			pterm.Warning.Printfln("Listener accept connect failed: %v", err)
			continue
		}

		conn := customRaknet.NewRaknet(customRaknet.DecodePacket, customRaknet.EncodePacket)
		conn.SetConnection(incomingConn)
		go conn.ProcessIncomingPackets()

		// Get file size
		conn.WriteSinglePacket(&protocol.FileSizeRequest{})
		pks := conn.ReadPackets()
		if len(pks) != 1 {
			continue
		}
		fileSize := pks[0].(*protocol.FileSizeResponse).Size

		// File seek
		conn.WriteSinglePacket(&protocol.FileSeekRequest{Offset: int64(currentOffset)})
		pks = conn.ReadPackets()
		if len(pks) != 1 {
			continue
		}
		if seekResp := pks[0].(*protocol.FileSeekResponse); !seekResp.Success {
			panic(fmt.Sprintf("File seek error: %v", seekResp.ErrorInfo))
		}

		for {
			conn.WriteSinglePacket(&protocol.FileChunkRequest{})

			pks = conn.ReadPackets()
			if len(pks) != 1 {
				break
			}
			pk := pks[0].(*protocol.FileChunkResponse)

			_, err = file.Write(pk.ChunkData)
			if err != nil {
				panic(fmt.Sprintf("Write file error: %v", err))
			}
			currentOffset += len(pk.ChunkData)

			if pk.IsFinalChunk {
				conn.CloseConnection()
				pterm.Success.Printfln("Success to download data to %s", filePath)
				return
			}

			pterm.Info.Printfln("Receive data: %d/%d", currentOffset, fileSize)
		}
	}
}
