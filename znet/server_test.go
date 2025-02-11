package znet

import (
	"fmt"
	"net"
	"testing"
	"time"
)

// 模拟客户端
func ClientTest() {
	fmt.Println("Client Test....start")
	// 3s后发起测试请求，给服务器开启服务的时间
	time.Sleep(3 * time.Second)

	conn, err := net.Dial("tcp", "127.0.0.1:7777")
	if err != nil {
		fmt.Println("Client start err: ", err, " exit")
		return
	}

	for {
		_, err := conn.Write([]byte("Hello Zinx"))
		if err != nil {
			fmt.Println("write error err:", err)
			return
		}
		buf := make([]byte, 512)
		cnt, err := conn.Read(buf)
		if err != nil {
			fmt.Println("read buf err:", err)
			return
		}
		fmt.Printf("server call back: %s, cnt :%d \n", buf[:cnt], cnt)
		time.Sleep(1 * time.Second)
	}
}

func TestServer(t *testing.T) {
	s := NewServer("[zinx0.1]")
	go ClientTest()
	s.Serve()

}
