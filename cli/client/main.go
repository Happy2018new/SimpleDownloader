package main

import (
	"fmt"
	"log"
	"os"

	"simple_downloader/protocol"
	"simple_downloader/utils"
	customWebsocket "simple_downloader/websocket"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/pterm/pterm"
)

var (
	downloadFinished bool   = false
	currentOffset    int64  = 0
	filePath         string = ""
)

func handler(c *gin.Context) {
	websocketConn, err := new(websocket.Upgrader).Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	defer func() {
		websocketConn.Close()
		if downloadFinished {
			os.Exit(0)
		}
	}()

	for {
		// Create wrapper
		conn := customWebsocket.NewWebsocket(customWebsocket.DecodePacket, customWebsocket.EncodePacket)
		conn.SetConnection(websocketConn)
		go conn.ProcessIncomingPackets()

		// Get file size
		conn.WritePacket(&protocol.FileSizeRequest{})
		pk, connClosed := conn.ReadPacket()
		if connClosed {
			return
		}
		fileSize := pk.(*protocol.FileSizeResponse).Size

		// File seek
		conn.WritePacket(&protocol.FileSeekRequest{Offset: currentOffset})
		pk, connClosed = conn.ReadPacket()
		if connClosed {
			return
		}
		if _, ok := pk.(*protocol.FileSeekResponse); !ok {
			panic(fmt.Sprintf("File seek error: Returned packet is not file seek response; pk = %#v", pk))
		}

		for {
			conn.WritePacket(&protocol.FileChunkRequest{})

			pk, connClosed := conn.ReadPacket()
			if connClosed {
				return
			}
			p := pk.(*protocol.FileChunkResponse)

			file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0600)
			if err != nil {
				panic(fmt.Sprintf("Open file error: %v", err))
			}
			_, err = file.WriteAt(p.ChunkData, currentOffset)
			if err != nil {
				panic(fmt.Sprintf("Write file error: %v", err))
			}
			err = file.Close()
			if err != nil {
				panic(fmt.Sprintf("Close file error: %v", err))
			}
			currentOffset += int64(len(p.ChunkData))

			if p.IsFinalChunk {
				downloadFinished = true
				conn.CloseConnection()
				pterm.Success.Printfln("Success to download data to %s", filePath)
				return
			}

			resultStr := fmt.Sprintf(
				"Receive data: %d/%d (%v",
				currentOffset, fileSize,
				float64(currentOffset)/float64(fileSize+1)*100,
			)
			pterm.Info.Println(resultStr + "%)")
		}
	}
}

func main() {
	filePath = utils.ReadStringFromPanel("Type the file path: ")
	_ = os.Remove(filePath)

	router := gin.Default()
	router.Any("/", handler)
	router.Run(":2018")
}
