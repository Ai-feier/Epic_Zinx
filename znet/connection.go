package znet

import (
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"zinx/utils"
	"zinx/ziface"
)

// 连接模块
type Connection struct {
	//当前Conn属于哪个Server
	TcpServer ziface.IServer //当前conn属于哪个server，在conn初始化的时候添加即可

	// 连接套接字
	Conn *net.TCPConn

	// 连接ID
	ConnID uint32

	// 连接关闭状态
	isClosed bool

	//消息管理MsgId和对应处理方法的消息管理模块
	MsgHandler ziface.IMsgHandle

	//	告知连接关闭的 channel
	ExitBufChan chan bool

	//无缓冲管道，用于读、写两个goroutine之间的消息通信
	msgChan chan []byte

	//有关冲管道，用于读、写两个goroutine之间的消息通信
	msgBuffChan chan []byte //定义channel成员


	//链接属性
	property     map[string]interface{}
	//保护链接属性修改的锁
	propertyLock sync.RWMutex
}

func (c *Connection) GetTCPConnection() *net.TCPConn {
	return c.Conn
}

func (c *Connection) RemoteAddr() net.Addr {
	return c.Conn.RemoteAddr()
}

func (c *Connection) GetConnID() uint32 {
	return c.ConnID
}

// NewConnection 新建连接
func NewConnection(server *Server, conn *net.TCPConn, connid uint32, msgHandle ziface.IMsgHandle) *Connection {
	newConn := &Connection{
		TcpServer:   server,
		Conn:        conn,
		ConnID:      connid,
		isClosed:    false,
		MsgHandler:  msgHandle,
		ExitBufChan: make(chan bool),
		msgChan:     make(chan []byte), //msgChan初始化
		msgBuffChan: make(chan []byte, utils.GlobalObject.MaxMsgChanLen),
		property:     make(map[string]interface{}), //对链接属性map初始化
	}
	newConn.TcpServer.GetConnMgr().Add(newConn)
	return newConn
}

// Start 启动连接
func (c *Connection) Start() {
	// 开启处理业务协程
	go c.StartReader()
	go c.StartWriter()

	// 调用 HOOK 方法
	c.TcpServer.CallOnConnStart(c)

	for {
		select {
		case <-c.ExitBufChan:
			// 获得退出消息
			return
		}
	}
}

// StartReader 从连接中读取数据, 调用相应的的api
func (c *Connection) StartReader() {
	fmt.Println("Reader Goroutine is  running")
	defer fmt.Println(c.RemoteAddr().String(), " conn reader exit!")
	defer c.Stop()

	for {
		// 创建拆包解包的对象
		dp := NewDataPack()
		// 分为两次读操作: 读头部, 读消息内容
		headData := make([]byte, dp.GetHeadLen())
		_, err := io.ReadFull(c.Conn, headData)
		if err != nil {
			fmt.Println("read msg head error", err)
			c.ExitBufChan <- true
			continue
		}

		msg, err := dp.Unpack(headData)
		if err != nil {
			fmt.Println("unpack headData error", err)
			c.ExitBufChan <- true
			continue
		}

		//根据 dataLen 读取 data，放在msg.Data中
		var data []byte
		if msg.GetDataLen() > 0 {
			data = make([]byte, msg.GetDataLen())
			if _, err := io.ReadFull(c.Conn, data); err != nil {
				fmt.Println("read msg data error", err)
				c.ExitBufChan <- true
				continue
			}
		}
		msg.SetData(data)

		// 得到当前客户端请求的Request数据
		req := &Request{
			conn: c,
			msg:  msg,
		}
		if utils.GlobalObject.WorkerPoolSize > 0 {
			//已经启动工作池机制，将消息交给Worker处理
			c.MsgHandler.SendMsgToTaskQueue(req)
		} else {
			//从绑定好的消息和对应的处理方法中执行对应的Handle方法
			go c.MsgHandler.DoMsgHandler(req)
		}
	}
}

// StartWriter 写消息Goroutine， 用户将数据发送给客户端
func (c *Connection) StartWriter() {
	fmt.Println("[Writer Goroutine is running]")
	defer fmt.Println(c.RemoteAddr().String(), "[conn Writer exit!]")

	for {
		select {
		case data := <-c.msgChan:
			// 从管道中读取数据, 有数据要写给客户端
			if _, err := c.Conn.Write(data); err != nil {
				fmt.Println("Send Data error:, ", err, " Conn Writer exit")
				return
			}
		case dataBuf, ok := <-c.msgBuffChan:
				if ok {
					if _, err := c.Conn.Write(dataBuf);err!=nil {
						fmt.Println("Send Buff Data error:, ", err, " Conn Writer exit")
						return
					}
				} else {
					fmt.Println("msgBuffChan is Closed")
					break
				}
		case <-c.ExitBufChan:
			return
		}
	}
}

// GetConnection 获取连接
func (c *Connection) GetConnection() *net.TCPConn {
	return c.Conn
}

// GetConnectionID 获取连接ID
func (c *Connection) GetConnectionID() uint32 {
	return c.ConnID
}

// GetConnectionAddr 获取客户段地址信息
func (c *Connection) GetConnectionAddr() net.Addr {
	return c.Conn.RemoteAddr()
}

// Stop 关闭连接
func (c *Connection) Stop() {
	if c.isClosed {
		return // 连接已经关闭
	}
	c.isClosed = true

	// 调用 HOOK 方法
	c.TcpServer.CallOnConnStop(c)

	// 关闭socket链接
	c.Conn.Close()

	// 把当前连接对应connManager删除该连接
	c.TcpServer.GetConnMgr().Remove(c)

	//关闭Writer Goroutine
	c.ExitBufChan <- true

	close(c.ExitBufChan)
	close(c.msgChan)

}

// Send 发送数据
func (c *Connection) SendMsg(msgId uint32, data []byte) error {
	if c.isClosed {
		return errors.New("Connection closed when send msg")
	}
	dp := NewDataPack()
	pack, err := dp.Pack(NewMsgPackage(msgId, data))
	if err != nil {
		fmt.Println("Pack error msg id = ", msgId)
		return errors.New("Pack error msg ")
	}

	// 写回客户端
	c.msgChan <- pack //将之前直接回写给conn.Write的方法 改为 发送给Channel 供Writer读取

	return nil
}

func (c *Connection) SendBuffMsg(msgId uint32, data []byte) error {
	if c.isClosed == true {
		return errors.New("Connection closed when send buff msg")
	}
	// 数据封包
	dp := &DataPack{}
	pack, err := dp.Pack(NewMsgPackage(msgId, data))
	if err != nil {
		fmt.Println("Pack error msg id = ", msgId)
		return errors.New("Pack error msg ")
	}

	c.msgBuffChan <- pack
	return nil
}

//设置链接属性
func (c *Connection) SetProperty(key string, value interface{}) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()

	c.property[key] = value
}

//获取链接属性
func (c *Connection) GetProperty(key string) (interface{}, error) {
	c.propertyLock.RLock()
	defer c.propertyLock.RUnlock()

	if value, ok := c.property[key]; ok  {
		return value, nil
	} else {
		return nil, errors.New("no property found")
	}
}

//移除链接属性
func (c *Connection) RemoveProperty(key string) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()

	delete(c.property, key)
}