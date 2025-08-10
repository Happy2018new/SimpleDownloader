package raknet

import (
	"context"

	"github.com/gorilla/websocket"
)

// 初始化一个空的 Websocket 客户端。
//
// decodePacket 用于解码数据包，
// 而 encodePacket 则用于编码数据包
func NewWebsocket(
	decodePacket func(buf []byte) (pk Packet),
	encodePacket func(pk Packet) (buf []byte),
) *Raknet {
	ctx, cancel := context.WithCancel(context.Background())
	return &Raknet{
		context:      ctx,
		cancel:       cancel,
		decodePacket: decodePacket,
		encodePacket: encodePacket,
	}
}

// 将底层 Raknet 连接设置为 connection
func (r *Raknet) SetConnection(connection *websocket.Conn) {
	r.connection = connection
	r.packet = make(chan Packet, 512)
}

// 关闭已建立的 Raknet 底层连接
func (r *Raknet) CloseConnection() {
	r.closedLock.Lock()
	defer r.closedLock.Unlock()

	r.cancel()

	if r.connection != nil {
		r.connection.Close()
	}

	if !r.closed && r.packet != nil {
		close(r.packet)
		r.closed = true
	}
}

// 获取当前的上下文
func (r *Raknet) GetContext() context.Context {
	return r.context
}

/*
从底层 Raknet 不断地读取多个数据包，
直到底层 Raknet 连接被关闭。

在大多数情况下，由于我们只需按原样传递数据包，
因此，我们只解码了一部分必须的数据包。

而对于其他的数据包，我们不作额外处理，
而是仅仅地保留它们的二进制负载

另，此函数应当只被调用一次
*/
func (r *Raknet) ProcessIncomingPackets() {
	// 确保该函数不会返回恐慌
	defer func() {
		recover()
	}()
	// 不断处理到来的一个或多个数据包
	for {
		// 从底层 Raknet 连接读取数据包
		_, message, err := r.connection.ReadMessage()
		// 从底层 Raknet 连接读取数据包
		if err != nil {
			// 此时从底层 Raknet 连接读取数据包遭遇了错误，
			// 因此我们认为连接已被关闭
			r.CloseConnection()
			return
		}
		// 解码并提交数据包
		select {
		case <-r.context.Done():
			r.CloseConnection()
			return
		default:
			r.packet <- r.decodePacket(message)
		}
	}
}

// 从已读取且已解码的数据包池中读取单个数据包。
// 当数据包池没有数据包时，将会阻塞，
// 直到新的已处理数据包抵达
func (r *Raknet) ReadPacket() (pk Packet, connClosed bool) {
	pk = <-r.packet
	select {
	case <-r.context.Done():
		connClosed = true
	default:
	}
	return
}

// 向底层 Raknet 连接写单个数据包 pk
func (r *Raknet) WritePacket(pk Packet) {
	// 将数据包写入底层 Raknet 连接
	err := r.connection.WriteMessage(websocket.BinaryMessage, r.encodePacket(pk))
	// 此时向底层 Raknet 连接写入数据包遭遇了错误，
	// 因此我们认为连接已被关闭
	if err != nil {
		r.CloseConnection()
	}
}
