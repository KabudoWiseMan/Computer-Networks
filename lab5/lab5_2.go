package main

import (
	"github.com/goftp/server"
	"github.com/goftp/file-driver"
    "github.com/mgutz/logxi/v1"
)

func main() {
    factory := &filedriver.FileDriverFactory{
		RootPath: "server",
		Perm:     server.NewSimplePerm("user", "group"),
	}
	opts := &server.ServerOpts{
		Factory:  factory,
		Port:     21,
		Hostname: "localhost",
		Auth:     &server.SimpleAuth{Name: "ft****", Password: "***********"},
	}

	server := server.NewServer(opts)
	err := server.ListenAndServe()
	if err != nil {
		log.Error("Error starting server", "error", err)
	}
}