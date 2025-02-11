package ziface

type IServer interface {
	Start()                                 //启动服务器
	Stop()                                  //停止服务器
	Serve()                                 //开启业务服务
	AddRouter(msgId uint32, router IRouter) //路由功能，给当前服务注册一个路由业务方法，供处理客户端连接使用
}
