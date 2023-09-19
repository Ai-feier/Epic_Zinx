package main

import (
	"fmt"
	"net"
	"time"
)

/**
	模拟客户端输入
 */
func main() {
	fmt.Println("clliet start...")
	time.Sleep(1*time.Second)
	conn, err := net.Dial("tcp4", "127.0.0.1:8008")
	if err != nil {
		fmt.Println("conn error ", err)
		return
	}
	// 模拟输入
	for {
		cnt, err := conn.Write([]byte("hello world"))
		if err != nil {
			fmt.Println("write error", err)
			return
		}

		// 读取回显的数据
		buf := make([]byte, 512)
		cnt, err = conn.Read(buf[:cnt])
		if err != nil {
			fmt.Println("recv conn error", err)
			return
		}

		fmt.Printf("server call back: %s, count: %d\n", string(buf[:cnt]), cnt)

		// cpu 阻塞
		time.Sleep(1*time.Second)
	}
}