package main

import (
	"bytes"
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
	downloadFinished bool          = false
	fileBuf          *bytes.Buffer = bytes.NewBuffer(nil)
	filePath         string
)

func handler(c *gin.Context) {
	websocketConn, err := new(websocket.Upgrader).Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	defer func() {
		_ = websocketConn.Close()
		if err = os.WriteFile(filePath, fileBuf.Bytes(), 0600); err != nil {
			panic(fmt.Sprintf("Write file error: %v", err))
		}
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
		conn.WritePacket(&protocol.FileSeekRequest{Offset: int64(fileBuf.Len())})
		pk, connClosed = conn.ReadPacket()
		if connClosed {
			return
		}
		if seekResp := pk.(*protocol.FileSeekResponse); !seekResp.Success {
			panic(fmt.Sprintf("File seek error: %v", seekResp.ErrorInfo))
		}

		for {
			conn.WritePacket(&protocol.FileChunkRequest{})

			pk, connClosed := conn.ReadPacket()
			if connClosed {
				return
			}
			p := pk.(*protocol.FileChunkResponse)

			_, err := fileBuf.Write(p.ChunkData)
			if err != nil {
				panic(fmt.Sprintf("Write buf error: %v", err))
			}

			if p.IsFinalChunk {
				downloadFinished = true
				conn.CloseConnection()
				pterm.Success.Printfln("Success to download data to %s", filePath)
				return
			}

			resultStr := fmt.Sprintf(
				"Receive data: %d/%d (%v",
				fileBuf.Len(), fileSize,
				float64(fileBuf.Len())/float64(fileSize+1)*100,
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
