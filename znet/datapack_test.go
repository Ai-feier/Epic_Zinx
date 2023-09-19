package znet

import (
	"fmt"
	"io"
	"net"
	"testing"
)

// 测试拆包, 封包
// 模拟客户端
func TestDataPack(t *testing.T) {
	// 1. 创建socket
	lis, err := net.Listen("tcp", "127.0.0.1:8007")
	if err != nil {
		fmt.Println("server listen error ", err)
		return
	}
	go func() {
		for {
			conn, err := lis.Accept()
			if err != nil {
				fmt.Println("accept conn error", err)
				continue
			}
			// 开启一个协程从 io 流中读取数据
			go func() {
				dp := NewDataPack()
				for {
					headData := make([]byte, dp.GetHeadLen())
					_, err := io.ReadFull(conn, headData)
					if err != nil {
						fmt.Println("receive package header error", err)

					}
					// 将请求头字节流, 转换为message 类
					msgHead, err := dp.Unpack(headData)
					if err != nil {
						fmt.Println("unpack header error", err)
						return
					}
					if msgHead.GetDataLen() > 0 {
						// 有消息内容
						msg := msgHead.(*Message)
						msg.Data = make([]byte, msg.GetDataLen())
						_, err := io.ReadFull(conn, msg.Data)
						if err != nil {
							fmt.Println("server unpack data error", err)
							return
						}
						fmt.Println("==> Recv Msg: ID=", msg.ID, ", len=", msg.DataLen, ", data=", string(msg.Data))
					}
				}
			}()
		}
	}()


	// 2. 模拟客户端
	conn, err := net.Dial("tcp", "127.0.0.1:8007")
	if err != nil {
		fmt.Println("client connect error", err)
		return
	}
	// 创建一个封包对象
	dp := NewDataPack()
	// 模拟粘包过程, 将两个包放在一起发送
	msg1 := &Message{
		ID: 3,
		DataLen: 3,
		Data: []byte{'a', 'b', 'c'},
	}
	msg2 := &Message{
		ID: 7,
		DataLen: 7,
		Data:    []byte{'w', 'o', 'r', 'l', 'd', '!', '!'},
	}

	m1, err := dp.Pack(msg1)
	if err != nil {
		fmt.Println("client pack error", err)
		return
	}
	m2, err := dp.Pack(msg2)
	if err != nil {
		fmt.Println("client pack error", err)
		return
	}
	m1 = append(m1, m2...)
	cnt, err := conn.Write(m1)
	if err != nil {
		fmt.Println("client send error", err)
		return
	}
	fmt.Println("client send data's count: ", cnt)

	// 阻塞主线程
	select {}

}
