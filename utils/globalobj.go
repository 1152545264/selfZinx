package utils

import (
	"encoding/json"
	"io/ioutil"
	"zinx/ziface"
)

/*
* 存储一切有关zinx框架的全局参数，供其他模块使用
一些参数也可以通过用户根据zinx.json来配置
*/
type GlobalObj struct {
	TcpServer ziface.IServer //当前zinx的全局Server对象
	Host      string         //当前服务器的主机IP
	TcpPort   int            // 当前服务器主机监听端口号
	Name      string         //当前服务器名称
	Version   string         //当前zinx的版本号

	MaxPacketSize uint32 //读取数据包的最大值
	MaxConn       int    //当前服务器主机允许的最大连接数

	WorkerPoolSize   uint32 //业务工作池worker的数量
	MaxWorkerTaskLen uint32 // 业务工作worker对应负责的任务队列的最大数量

	ConfigFilePath string // 配置文件路径
}

var GlobalObject *GlobalObj //定义一个全局对象

func (g *GlobalObj) Reload() {
	data, err := ioutil.ReadFile("conf/zinx.json")
	if err != nil {
		panic(err)
	}

	//将json数据解析到struct中
	err = json.Unmarshal(data, g)
	if err != nil {
		panic(err)
	}
}

func init() {
	//初始化GlobalObject设置一些默认值
	GlobalObject = &GlobalObj{
		Name:          "ZinxServerApp",
		Version:       "v0.8",
		TcpPort:       7777,
		Host:          "0.0.0.0",
		MaxConn:       12000,
		MaxPacketSize: 4096,

		WorkerPoolSize:   100,
		MaxWorkerTaskLen: 1024,
	}

	//从配置文件中加载一些用户配置的参数
	GlobalObject.Reload()
}
