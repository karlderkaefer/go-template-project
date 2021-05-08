package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"os/signal"
	"path"
	"runtime"
	"strings"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
)

var Version string

func main() {
	Execute()
}

func startServer() {
	initLogger()
	router := gin.New()
	router.GET("/health", healthCheck)
	router.GET("/version", version)
	server := &http.Server{
		Addr:    ":8000",
		Handler: router,
	}
	go func() {
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()
	gracefulShutdownHandler(server)
}

func initLogger() {
	log.SetFormatter(&log.JSONFormatter{
		FieldMap: log.FieldMap{
			log.FieldKeyLevel: "go_level",
			log.FieldKeyFile:  "go_source",
			log.FieldKeyFunc:  "go_source_call",
		},
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			s := strings.Split(f.Function, ".")
			funcName := s[len(s)-1]
			return funcName, fmt.Sprintf("%s:%d", path.Base(f.File), f.Line)
		},
	})
	log.SetReportCaller(true)
}

func healthCheck(c *gin.Context) {
	ok := gin.H{"status": "ok"}
	c.JSON(http.StatusOK, ok)
}

func version(c *gin.Context) {
	version := gin.H{"version": Version}
	c.JSON(http.StatusOK, version)
}

func gracefulShutdownHandler(srv *http.Server) {
	quit := make(chan os.Signal, 3)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	<-quit
	log.Info("Detected shutdown signal. Gracefully shutting down..\n")
	// The context is used to inform the server it has 20 seconds to finish
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	err := srv.Shutdown(ctx)
	if err != nil {
		log.Fatal("Server failed to shutdown:", err)
	}
	log.Info("Server exiting")
}
