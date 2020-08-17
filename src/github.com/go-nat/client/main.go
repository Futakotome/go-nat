package main

import (
	"flag"
	"fmt"
	"net"
	"os"
)

func handler(addr net.Conn, localPort int) {
	buf := make([]byte, 1024)
	for {
		//从远程读取数据
		n, err := addr.Read(buf)
		if err != nil {
			continue
		}
		data := buf[:n]
		local, err := net.Dial("tcp", fmt.Sprintf(":%d", localPort))
		if err != nil {
			continue
		}
		//向80服务写数据
		n, err = local.Write(data)
		if err != nil {
			continue
		}
		//读取80服务返回得数据
		n, err = local.Read(buf)
		//关闭80服务，http服务不是持久连接
		//一个请求结束，就会自动断开，所以在for循环里需要不断得dial申请连接，然后关闭
		local.Close()
		if err != nil {
			continue
		}
		data = buf[:n]
		//向远程写数据
		n, err = addr.Write(data)
		if err != nil {
			continue
		}
	}
}

func main() {
	host := flag.String("host", "127.0.0.1", "服务器地址")
	remotePort := flag.Int("remotePort", 8888, "服务器端口")
	localPort := flag.Int("localPort", 80, "本地端口")
	flag.Parse()
	if flag.NFlag() != 3 {
		flag.PrintDefaults()
		os.Exit(1)
	}
	remote, err := net.Dial("tcp", fmt.Sprintf("%s:%d", *host, *remotePort))
	if err != nil {
		fmt.Println(err)
	}
	go handler(remote, *localPort)
	select {}
}
