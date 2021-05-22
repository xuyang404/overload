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

func main() {


			fmt.Println("123123")
			time.Sleep(1 * time.Second)


	overload(nil)
}

func overload(server *http.Server) {
	fmt.Println("pid is ", os.Getpid())
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGUSR2)
	sign := <-quit
	switch sign {
	case syscall.SIGUSR2:
		log.Println("reload")
		pid, err := runNewProcess()
		if err != nil {
			log.Printf("[ service reload ] error: %v\n", err)
			return
		}

		if server != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			if err := server.Shutdown(ctx); err != nil {
				log.Fatal("Server Shutdown:", err)
			}
		}

		log.Println("[ service reload ] success, pid is ", pid)
		return
	default:
		if server != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()
			if err := server.Shutdown(ctx); err != nil {
				log.Fatal("Server Shutdown:", err)
			}
		}
		log.Println("shutdown")
		return
	}
}

func runNewProcess() (pid int, err error) {
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
