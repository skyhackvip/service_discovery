package main

import (
	"context"
	"flag"
	"github.com/skyhackvip/service_discovery/api"
	"github.com/skyhackvip/service_discovery/configs"
	"github.com/skyhackvip/service_discovery/global"
	"github.com/skyhackvip/service_discovery/model"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	//init config
	c := flag.String("c", "", "config file path")
	flag.Parse()
	config, err := configs.LoadConfig(*c)
	if err != nil {
		log.Println("load config error:", err)
		return
	}

	//global discovery
	global.Discovery = model.NewDiscovery(config)

	//init router and start server
	router := api.InitRouter()
	srv := &http.Server{
		Addr:    config.HttpServer,
		Handler: router,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen:%s\n", err)
		}
	}()

	//graceful restart
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)
	<-quit
	log.Println("shutdown discovery server...")
	//cancel
	global.Discovery.CancelSelf()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("server shutdown error:", err)
	}
	select {
	case <-ctx.Done():
		log.Println("timeout of 5 seconds")
	}
	log.Println("server exiting")
}
