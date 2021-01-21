package server

import (
	"github.com/tarasova-school/internal/tarasova-school/server/handlers"
	"github.com/tarasova-school/pkg/logger"
	"net/http"
)

func StartServer(handlers *handlers.Handlers, port string) {
	router := NewRouter(handlers)
	logger.LogInfo("Restart server")
	if err := http.ListenAndServe(port, router); err != nil {
		logger.LogFatal(err)
	}
}
