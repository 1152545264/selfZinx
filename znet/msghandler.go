package znet

import (
	"fmt"
	"strconv"
	"zinx/ziface"
)

type MsgHandle struct {
	Apis map[uint32]ziface.IRouter //存放每个msgId所对应的处理方法
}

func NewMsgHandle() *MsgHandle {
	return &MsgHandle{Apis: make(map[uint32]ziface.IRouter)}
}

func (msh *MsgHandle) DoMsgHandler(request ziface.IRequest) { //以非阻塞方式处理消息
	handler, ok := msh.Apis[request.GetMsgID()]
	if !ok {
		fmt.Println("api msgId=", request.GetMsgID(), " is not FOUND")
		return
	}

	//执行对应的处理方法
	handler.PreHandle(request)
	handler.Handle(request)
	handler.PostHandle(request)
}
func (msh *MsgHandle) AddRouter(msgId uint32, router ziface.IRouter) { //为消息添加具体的处理逻辑
	//判断当前msg绑定的api处理方法是否已经存在
	if _, ok := msh.Apis[msgId]; ok {
		panic("repeated api, msgId = " + strconv.Itoa(int(msgId)))
	}

	//2.添加msg与api的绑定关系
	msh.Apis[msgId] = router
	fmt.Println("Add api msgId = ", msgId)
}
