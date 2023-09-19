package znet

import (
	"fmt"
	"net"
	"zinx/utils"
	"zinx/ziface"
)

// Server IServer 的接口实现
type Server struct {
	// 服务器名称
	Name string
	// 服务器绑定 IP 的版本
	IPVersion string
	// 服务器监听的IP
	IP string
	// 服务器的端口
	Port int
	//当前Server的消息管理模块，用来绑定MsgId和对应的处理方法
	MsgHandler ziface.IMsgHandle
	//当前Server的链接管理器
	ConnMgr ziface.IConnManager

	//该Server的连接创建时Hook函数
	OnConnStart	func(conn ziface.IConnection)
	//该Server的连接断开时的Hook函数
	OnConnStop func(conn ziface.IConnection)
}

// NewServer 初始化 Server
func NewServer() ziface.IServer {
	utils.GlobalObject.Reload()  // 从全局配置中初始化
	return &Server{
		Name:      utils.GlobalObject.Name,
		IPVersion: "tcp4",
		IP:        utils.GlobalObject.Host,
		Port:      utils.GlobalObject.TcpPort,
		MsgHandler: NewMsgHandle(),
		ConnMgr: NewConnManager(),
	}
}

// Start 启动服务器方法
func (s *Server) Start() {
	fmt.Printf("[START] Server name: %s,listenner at IP: %s, Port %d is starting\n", s.Name, s.IP, s.Port)
	fmt.Printf("[Zinx] Version: %s, MaxConn: %d,  MaxPacketSize: %d\n",
		utils.GlobalObject.Version,
		utils.GlobalObject.MaxConn,
		utils.GlobalObject.MaxPacketSize)  // 校验全局配置是否加载成功

	// 避免由于协程的阻塞上升到主线程阻塞的地步
	go func() {
		//0 启动worker工作池机制
		s.MsgHandler.StartWorkerPool()

		// 1. 获取 TCP 的 ADDR
		addr, err := net.ResolveTCPAddr(s.IPVersion, fmt.Sprintf("%s:%d", s.IP, s.Port))
		if err != nil {
			fmt.Println("resolve tcp addr error: ", err)
			return
		}

		// 2. 监听服务器地址
		listener, err := net.ListenTCP(s.IPVersion, addr)
		if err != nil {
			fmt.Println("listen ip addr error: ", err)
			return
		}

		fmt.Println("Start Zinx server success: ", s.Name, "Listening...")

		// 3. 阻塞等待客户端连接, 处理客户端连接的业务数据
		// 拥有自增 id, 避免连接 id 重复
		var id uint32
		id = 0
		for {
			//3.1 阻塞等待客户端建立连接请求
			conn, err := listener.AcceptTCP()
			if err != nil {
				fmt.Println("Accept error: ", err)
				continue
			}
			//3.2 Server.Start() 设置服务器最大连接控制,如果超过最大连接，那么则关闭此新的连接
			if s.GetConnMgr().Len() > utils.GlobalObject.MaxConn {
				conn.Close()
				continue
			}

			//3.3 处理该新连接请求的 业务 方法， 此时应该有 handler 和 conn是绑定的
			dealConn := NewConnection(s, conn, id, s.MsgHandler)
			id++

			go dealConn.Start()

		}
	}()
}

//============== 实现 ziface.IServer 里的全部接口方法 ========

// Stop 停止服务器方法
func (s *Server) Stop() {
	fmt.Println("[STOP] Zinx server , name " , s.Name)

	//将其他需要清理的连接信息或者其他信息 也要一并停止或者清理
	s.ConnMgr.ClearConn()
}

// Serve 开启业务的方法
func (s *Server) Serve() {
	s.Start()

	// TODO 扩展额外业务

	// 阻塞线程
	select {}
}

// AddRouter 路由功能：给当前服务注册一个路由业务方法，供客户端链接处理使用
func (s *Server)AddRouter(msgId uint32, router ziface.IRouter) {
	s.MsgHandler.AddRouter(msgId, router)

	fmt.Println("Add Router succ! " )
}

// GetConnMgr 得到链接管理
func (s *Server) GetConnMgr() ziface.IConnManager {
	return s.ConnMgr
}

func (s *Server) SetOnConnStart(f func(ziface.IConnection)) {
	s.OnConnStart = f
}

func (s *Server) SetOnConnStop(f func(ziface.IConnection)) {
	s.OnConnStop = f
}

func (s *Server) CallOnConnStart(conn ziface.IConnection) {
	if s.OnConnStart != nil {
		fmt.Println("---> CallOnConnStart....")
		s.OnConnStart(conn)
	}
}

func (s *Server) CallOnConnStop(conn ziface.IConnection) {
	if s.OnConnStop != nil {
		fmt.Println("---> CallOnConnStop....")
		s.OnConnStop(conn)
	}
}
