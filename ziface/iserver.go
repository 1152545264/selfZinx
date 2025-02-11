package ziface

type IServer interface {
	Start()                                 //启动服务器
	Stop()                                  //停止服务器
	Serve()                                 //开启业务服务
	AddRouter(msgId uint32, router IRouter) //路由功能，给当前服务注册一个路由业务方法，供处理客户端连接使用
	GetConnMgr() IConnManager               //获取连接管理器

	SetOnConnStart(func(conn IConnection)) //设置该Server连接创建时的hook函数
	SetOnConnStop(func(conn IConnection))  //设置该Server连接断开时的HOOK函数
	CallOnConnStart(conn IConnection)      //调用连接OnConnStart hook函数
	CallOnConnStop(conn IConnection)       //调用连接 OnConnStop hook函数
}
