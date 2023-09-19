package znet

import (
	"zinx/ziface"
)

// 实现抽象类IRequest
type Request struct {
	conn ziface.IConnection  // 已建立连接
	msg ziface.IMessage  // 客户端请求的数据
}

func (r *Request) GetConnection() ziface.IConnection {
	return r.conn
}

func (r *Request) GetData() []byte {
	return r.msg.GetData()
}

//获取请求的消息的ID
func (r *Request) GetMsgID() uint32 {
	return r.msg.GetMsgID()
}

