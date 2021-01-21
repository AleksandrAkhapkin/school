package main

import (
	"flag"
	"fmt"
	"github.com/tarasova-school/internal/clients/postgres"
	"github.com/tarasova-school/internal/tarasova-school/server"
	"github.com/tarasova-school/internal/tarasova-school/server/handlers"
	"github.com/tarasova-school/internal/tarasova-school/service"
	"github.com/tarasova-school/internal/types/config"
	"github.com/tarasova-school/pkg/logger"
	"gopkg.in/yaml.v3"
	"os"
)

func main() {

	configPath := new(string)
	debug := new(bool)
	flag.StringVar(configPath, "config-path", "configs/config.yaml", "path to yaml config file")
	flag.BoolVar(debug, "debug", true, "debug-alarm")
	flag.Parse()
	f, err := os.Open(*configPath)
	if err != nil {
		logger.LogFatal(fmt.Errorf("err with open config file %v, %s", err, *configPath))
	}
	logger.Debug = *debug
	cnf := config.Config{}
	if err = yaml.NewDecoder(f).Decode(&cnf); err != nil {
		logger.LogFatal(fmt.Errorf("err with parse config %v, %s", err, *configPath))
	}

	pg, err := postgres.NewPostgres(cnf.PostgresDsn)
	if err != nil {
		logger.LogFatal(err)
	}

	srv, err := service.NewService(pg, &cnf)
	if err != nil {
		logger.LogFatal(err)
	}

	err = logger.NewLogger(cnf.Telegram)
	if err != nil {
		logger.LogFatal(err)
	}

	handls := handlers.NewHandlers(srv, &cnf)
	logger.CheckDebug()
	server.StartServer(handls, cnf.ServerPort)
}
