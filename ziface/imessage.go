package ziface

/*
	将请求的一个消息封装到message中，定义抽象层接口
*/

type IMessage interface {
	GetMsgID() uint32  // 获得消息ID
	GetDataLen() uint32  // 获得消息数据字段的长度
	GetData() []byte  // 获得消息内容

	SetMsgID(uint32)
	SetDataLen(uint32)
	SetData([]byte)
}
