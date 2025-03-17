package main

import (
	Registry "github.com/AminCoder/Go-DyFunc/internal/registery"
	"github.com/AminCoder/Go-DyFunc/pkg/server"
)

func main() {
	reg := Registry.New_Registry()
	reg.Add("sum", func(x int, y int) int {
		return x + y
	})
	server.Run_HTTP_Server(":5001", "/call-remote", reg)

}
