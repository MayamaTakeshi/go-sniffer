package main

import (
	"net/http"
	_ "net/http/pprof"

	"go-sniffer/core"
)

func profiling() {
	http.ListenAndServe("0.0.0.0:2022", nil)
}

func main() {
	//go profiling()

	core := core.New()
	core.Run()
}
