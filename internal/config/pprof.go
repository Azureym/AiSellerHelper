package config

import (
	"context"
	"log"
	"net/http"
	_ "net/http/pprof"
	"strconv"
)

const (
	pprofPort = 8888
)

func StartPprof(ctx context.Context) {
	go func() {
		log.Println(http.ListenAndServe(":"+strconv.Itoa(pprofPort), nil))
	}()
	log.Printf("pprof server listening at :%d", pprofPort)
}
