package cmd

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/pustato/otus_home_work/hw12_13_14_15_calendar/internal/logger"
	"github.com/pustato/otus_home_work/hw12_13_14_15_calendar/internal/storage"
	memorystorage "github.com/pustato/otus_home_work/hw12_13_14_15_calendar/internal/storage/memory"
	sqlstorage "github.com/pustato/otus_home_work/hw12_13_14_15_calendar/internal/storage/sql"
)

type CleanUpFunc = func()

func requireConfig(configPath string) *Config {
	c, err := os.Open(configPath)
	if err != nil {
		log.Fatalln("cannot read config file:", configPath)
	}
	config, err := NewConfig(c)
	if err != nil {
		log.Fatalln("config creating error:", err)
	}

	return config
}

func requireLogger(config LoggerConf) (logger.Logger, CleanUpFunc) {
	logg, err := logger.New(config.Level, config.Target, config.Encoding)
	if err != nil {
		log.Fatalln("cannot create logger:", err)
	}

	return logg, func() {
		_ = logg.Close()
	}
}

func requireEventRepo(config StorageConf) (storage.EventStorage, CleanUpFunc) {
	if config.Driver == "memory" {
		return memorystorage.New(), func() {}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	sqlStorage := sqlstorage.New()
	if err := sqlStorage.Connect(ctx, config.dbConnectionString()); err != nil {
		log.Fatalln("cannot create event repository:", err)
	}
	defer cancel()

	return sqlStorage, func() {
		_ = sqlStorage.Close()
	}
}
