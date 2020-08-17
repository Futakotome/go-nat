package main

import (
	"flag"
	"os"
)

func main() {
	localPort := flag.Int("localPort", 8080, "客户端访问端口")
	remotePort := flag.Int("remotePort", 8888, "服务端访问端口")
	flag.Parse()
	if flag.NFlag() != 2 {
		flag.PrintDefaults()
		os.Exit(1)
	}
	middleServer := New()
	middleServer.Start(*localPort, *remotePort)
	//循环
	select {}
}
