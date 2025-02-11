package znet

import (
	"errors"
	"fmt"
	"net"
	"zinx/utils"
	"zinx/ziface"
)

type Server struct {
	Name       string              // 服务器名称
	IPVersion  string              // tcp4 or other
	IP         string              // 服务器绑定的IP地址
	Port       int                 // 服务器绑定的端口
	msgHandler ziface.IMsgHandle   // 当前Server由用户绑定回调router，也即Server注册的连接对应的处理业务
	ConnMgr    ziface.IConnManager //当前Server的连接管理器

	OnConnStart func(conn ziface.IConnection) // 该Server连接创建时的hook函数
	OnConnStop  func(conn ziface.IConnection) // 该Server连接断开时的hook函数
}

// CallBackToClient *******************定义当前客户端连接的handleAPI********************
func CallBackToClient(conn *net.TCPConn, data []byte, cnt int) error {
	//回显业务
	fmt.Println("[Conn Handle] CallBackToClient...")
	if _, err := conn.Write(data[:cnt]); err != nil {
		fmt.Println("write back buf err ", err)
		return errors.New("CallbackToClient error")
	}
	return nil
}

// Start 开启网络服务
func (s *Server) Start() {
	fmt.Printf("[START] Server listenerat IP: %s, Port:%d, is starting\n", s.IP, s.Port)
	fmt.Printf("[Zinx] version: %s, MaxConn: %d, MaxPacketSize: %d\n",
		utils.GlobalObject.Version, utils.GlobalObject.MaxConn, utils.GlobalObject.MaxPacketSize)

	// 开启一个goroutine去做服务器listener业务
	go func() {
		//0 启动工作池机制
		s.msgHandler.StartWorkerPool()

		//1.获取一个TCP的addr
		addr, err := net.ResolveTCPAddr(s.IPVersion, fmt.Sprintf("%s:%d", s.IP, s.Port))
		if err != nil {
			fmt.Println("resolve tcp addr err:", err)
			return
		}

		// 2.监听服务器地址
		listener, err := net.ListenTCP(s.IPVersion, addr)
		if err != nil {
			fmt.Println("Listen", s.IPVersion, "err: ", err)
			return
		}

		//已经监听成功
		fmt.Println("start Zinx server: ", s.Name, " success, now listening...")

		// TODO: server.go应该有一个自动生成ID的方法
		var cid uint32
		cid = 0

		//3.启动server网络连接业务
		for {
			//3.1.阻塞等会带客户端建立连接请求
			conn, err := listener.AcceptTCP()
			if err != nil {
				fmt.Println("Accept err: ", err)
				continue
			}
			//3.2 Server.Start()设置服务器最大连接控制，如果超过最大连接，则关闭新的连接
			if s.ConnMgr.Len() >= utils.GlobalObject.MaxConn {
				conn.Close()
				continue
			}

			//3.3 处理该连接请求的业务方法，此时handle和conn应该是绑定的
			dealConn := NewConnection(s, conn, cid, s.msgHandler)
			cid++

			//3.4启动当前连接的处理业务
			go dealConn.Start()
		}
	}()
}

func (s *Server) Stop() {
	fmt.Sprintln("[STOP] Zinx server, name: ", s.Name)

	// Server.Stop()将需要清理的连接信息或者其他信息一并停止或者清理
	s.ConnMgr.ClearConn()

}
func (s *Server) Serve() {
	s.Start()

	//TODO: Server.Serve()如果在启动服务的时候还需要处理其他的事情，则可以在这里添加

	//阻塞，否则主goroutine退出，listener的go也将会退出
	select {}
}

// AddRouter 路由功能，给当前服务注册一个路由业务方法，供客户端连接处理使用
func (s *Server) AddRouter(msgId uint32, router ziface.IRouter) {
	s.msgHandler.AddRouter(msgId, router)

	fmt.Println("Add Router success!")

}

func (s *Server) GetConnMgr() ziface.IConnManager { //获取连接管理器
	return s.ConnMgr
}

func (s *Server) SetOnConnStart(hookFunc func(connection ziface.IConnection)) {
	s.OnConnStart = hookFunc
}

func (s *Server) SetOnConnStop(hookFunc func(connection ziface.IConnection)) {
	s.OnConnStop = hookFunc
}

// CallOnConnStart 调用连接 OnConnStart hook函数
func (s *Server) CallOnConnStart(conn ziface.IConnection) {
	if s.OnConnStart != nil {
		fmt.Println("====>CallOnConnStart.....")
		s.OnConnStart(conn)
	}
}

func (s *Server) CallOnConnStop(conn ziface.IConnection) {
	if s.OnConnStop != nil {
		fmt.Println("====>CallOnConnStop.....")
		s.OnConnStop(conn)
	}
}

func NewServer(name string) ziface.IServer {
	nameStr := utils.GlobalObject.Name
	if len(nameStr) == 0 {
		nameStr = name
	}
	s := &Server{ //从全局参数获取配置
		Name:       nameStr,
		IPVersion:  "tcp4",
		IP:         utils.GlobalObject.Host,
		Port:       utils.GlobalObject.TcpPort,
		msgHandler: NewMsgHandle(),   //默认不指定
		ConnMgr:    NewConnManager(), //创建ConnMgr
	}
	return s
}
