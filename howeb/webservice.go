package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/cocobao/howfw/howeb/conf"
	"github.com/cocobao/howfw/howeb/router"
	"github.com/facebookgo/grace/gracehttp"
)

func main() {
	err := gracehttp.Serve(
		&http.Server{
			Addr:    conf.GCfg.LocalPort,
			Handler: router.LoadRouter(),
		},
	)
	if err != nil {
		fmt.Println(err, "setup api service fail")
		os.Exit(0)
	}
}
