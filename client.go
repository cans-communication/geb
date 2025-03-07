package geb

import (
	"context"
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type PG struct {
	DB *gorm.DB
}

func (pg *PG) Ping(ctx context.Context) error {
	sqlDB, err := pg.DB.
		WithContext(ctx).
		DB()

	if err != nil {
		return err
	}

	err = sqlDB.Ping()

	if err != nil {
		return err
	}

	return nil
}

func (pg *PG) Close(ctx context.Context) error {
	sqlDB, err := pg.DB.
		WithContext(ctx).
		DB()

	if err != nil {
		return err
	}

	err = sqlDB.Close()

	if err != nil {
		return err
	}

	return nil
}

type ConnectConfig struct {
	DBHost     string
	DBPort     int
	DBUser     string
	DBPassword string
	DBName     string
	MaxIdleCon int
}

func Connect(conf ConnectConfig) (*PG, error) {

	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s application_name=xl_pgclient TimeZone=UTC",
		conf.DBHost,
		conf.DBPort,
		conf.DBUser,
		conf.DBPassword,
		conf.DBName,
	)

	db, err := gorm.Open(
		postgres.Open(dsn),
		&gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		},
	)

	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()

	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxIdleConns(conf.MaxIdleCon)

	return &PG{
		DB: db,
	}, nil
}
