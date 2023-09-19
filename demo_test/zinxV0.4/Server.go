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
	_, err := req.GetConnection().GetTCPConnection().Write([]byte("ping...ping...ping\n"))
	if err !=nil {
		fmt.Println("call back ping ping ping error")
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