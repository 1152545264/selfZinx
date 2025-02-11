package znet

import (
	"errors"
	"fmt"
	"strconv"
	"sync"
	"zinx/ziface"
)

type ConnManager struct {
	connections map[uint32]ziface.IConnection //管理的连接信息
	connLock    sync.RWMutex                  //读写连接的读写锁
}

func NewConnManager() *ConnManager {
	res := &ConnManager{connections: make(map[uint32]ziface.IConnection)}
	return res
}

func (cm *ConnManager) Add(conn ziface.IConnection) { //添加连接
	//保护共享资源加读写锁
	cm.connLock.Lock()
	defer cm.connLock.Unlock()

	//将conn连接添加到ConnManager中
	cm.connections[conn.GetConnID()] = conn

	fmt.Println("connection add to ConnManager successfully, conn num= ", cm.Len())
}

func (cm *ConnManager) Remove(conn ziface.IConnection) { //删除连接
	cm.connLock.Lock()
	defer cm.connLock.Unlock()

	//删除连接信息
	delete(cm.connections, conn.GetConnID())
	fmt.Println("connection remove connId=", conn.GetConnID(), " successfully: conn num=",
		cm.Len())
}

func (cm *ConnManager) Get(connId uint32) (ziface.IConnection, error) { //获取连接
	cm.connLock.RLock()
	defer cm.connLock.RUnlock()

	if conn, ok := cm.connections[connId]; ok {
		return conn, nil
	} else {
		return nil, errors.New("connections not found, " +
			"connId=" + strconv.Itoa(int(connId)))
	}
}

func (cm *ConnManager) Len() int { //获取当前连接的数量
	return len(cm.connections)
}

func (cm *ConnManager) ClearConn() { //删除并停止所有连接
	cm.connLock.Lock() //保护共享资源加读写锁
	defer cm.connLock.Unlock()

	//停止并删除全部连接
	for connId, conn := range cm.connections {
		//停止
		conn.Stop()
		//删除
		delete(cm.connections, connId)
	}

	fmt.Println("Clear All connections successfully, conn num=", cm.Len())
}
