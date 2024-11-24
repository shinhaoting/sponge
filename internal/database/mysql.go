package database

import (
	"time"

	"github.com/zhufuyi/sponge/pkg/logger"
	"github.com/zhufuyi/sponge/pkg/sgorm"
	"github.com/zhufuyi/sponge/pkg/sgorm/mysql"
	"github.com/zhufuyi/sponge/pkg/utils"

	"github.com/zhufuyi/sponge/internal/config"
)

// InitMysql connect mysql
func InitMysql() *sgorm.DB {
	mysqlCfg := config.Get().Database.Mysql
	opts := []mysql.Option{
		mysql.WithMaxIdleConns(mysqlCfg.MaxIdleConns),
		mysql.WithMaxOpenConns(mysqlCfg.MaxOpenConns),
		mysql.WithConnMaxLifetime(time.Duration(mysqlCfg.ConnMaxLifetime) * time.Minute),
	}
	if mysqlCfg.EnableLog {
		opts = append(opts,
			mysql.WithLogging(logger.Get()),
			mysql.WithLogRequestIDKey("request_id"),
		)
	}

	if config.Get().App.EnableTrace {
		opts = append(opts, mysql.WithEnableTrace())
	}

	// setting mysql slave and master dsn addresses
	//opts = append(opts, mysql.WithRWSeparation(
	//	mysqlCfg.SlavesDsn,
	//	mysqlCfg.MastersDsn...,
	//))

	// add custom gorm plugin
	//opts = append(opts, mysql.WithGormPlugin(yourPlugin))

	dsn := utils.AdaptiveMysqlDsn(mysqlCfg.Dsn)
	db, err := mysql.Init(dsn, opts...)
	if err != nil {
		panic("init mysql error: " + err.Error())
	}
	return db
}