package znet

import (
	"fmt"
	"zinx/utils"
	"zinx/ziface"
)

type MsgHandle struct {
	// 多叉树实现, 一种消息对应一组 router
	Apis           map[uint32]ziface.IRouter //存放每个MsgId 所对应的处理方法的map属性
	WorkerPoolSize uint32                    //业务工作Worker池的数量
	TaskQueue      []chan ziface.IRequest    //Worker负责取任务的消息队列
}

func NewMsgHandle() ziface.IMsgHandle {
	return &MsgHandle{
		Apis:           make(map[uint32]ziface.IRouter),
		WorkerPoolSize: utils.GlobalObject.WorkerPoolSize,
		//一个worker对应一个queue
		TaskQueue: make([]chan ziface.IRequest, utils.GlobalObject.WorkerPoolSize),
	}
}

func (m *MsgHandle) DoMsgHandler(request ziface.IRequest) {
	// 根据 msgID 获取相应的 router
	handle, ok := m.Apis[request.GetMsgID()]
	if !ok {
		fmt.Println("handle error: 当前消息类型未注册相应的 router")
		return
	}

	// 调用相应的 handle
	handle.PreHandle(request)
	handle.Handle(request)
	handle.PostHandle(request)
}

// AddRouter 为消息添加具体的处理逻辑
func (m *MsgHandle) AddRouter(msgId uint32, router ziface.IRouter) {
	// 1. 判断是否重复注册
	if _, ok := m.Apis[msgId]; ok {
		panic("msgHandle error: handle is existed")
	}
	//2 添加msg与api的绑定关系
	m.Apis[msgId] = router
	fmt.Printf("msgID: %d, add router success!", msgId)
}

// StartWorkerPool 启动worker工作池
func (m *MsgHandle) StartWorkerPool() {
	for i := 0; i < int(m.WorkerPoolSize); i++ {
		// 开启相应数量 goroutine 的任务队列
		m.TaskQueue[i] = make(chan ziface.IRequest, utils.GlobalObject.MaxWorkerTaskLen)

		// 开启一个 goroutine 监听该任务队列
		go m.StartOneWorker(i, m.TaskQueue[i])
	}
}

// StartOneWorker 开启单个 goroutine 监听任务队列
func (m *MsgHandle) StartOneWorker(id int, taskQueue chan ziface.IRequest) {
	fmt.Println("Worker ID = ", id, " is started.")
	//不断的等待队列中的消息
	for {
		select {
		case task := <-taskQueue:
			// 当前管道上存在任务, 后处理任务
			m.DoMsgHandler(task)
		}
	}
}

// SendMsgToTaskQueue 将消息交给TaskQueue,由worker进行处理
func (m *MsgHandle) SendMsgToTaskQueue(request ziface.IRequest) {
	//根据ConnID来分配当前的连接应该由哪个worker负责处理
	//轮询的平均分配法则

	//得到需要处理此条连接的workerID
	workerID := request.GetConnection().GetConnID()%utils.GlobalObject.WorkerPoolSize
	fmt.Println("Add ConnID=", request.GetConnection().GetConnID()," request msgID=", request.GetMsgID(), "to workerID=", workerID)

	m.TaskQueue[workerID] <- request
}
