package database

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"os"
)

type Database interface {
	GetDB() *gorm.DB
}

type mariaDB struct {
	db *gorm.DB
}

func (db *mariaDB) GetDB() *gorm.DB {
	return db.db
}

func ConnectToDB() (Database, error) {
    dbName := os.Getenv("DB_NAME")
    dbHost := os.Getenv("DB_HOST")
    dbUser := os.Getenv("DB_USER")
    dbPass := os.Getenv("DB_PASS")
    dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", dbUser, dbPass, dbHost, dbName)
    gormDB, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
    if err != nil {
        return nil, err
    }
    return &mariaDB{db: gormDB}, nil
}