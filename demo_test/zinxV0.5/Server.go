package main

import (
	"fmt"
	"zinx/ziface"
	"zinx/znet"
)

type PingRouter struct {
	znet.BaseRouter
}

func (br *PingRouter)Handle(req ziface.IRequest){
	fmt.Println("Call PingRouter Handle")
	//先读取客户端的数据，再回写ping...ping...ping
	fmt.Println("recv from client : msgId=", req.GetMsgID(), ", data=", string(req.GetData()))

	//回写数据
	err := req.GetConnection().SendMsg(1, []byte("ping...ping...ping"))
	if err != nil {
		fmt.Println(err)
	}
}

func main() {
	// 1. 创建 server 句柄, 使用 zinx 的 api
	server := znet.NewServer()

	// 2. 添加服务器路由类
	server.AddRouter(&PingRouter{})

	// 3. 启动服务
	server.Serve()
}