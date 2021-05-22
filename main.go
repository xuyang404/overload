package main

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

func hello(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hello,word1231231312"))
}

var stdout, _ = os.OpenFile("stdout.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
var stderr, _ = os.OpenFile("stderr.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)

//kill -12 pid可以重启
func main() {

	defer stdout.Close()
	defer stderr.Close()

	http.HandleFunc("/", hello)
	server := &http.Server{
		Addr:    ":7890",
		Handler: nil,
	}
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("listen error:%v\n", err)
		}
	}()
	//deamo()



	handleSig(server)
	fmt.Println(123123)
}

//守护进程
func deamo() {
	//记录pid
	ioutil.WriteFile("reload.pid", []byte(fmt.Sprintf("%d", os.Getpid())), 0777)
	//判断当前是否是子进程，当父进程退出后，子进程会被1号进程接管
	if os.Getppid() != 1 {
		startNewProcess()
		os.Exit(0)
	}
}

func handleSig(server *http.Server) {
	fmt.Println("pid", os.Getpid())
	time.Sleep(1 * time.Second) //睡一下才能挂起

	sign := make(chan os.Signal)
	signal.Notify(sign, syscall.SIGUSR2)
	fmt.Println("handleSig")
	sig := <-sign
	fmt.Println("sig", sig)
	switch sig {
	case syscall.SIGUSR2:
		log.Println("reload")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Fatal("Server Shutdown:", err)
		}

		if _,err := startNewProcess(); err != nil {
			log.Printf("[ service reload ] error: %v\n", err)
			return
		}

		log.Println("[ service reload ] success")
		return
	default:
		syscall.Kill(os.Getpid(), syscall.SIGKILL)

	}

}

//func startNewProcess() error {
//	filePath, err := filepath.Abs(os.Args[0]) //将命令行参数中执行文件路径转换为可用路径
//	if err != nil {
//		fmt.Println(err)
//		return err
//	}
//	cmd := exec.Command(filePath, os.Args[1:]...) //将其他参数传入给新创建的进程
//	//给新进程设置文件描述符，可以重定向到文件中
//	cmd.Stdin = os.Stdin
//	cmd.Stdout = stdout
//	cmd.Stderr = stderr
//	return cmd.Start() //开始执行新进程
//}

func startNewProcess() (pid int, err error) {
	args := os.Args
	file := args[0]

	if tmp, _ := ioutil.TempDir("", ""); tmp != "" {
		tmp := filepath.Dir(tmp)
		if strings.Contains(file, tmp) {
			return 0, errors.New("temporary program does not support startup")
		}
	}

	execSpec := &syscall.ProcAttr{
		Env:   os.Environ(),
		Files: []uintptr{os.Stdin.Fd(), os.Stdout.Fd(), os.Stderr.Fd()},
	}

	return syscall.ForkExec(file, args, execSpec)
}
