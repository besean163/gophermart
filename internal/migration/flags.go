package migration

import (
	"flag"
	"os"
)

type MigrationConfig struct {
	DatabaseDSN string
}

func NewConfig() MigrationConfig {
	config := MigrationConfig{}
	flag.StringVar(&config.DatabaseDSN, "d", "", "data base dsn")
	flag.Parse()

	if DatabaseDSNEnv := os.Getenv("DATABASE_URI"); DatabaseDSNEnv != "" && config.DatabaseDSN == "" {
		config.DatabaseDSN = DatabaseDSNEnv
	}

	return config
}
