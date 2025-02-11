package znet

import (
	"errors"
	"fmt"
	"io"
	"net"
	"zinx/ziface"
)

type Connection struct {
	Conn         *net.TCPConn      //当前连接的socket tcp套接字
	ConnID       uint32            //当前连接的ID也可以称为SessionID, ID全局唯一
	isClosed     bool              //当前连接的关闭状态
	msgHandler   ziface.IMsgHandle //该连接的处理方法
	ExitBuffChan chan bool         //告知该连接已经退出/停止的channel
}

func NewConnection(conn *net.TCPConn, connID uint32, msgHandler ziface.IMsgHandle) *Connection {
	c := &Connection{
		Conn:         conn,
		ConnID:       connID,
		isClosed:     false,
		msgHandler:   msgHandler,
		ExitBuffChan: make(chan bool, 1),
	}
	return c
}

// StartReader 处理conn读数据的Goroutine
func (c *Connection) StartReader() {
	fmt.Sprintf("Reader Goroutine is running....")
	defer fmt.Println(c.RemoteAddr().String(), " conn reader exit")
	defer c.Stop()

	dp := NewDataPack()
	for {
		//将最大的数据读取到buf中
		headData := make([]byte, dp.GetHeadLen())
		if _, err := io.ReadFull(c.GetTCPConnection(), headData); err != nil {
			fmt.Println("read msg head error: ", err)
			c.ExitBuffChan <- true
			continue
		}

		//拆包得到msgId和dataLen后放入msg中
		msg, err := dp.Unpack(headData)
		if err != nil {
			fmt.Println("unpack error", err)
			c.ExitBuffChan <- true
			continue
		}

		//根据dataLen读取data，放到msg.Data中
		var data []byte
		if msg.GetDataLen() > 0 {
			data = make([]byte, msg.GetDataLen())
			if _, err := io.ReadFull(c.GetTCPConnection(), data); err != nil {
				fmt.Println("read msg data error ", err)
				c.ExitBuffChan <- true
				continue
			}
		}
		msg.SetData(data)

		//得到当前客户端请求的Request数据
		req := Request{
			conn: c,
			msg:  msg,
		}

		//从路由Routers中找到注册绑定conn的对应Handle
		go c.msgHandler.DoMsgHandler(&req)
	}
}

// Start 启动连接，让当前连接开始工作
func (c *Connection) Start() {
	//开启处理该连接 读取客户端数据之后的请求业务
	go c.StartReader()

	for {
		select {
		case <-c.ExitBuffChan:
			//得到退出消息不再阻塞
			return
		}
	}

}

// Stop 停止连接
func (c *Connection) Stop() {
	//1.如果当前连接已经关闭
	if c.isClosed == true {
		return
	}

	c.isClosed = true

	//TODO: Connection Stop() 如果用户注册了该连接的关闭业务，则在此刻应该显示调用

	//关闭Socket连接
	c.Conn.Close()

	//通知从缓冲队列读数据的业务，该连接已经关闭
	c.ExitBuffChan <- true
	close(c.ExitBuffChan)
}

func (c *Connection) GetTCPConnection() *net.TCPConn {
	return c.Conn
}

func (c *Connection) GetConnID() uint32 {
	return c.ConnID
}

func (c *Connection) RemoteAddr() net.Addr {
	return c.Conn.RemoteAddr()
}

func (c *Connection) SendMsg(msgId uint32, data []byte) error {
	if c.isClosed == true {
		return errors.New("connection closed while send msg")
	}

	//将dataLen封包并发送
	dp := NewDataPack()
	msg, err := dp.Pack(NewMessage(msgId, data))
	if err != nil {
		fmt.Println("Pack msg id = ", msgId)
		return errors.New("pack error msg")
	}

	//写回客户端
	if _, err := c.Conn.Write(msg); err != nil {
		fmt.Println("write msg id ", msgId, " error")
		c.ExitBuffChan <- true
		return errors.New("conn write err")
	}
	return nil
}
