package znet

type Message struct {
	ID uint32
	DataLen uint32
	Data []byte
}

// 设计一个创建消息的工厂方法
func NewMessage(id uint32, dataLen uint32, data []byte) *Message {
	return &Message{
		ID: id,
		DataLen: dataLen,
		Data: data,
	}
}

func (m *Message) GetMsgID() uint32 {
	return m.ID
}

func (m *Message) GetDataLen() uint32 {
	return m.DataLen
}

func (m *Message) GetData() []byte {
	return m.Data
}

func (m *Message) SetMsgID(u uint32) {
	m.ID = u
}

func (m *Message) SetDataLen(u uint32) {
	m.DataLen = u
}

func (m *Message) SetData(bytes []byte) {
	m.Data = bytes
}

