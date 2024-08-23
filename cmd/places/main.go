package main

import (
	"fmt"
	. "main/internal/server"
	cfg "main/pkg/config"
	"main/pkg/handler"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	time.Sleep(time.Second * 1)
	cnf := cfg.ReadInConfig()
	server := new(Server)

	h := new(handler.Handler)

	r := h.InitRoutes()

	go func() {
		if err := server.Run(cnf.ServerHost, cnf.ServerPort, r); err != nil {
			fmt.Printf("error occured while running http server: %s\n", string(err.Error()))
			return
		}
	}()

	fmt.Printf("Server started on %s\n", "http://"+cnf.ServerHost+":"+cnf.ServerPort)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	fmt.Println("Shutting down server...")

	defer server.Shutdown(nil)
}
