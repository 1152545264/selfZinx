package main

import (
	"fmt"
	"io"
	"net"
	"zinx/ziface"
	"zinx/znet"
)

type PingRouter struct {
	znet.BaseRouter //一定要先定义基础路由BaseRouter
	msgId           uint32
}

func (this *PingRouter) Handle(request ziface.IRequest) {
	fmt.Println("Call PingRouter Handle")
	fmt.Println("recv form client, msgId=", request.GetMsgID(), " data=", string(request.GetData()))

	//回写数据
	err := request.GetConnection().SendMsg(this.msgId, []byte("ping...ping...ping  "))
	if err != nil {
		fmt.Println(err)
	}
	this.msgId += 1
}

func test1() {
	s := znet.NewServer("[zinxv0.3]")
	s.AddRouter(&PingRouter{})
	s.Serve()
}

// test2 只是负责测试datapack拆包和封包命令
func test2() {
	listener, err := net.Listen("tcp", "127.0.0.1:7777")
	if err != nil {
		fmt.Println("server listen err:", err)
		return
	}

	//创建服务器goroutine，负责从客户端goroutine读取粘包的数据然后进行解析
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Server accept err: ", err)
			return
		}

		//处理客户端请求
		go func(conn net.Conn) {
			//创建封包拆包对象dp
			dp := znet.NewDataPack()
			for {
				//1.先读出流中的head部分
				headData := make([]byte, dp.GetHeadLen())
				// readFull会把msg填充满为止
				_, err := io.ReadFull(conn, headData)
				if err != nil {
					fmt.Println("read head error")
					break
				}

				//2.将headData字节流拆包到msg中
				msgHead, err := dp.Unpack(headData)
				if err != nil {
					fmt.Println("server unpack err: ", err)
					return
				}

				//3.根据dataLen从io中读取字节流
				if msgHead.GetDataLen() > 0 {
					//msg中有data数据，需要在此读取data数据
					msg := msgHead.(*znet.Message)
					msg.Data = make([]byte, msg.GetDataLen())
					_, err := io.ReadFull(conn, msg.Data)
					if err != nil {
						fmt.Println("server unpack error:", err)
						return
					}
					fmt.Println("====> Recv Msg: ID=", msg.Id, ", len=",
						msg.DataLen, ", data=", string(msg.Data))
				}
			}
		}(conn)
	}
}

func main() {
	test1()
	//test2()
}
