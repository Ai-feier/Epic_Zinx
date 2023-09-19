package main

import "zinx/znet"

/**
	基于 zinx 框架实现应用程序
 */

func main() {
	// 1. 创建 server 句柄, 使用 zinx 的 api
	server := znet.NewServer("[zinx v0.2]")
	// 2. 启动服务
	server.Serve()
}