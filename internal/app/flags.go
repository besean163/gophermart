package app

import (
	"flag"
	"os"
)

type AppConfig struct {
	RunAddress        string
	RunAccrualAddress string
	DatabaseDSN       string
	HashSecret        string
}

func NewConfig() AppConfig {
	config := AppConfig{}
	flag.StringVar(&config.RunAddress, "a", "", "server run port")
	flag.StringVar(&config.RunAccrualAddress, "r", "", "accrual run port")
	flag.StringVar(&config.DatabaseDSN, "d", "", "data base dsn")
	flag.StringVar(&config.HashSecret, "k", "secret", "hash secret")
	flag.Parse()

	if runAddressEnv := os.Getenv("RUN_ADDRESS"); runAddressEnv != "" && config.RunAddress == "" {
		config.RunAddress = runAddressEnv
	}

	if runAccrualAddressEnv := os.Getenv("ACCRUAL_SYSTEM_ADDRESS"); runAccrualAddressEnv != "" && config.RunAccrualAddress == "" {
		config.RunAccrualAddress = runAccrualAddressEnv
	}

	if DatabaseDSNEnv := os.Getenv("DATABASE_URI"); DatabaseDSNEnv != "" && config.DatabaseDSN == "" {
		config.DatabaseDSN = DatabaseDSNEnv
	}

	if HashSecretEnv := os.Getenv("HASH_SECRET"); HashSecretEnv != "" && config.HashSecret == "" {
		config.HashSecret = HashSecretEnv
	}

	return config
}
