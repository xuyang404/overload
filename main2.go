package main

import (
	"context"
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	err      error
	server   *http.Server
	listener net.Listener = nil

	graceful = flag.Bool("graceful", false, "listen on fd open 3 (internal use only)")
)

func hello2(w http.ResponseWriter, r *http.Request) {
	time.Sleep(5 * time.Second)
	w.Write([]byte("hello word"))
}

func main() {
	flag.Parse()

	http.HandleFunc("/hello2", hello2)
	server = &http.Server{
		Addr: ":7890",
	}

	//设置监听器的监听对象（新建的或已存在的socket描述符）
	if *graceful {
		//子进程监听父进程传递的socket描述符
		log.Println("监听现有文件描述符 3")
		//子进程的0，1，2是预留给标准输入、标准输出、错误输出，故传递的socket描述符应放在子进程的3
		f := os.NewFile(3, "")
		listener, err = net.FileListener(f)
	} else {
		//父进程监听新建的socket描述符
		log.Panicln("监听新的文件描述符")
		listener, err = net.Listen("tcp", server.Addr)
	}

	go func() {
		err = server.Serve(listener)
		log.Printf("server.Serve err: %v\n", err)
	}()

	//监听信号
	handleSignal2()
	log.Panicln("signal end")
}

func handleSignal2() {
	ch := make(chan os.Signal, 1)

	//监听信号
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR2)

	for {
		sig := <-ch
		log.Printf("signal receive: %v\n", sig)
		ctx, _ := context.WithTimeout(context.Background(), 20*time.Second)
		switch sig {
		case syscall.SIGINT, syscall.SIGTERM: //终止进程执行
			log.Println("shutdown")
			signal.Stop(ch)
			server.Shutdown(ctx)
			log.Println("graceful shutdown")
			return
		case syscall.SIGUSR2:
			log.Println("reload")
			err := reload()
			if err != nil { //执行热重启函数
				log.Fatalf("graceful reload error: %v", err)
			}
			server.Shutdown(ctx)
			log.Println("graceful reload")
			return
		}
	}
}

func reload() error {
	return nil
}
