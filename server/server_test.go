package server_test

import (
	"net/http"
	"testing"
	"time"
	"todo-app-api/server"

	"github.com/sirupsen/logrus"
)

func TestServer(t *testing.T) {
	httpHandler := http.NewServeMux()

	srv := server.NewServer(logrus.New(), httpHandler, "9091")
	srv.Start()
	time.Sleep(time.Second * 1)
	srv.Close()
}
