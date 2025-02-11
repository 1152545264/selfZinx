package znet

import (
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"zinx/utils"
	"zinx/ziface"
)

type Connection struct {
	TcpServer    ziface.IServer    //当前conn属于哪个Server，在conn初始化的时候添加即可
	Conn         *net.TCPConn      //当前连接的socket tcp套接字
	ConnID       uint32            //当前连接的ID也可以称为SessionID, ID全局唯一
	isClosed     bool              //当前连接的关闭状态
	msgHandler   ziface.IMsgHandle //该连接的处理方法
	ExitBuffChan chan bool         //告知该连接已经退出/停止的channel
	msgChan      chan []byte       //无缓冲管道，用于读写两个goroutine之间的消息通信
	msgBuffChan  chan []byte       //有缓冲管道，用于读写两个goroutine之间的消息通信

	property     map[string]any
	propertyLock sync.RWMutex //保护连接属性的锁
}

func NewConnection(server ziface.IServer, conn *net.TCPConn, connID uint32, msgHandler ziface.IMsgHandle) *Connection {
	c := &Connection{
		TcpServer:    server, //将隶属的Server传进来
		Conn:         conn,
		ConnID:       connID,
		isClosed:     false,
		msgHandler:   msgHandler,
		ExitBuffChan: make(chan bool, 1),
		msgChan:      make(chan []byte), //msgChan初始化
		msgBuffChan:  make(chan []byte, utils.GlobalObject.MaxMsgChanLen),
		property:     make(map[string]any), //对连接属性map初始化
	}
	//将新建的conn添加到连接管理中
	c.TcpServer.GetConnMgr().Add(c)
	return c
}

// Start 启动连接，让当前连接开始工作
func (c *Connection) Start() {
	//1. 开启用户从客户端读取数据流程的goroutine
	go c.StartReader()
	//2. 开启用于写回客户端数据流程的goroutine
	go c.StartWriter()

	//按照用户传递进来的创建连接是需要处理的业务执行钩子方法
	c.TcpServer.CallOnConnStart(c)

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

	// 如果用户注册了该连接的关闭业务，则在此刻应该显示调用
	c.TcpServer.CallOnConnStop(c)

	//关闭Socket连接
	c.Conn.Close()

	//通知从缓冲队列读数据的业务，该连接已经关闭
	c.ExitBuffChan <- true

	//将conn从ConnManager中删除
	c.TcpServer.GetConnMgr().Remove(c)

	//关闭该连接的全部管道
	close(c.ExitBuffChan)
	close(c.msgChan)
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

	//写回客户端, 发送给Channel，供writer 读取
	c.msgChan <- msg
	return nil
}

// SendBuffMsg 带缓冲区的消息发送接口
func (c *Connection) SendBuffMsg(msgId uint32, data []byte) error {
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

	//写回客户端, 发送给Channel，供writer 读取
	c.msgChan <- msg
	return nil
}

func (c *Connection) SetProperty(key string, value any) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()
	c.property[key] = value
}
func (c *Connection) GetProperty(key string) (any, error) {
	c.propertyLock.RLock()
	defer c.propertyLock.RUnlock()

	if value, ok := c.property[key]; ok {
		return value, nil
	} else {
		return nil, errors.New("no property found, key=" + key)
	}
}

func (c *Connection) RemoveProperty(key string) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()

	delete(c.property, key)
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
			break
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
		if utils.GlobalObject.WorkerPoolSize > 0 {
			//已经启动工作池机制则将消息交给worker处理
			c.msgHandler.SendMsgToTaskQueue(&req)
		} else {
			//从绑定好的消息和对应的处理方法中执行对应的Handle方法
			go c.msgHandler.DoMsgHandler(&req)
		}

	}
}

func (c *Connection) StartWriter() {
	fmt.Println("[Writer Goroutine is running]")
	defer fmt.Println(c.RemoteAddr().String(), "[conn Writer exit!]")

	for {
		select {
		case data := <-c.msgChan:
			//有数据需要写给客户端
			if _, err := c.Conn.Write(data); err != nil {
				fmt.Println("Send data error:, ", err, " Conn writer exit")
				return
			}
		case data, ok := <-c.msgBuffChan:
			if ok {
				//有数据需要发送给客户端
				if _, err := c.Conn.Write(data); err != nil {
					fmt.Println("Send buf Data error:, ", err, " Conn write exit")
					return
				} else {
					fmt.Println("msgBufChan is Closed")
					break
				}
			}

		case <-c.ExitBuffChan:
			//conn已经关闭
			return
		}
	}
}
