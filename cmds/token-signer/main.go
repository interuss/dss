package main

import (
	"token-signer/httpServer"
	"token-signer/tokenSigner"
)

type Server interface {
	Serve(signer tokenSigner.TokenSigner)
}

func main() {
	var server Server = httpServer.HttpServer{}
	server.Serve(tokenSigner.New())
}
