package znet

import (
	"bytes"
	"encoding/binary"
	"errors"
	"zinx/utils"
	"zinx/ziface"
)

type DataPack struct{}

func (d *DataPack) GetHeadLen() uint32 {
	// 包头固定 8 字节, 长度 uint32(4字节), ID uint32(4字节)
	return 8
}

func (d *DataPack) Pack(msg ziface.IMessage) ([]byte, error) {
	// 创建一个存放 byte 数组的缓冲
	buf := bytes.NewBuffer([]byte{})

	// 写消息长度
	err := binary.Write(buf, binary.LittleEndian, msg.GetDataLen())
	if err != nil {
		 return nil, err
	}

	// 写消息 ID
	err = binary.Write(buf, binary.LittleEndian, msg.GetMsgID())
	if err != nil {
		return nil, err
	}

	// 写消息内容
	err = binary.Write(buf, binary.LittleEndian, msg.GetData())
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (d *DataPack) Unpack(data []byte) (ziface.IMessage, error) {
	buf := bytes.NewReader(data)

	msg := &Message{}

	// 读包头的消息长度
	if err := binary.Read(buf, binary.LittleEndian, &msg.DataLen); err != nil {
		return nil, err
	}

	// 读包头的消息 ID
	if err := binary.Read(buf, binary.LittleEndian, &msg.ID); err != nil {
		return nil, err
	}

	// 判断dataLen是否超过最大长度
	if utils.GlobalObject.MaxPacketSize > 0 && msg.DataLen > utils.GlobalObject.MaxPacketSize {
		return nil, errors.New("Too large msg data received")
	}

	return msg, nil // 此时 msg 中没有 data


}

func NewDataPack() *DataPack {
	return &DataPack{}
}


func NewMsgPackage(msgId uint32, data []byte) ziface.IMessage {
	msg := &Message{
		ID: msgId,
		DataLen: uint32(len(data)),
		Data: data,
	}
	return msg
}