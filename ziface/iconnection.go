package ziface

import "net"

type IConnection interface {
	Start()                                      //启动连接
	Stop()                                       //停止连接
	GetTCPConnection() *net.TCPConn              //从当前连接获取原始的 socket TCPConnection
	GetConnID() uint32                           //获取当前连接ID
	RemoteAddr() net.Addr                        //获取远程客户端的地址信息
	SendMsg(msgId uint32, data []byte) error     //直接将Message发送给远程的TCP客户端,无缓冲
	SendBuffMsg(msgId uint32, data []byte) error //添加带缓冲区多的发送消息的接口

	SetProperty(key string, value any)   //设置连接属性
	GetProperty(key string) (any, error) //获取连接属性
	RemoveProperty(key string)           //移除连接属性
}

type HandFunc func(*net.TCPConn, []byte, int) error
