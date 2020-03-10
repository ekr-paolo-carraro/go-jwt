package main

import (
	"github.com/ekr-paolo-carraro/go-jwt/server"
	"github.com/subosito/gotenv"
)

func main() {
	gotenv.Load()
	server.Run()
}
