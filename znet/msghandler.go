package znet

import (
	"fmt"
	"strconv"
	"zinx/utils"
	"zinx/ziface"
)

type MsgHandle struct {
	Apis           map[uint32]ziface.IRouter //存放每个msgId所对应的处理方法
	WorkerPoolSize uint32                    //业务工作worker池的数量
	TaskQueue      []chan ziface.IRequest    // worker负责任务的消息队列
}

func NewMsgHandle() *MsgHandle {
	res := &MsgHandle{
		Apis:           make(map[uint32]ziface.IRouter),
		WorkerPoolSize: utils.GlobalObject.WorkerPoolSize,
		TaskQueue:      make([]chan ziface.IRequest, utils.GlobalObject.WorkerPoolSize),
	}
	return res
}

func (mh *MsgHandle) DoMsgHandler(request ziface.IRequest) { //以非阻塞方式处理消息
	handler, ok := mh.Apis[request.GetMsgID()]
	if !ok {
		fmt.Println("api msgId=", request.GetMsgID(), " is not FOUND")
		return
	}

	//执行对应的处理方法
	handler.PreHandle(request)
	handler.Handle(request)
	handler.PostHandle(request)
}
func (mh *MsgHandle) AddRouter(msgId uint32, router ziface.IRouter) { //为消息添加具体的处理逻辑
	//判断当前msg绑定的api处理方法是否已经存在
	if _, ok := mh.Apis[msgId]; ok {
		panic("repeated api, msgId = " + strconv.Itoa(int(msgId)))
	}

	//2.添加msg与api的绑定关系
	mh.Apis[msgId] = router
	fmt.Println("Add api msgId = ", msgId)
}

func (mh *MsgHandle) StartWorkerPool() {
	for i := 0; i < int(mh.WorkerPoolSize); i++ {
		//一个worker被启动
		//给当前worker对应的任务队列开辟空间
		mh.TaskQueue[i] = make(chan ziface.IRequest, utils.GlobalObject.MaxWorkerTaskLen)
		//启动当前worker，阻塞地等待对应的任务队列是否有消息传递进来
		go mh.StartOneWorker(i, mh.TaskQueue[i])
	}
}

// SendMsgToTaskQueue 将消息交给TaskQueue,由worker进行处理
func (mh *MsgHandle) SendMsgToTaskQueue(request ziface.IRequest) {
	// 根据ConnID和轮询原则来分配当前的连接应该由哪个worker负责处理
	workerID := request.GetConnection().GetConnID() % mh.WorkerPoolSize

	fmt.Println("Add ConnID = ", request.GetConnection().GetConnID(), "request msgId=",
		request.GetMsgID(), "to workerID=", workerID)

	//将消息请求发送到任务队列
	mh.TaskQueue[workerID] <- request
}

func (mh *MsgHandle) StartOneWorker(workerID int, taskQueue chan ziface.IRequest) {
	fmt.Println("WorkerID = ", workerID, ", is started")
	//不断地等待队列中的消息
	for {
		select {
		case req := <-taskQueue:
			mh.DoMsgHandler(req)
		}
	}
}
