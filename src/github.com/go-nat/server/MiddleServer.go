package main

import (
	"fmt"
	"net"
)

type MiddleServer struct {
	clients      *net.TCPListener //客戶端監聽
	transfers    *net.TCPListener //服務端監聽
	channels     map[int]*Channel
	curChannelId int
}
type Channel struct {
	id              int
	client          net.Conn    //客户端
	transfer        net.Conn    //服务端
	clientRecvMsg   chan []byte //客户端接收得信息
	transferSendMsg chan []byte //服务端发送得消息
}

//创建一个服务器
func New() *MiddleServer {
	return &MiddleServer{
		channels:     make(map[int]*Channel),
		curChannelId: 0,
	}
}

//启动服务
func (m *MiddleServer) Start(clientPort int, transferPort int) {
	addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf(":%d", clientPort))
	if err != nil {
		return
	}
	m.clients, err = net.ListenTCP("tcp", addr)
	if err != nil {
		return
	}
	addr, err = net.ResolveTCPAddr("tcp", fmt.Sprintf(":%d", transferPort))
	if err != nil {
		return
	}
	m.transfers, err = net.ListenTCP("tcp", addr)
	if err != nil {
		return
	}
	go m.AcceptLoop()
	return
}

func (m *MiddleServer) Stop() {
	m.clients.Close()
	m.transfers.Close()
	//循环关闭通道连接
	for _, v := range m.channels {
		v.client.Close()
		v.transfer.Close()
	}
}

//删除通道
func (m *MiddleServer) DeleteChannel(id int) {
	chs := m.channels
	delete(chs, id)
	m.channels = chs
}

//处理连接
func (m *MiddleServer) AcceptLoop() {
	transfer, err := m.transfers.Accept()
	if err != nil {
		return
	}
	for {
		client, err := m.clients.Accept()
		if err != nil {
			continue
		}
		ch := &Channel{
			id:              m.curChannelId,
			client:          client,
			transfer:        transfer,
			clientRecvMsg:   make(chan []byte),
			transferSendMsg: make(chan []byte),
		}
		m.curChannelId++

		chs := m.channels
		chs[ch.id] = ch
		m.channels = chs

		//处理客户端消息
		go m.ClientMsgLoop(ch)
		//处理服务端消息
		go m.TransferMsgLoop(ch)
		//处理服务端和客户端得消息
		go m.MsgLoop(ch)
	}
}

//处理客户端消息
func (m *MiddleServer) ClientMsgLoop(ch *Channel) {
	defer func() {
		fmt.Println("ClientMsgLoop exit ..")
	}()

	for {
		select {
		case data, isClose := <-ch.transferSendMsg:
			{
				if !isClose {
					return
				}
				_, err := ch.client.Write(data)
				if err != nil {
					return
				}
			}
		}
	}
}

//处理服务端消息
func (m *MiddleServer) TransferMsgLoop(ch *Channel) {
	defer func() {
		fmt.Println("TransferMsgLoop exit..")
	}()
	for {
		select {
		case data, isClose := <-ch.clientRecvMsg:
			{
				if !isClose {
					return
				}
				_, err := ch.transfer.Write(data)
				if err != nil {
					return
				}
			}
		}
	}
}

//客户端和服务端消息处理
func (m *MiddleServer) MsgLoop(ch *Channel) {
	defer func() {
		//关闭channel
		close(ch.clientRecvMsg)
		close(ch.transferSendMsg)
		m.DeleteChannel(ch.id)
		fmt.Println("Msg Loop exit..")
	}()
	buf := make([]byte, 1024)
	for {
		n, err := ch.client.Read(buf)
		if err != nil {
			return
		}
		ch.clientRecvMsg <- buf[:n]
		n, err = ch.transfer.Read(buf)
		if err != nil {
			return
		}
		ch.transferSendMsg <- buf[:n]
	}
}
