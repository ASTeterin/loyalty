package config

import (
	"flag"
	"os"
)

type Config struct {
	AppAddr   string
	DBConnStr string
}

func ParseFlags() Config {
	var appAddr string
	var dbConnectionString string
	flag.StringVar(&appAddr, "a", ":8080", "port to run server")
	flag.StringVar(&dbConnectionString, "d", "postgres://admin:1234@localhost:5432/loyalty?sslmode=disable", "database DSN")
	flag.Parse()

	if envAppAddr, exist := os.LookupEnv("RUN_ADDRESS"); exist {
		appAddr = envAppAddr
	}

	if envDBConnectionStr, exist := os.LookupEnv("DATABASE_URI"); exist {
		dbConnectionString = envDBConnectionStr
	}

	return Config{
		AppAddr:   appAddr,
		DBConnStr: dbConnectionString,
	}
}
