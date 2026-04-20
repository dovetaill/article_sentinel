package database

import (
	"fmt"
	"strings"

	"github.com/dovetaill/article-sentinel/pkg/config"
	"gorm.io/gorm"
)

const (
	driverMySQL    = "mysql"
	driverPostgres = "postgres"
)

func openPrimaryDatabase(cfg *config.Config) (*gorm.DB, error) {
	driver := normalizedDriver(cfg.Database.Driver)
	switch driver {
	case driverMySQL:
		return openMySQLFn(cfg.Database.MySQL)
	case driverPostgres:
		return openPostgresFn(cfg.Database.Postgres)
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", cfg.Database.Driver)
	}
}

func normalizedDriver(driver string) string {
	if strings.TrimSpace(driver) == "" {
		return driverMySQL
	}
	return strings.ToLower(strings.TrimSpace(driver))
}
