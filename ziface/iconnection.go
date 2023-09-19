package ziface

import "net"

// IConnection 抽象连接模块
type IConnection interface {
	// Start 启动连接
	Start()

	// Stop Start停止链接
	Stop()

	// GetConnID 获得连接 ID
	GetConnID() uint32

	// GetTCPConnection 获取连接的socket
	GetTCPConnection() *net.TCPConn

	// RemoteAddr 获取当前连接的ID
	RemoteAddr() net.Addr

	// SendMsg 发送数据 (无缓冲)
	SendMsg(msgId uint32, data []byte) error

	// SendBuffMsg 直接将Message数据发送给远程的TCP客户端(有缓冲)
	SendBuffMsg(msgId uint32, data []byte) error   //添加带缓冲发送消息接

	// SetProperty 设置链接属性
	SetProperty(key string, value interface{})

	// GetProperty 获取链接属性
	GetProperty(key string)(interface{}, error)

	// RemoveProperty 移除链接属性
	RemoveProperty(key string)
}

// 定义处理连接业务的方法
// 参数: tcp连接, 要处理的数据, 数据的长度
type HandleFunc func(*net.TCPConn, []byte, int) error
