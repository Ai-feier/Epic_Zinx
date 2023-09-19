package main

import (
	"fmt"
	"time"
	"zinx/ziface"
	"zinx/znet"
)

//ping test 自定义路由
type PingRouter struct {
	znet.BaseRouter
}

// Handle Ping Handle
func (this *PingRouter) Handle(request ziface.IRequest) {
	fmt.Println("Call PingRouter Handle")
	//先读取客户端的数据，再回写ping...ping...ping
	fmt.Println("recv from client : msgId=", request.GetMsgID(), ", data=", string(request.GetData()))

	err := request.GetConnection().SendMsg(0, []byte("ping...ping...ping"))
	if err != nil {
		fmt.Println(err)
	}
}

//HelloZinxRouter Handle
type HelloZinxRouter struct {
	znet.BaseRouter
}

func (this *HelloZinxRouter) Handle(request ziface.IRequest) {
	fmt.Println("Call HelloZinxRouter Handle")
	//先读取客户端的数据，再回写ping...ping...ping
	fmt.Println("recv from client : msgId=", request.GetMsgID(), ", data=", string(request.GetData()))

	err := request.GetConnection().SendMsg(1, []byte("Hello Zinx Router V1.0"))
	if err != nil {
		fmt.Println(err)
	}
}

func CallConnBegin(c ziface.IConnection) {
	fmt.Println("CallConnBegin is Called ... ")
	err := c.SendMsg(2, []byte("CallConnBegin BEGIN..."))
	if err != nil {
		fmt.Println(err)
	}

	//=============设置两个链接属性，在连接创建之后===========
	fmt.Println("Set conn Name, Home done!")
	c.SetProperty("Name", "yangzhuolin")
	c.SetProperty("Home", "http://github.com/Ai-feier")
	//===================================================


	// 开启一个协程, 用于测试关闭的 HOOK 方法
	go func() {
		select {
		case <-time.After(23*time.Second):
			c.Stop()
		}
	}()
}

//连接断开的时候执行
func CallConnEnd(conn ziface.IConnection) {
	//============在连接销毁之前，查询conn的Name，Home属性=====
	if name, err:= conn.GetProperty("Name"); err == nil {
		fmt.Println("Conn Property Name = ", name)
	}

	if home, err := conn.GetProperty("Home"); err == nil {
		fmt.Println("Conn Property Home = ", home)
	}
	//===================================================

	fmt.Println("CallConnEnd is Called ... ")
}

func main() {
	//创建一个server句柄
	s := znet.NewServer()

	//配置路由
	s.AddRouter(0, &PingRouter{})
	s.AddRouter(1, &HelloZinxRouter{})

	// 注册 HOOK 方法
	s.SetOnConnStart(CallConnBegin)
	s.SetOnConnStop(CallConnEnd)

	//开启服务
	s.Serve()
}