package main

import (
	"fmt"
	"io"
	"net"
	"time"
	"zinx/utils"
	"zinx/znet"
)

func Test1() {
	fmt.Println("Client Test....start")
	// 3s后发起测试请求，给服务器开启服务的时间
	time.Sleep(3 * time.Second)

	conn, err := net.Dial("tcp", "127.0.0.1:7777")
	if err != nil {
		fmt.Println("Client start err: ", err, " exit")
		return
	}

	var i uint32 = 0
	for {
		//发封包message消息
		dp := znet.NewDataPack()
		msgPk := znet.NewMessage(i%2, []byte(utils.GlobalObject.Name+utils.GlobalObject.Version+", client Test Message"))
		msg, _ := dp.Pack(msgPk)
		_, err := conn.Write(msg)
		if err != nil {
			fmt.Println("write error err: ", err)
			return
		}

		//先读出流中的header部分
		headData := make([]byte, dp.GetHeadLen())
		_, err = io.ReadFull(conn, headData)
		if err != nil {
			fmt.Println("read head error")
			break
		}

		//将headData字节流拆包到msg中
		msgHead, err := dp.Unpack(headData)
		if err != nil {
			fmt.Println("server unpack err: ", err)
			return
		}
		if msgHead.GetDataLen() > 0 {
			// msg有数据需要再次读取data数据
			msg := msgHead.(*znet.Message)
			msg.Data = make([]byte, msg.GetDataLen())

			//根据dataLenlen从io流中读取字节流
			_, err := io.ReadFull(conn, msg.Data)
			if err != nil {
				fmt.Println("server unpack data err: ", err)
				return
			}

			fmt.Println("===> Recv Msg: ID=", msg.Id, ", len=", msg.DataLen,
				", data=", string(msg.Data))
		}

		i += 1
		time.Sleep(1 * time.Second)
		if i >= 4 {
			break
		}
	}
}

func Test2() {
	var i uint32 = 0
	// 客户端goroutine 负责模拟粘包的数据然后进行发送
	conn, err := net.Dial("tcp", "127.0.0.1:7777")
	if err != nil {
		fmt.Println("Client dial err:", err)
		return
	}

	//1. 创建一个封包对象dp
	dp := znet.NewDataPack()

	for {
		//2.封装一个msg1包
		msg1 := &znet.Message{
			Id:      i,
			DataLen: 5,
			Data:    []byte("hello"),
		}
		sendData1, err := dp.Pack(msg1)
		if err != nil {
			fmt.Println("Client pack msg1 failed:", err)
			return
		}

		//3.封装一个msg2包
		msg2 := &znet.Message{
			Id:      i + 1,
			DataLen: 7,
			Data:    []byte("world!!"),
		}

		sendData2, err := dp.Pack(msg2)
		if err != nil {
			fmt.Println("Client pack msg1 failed:", err)
			return
		}

		// 4.将sendData1和sendData2拼接到一起组成粘包
		sendData1 = append(sendData1, sendData2...)

		// 5.向服务器写数据
		conn.Write(sendData1)

		i += 2
		time.Sleep(1 * time.Second)
	}

}

func main() {
	Test1()

	//Test2()
}
