package connector

import (
	"fmt"
	"strconv"
	"time"

	"github.com/happay/cms-utils-go/v3/logger"
	"github.com/jinzhu/gorm"

	"github.com/go-sql-driver/mysql"
)

var mySqlDb *gorm.DB

const (
	MySqlConfigKey = "mysql"
)

func GetMySqlConn(mySQLDbConfigs map[string]string) *gorm.DB {
	logger.GetLoggerV3().Info("initiating db connection")
	var err error
	connStr := formatConnString(mySQLDbConfigs)
	if mySqlDb, err = gorm.Open("mysql", connStr); err != nil {
		err = fmt.Errorf("initialize db connection failed: %s", err)
		logger.GetLoggerV3().Error(err.Error())
		return mySqlDb
	}
	if mySqlDb, err = setConnLimits(mySqlDb, mySQLDbConfigs); err != nil {
		err = fmt.Errorf("error setting connection limits: %s", err)
		logger.GetLoggerV3().Error(err.Error())
		return mySqlDb
	}

	mySqlDb.BlockGlobalUpdate(true)
	logger.GetLoggerV3().Info("db connection established")
	return mySqlDb
}

func formatConnString(mySQLDbConfigs map[string]string) string {
	config := mysql.Config{
		User:                 mySQLDbConfigs["user"],
		Passwd:               mySQLDbConfigs["password"],
		Net:                  "tcp",
		Addr:                 mySQLDbConfigs["host"],
		DBName:               mySQLDbConfigs["dbname"],
		ParseTime:            true,
		AllowNativePasswords: true,
		Loc:                  time.UTC,
	}
	return config.FormatDSN()
}

func setConnLimits(db *gorm.DB, pgDbConfigs map[string]string) (*gorm.DB, error) {
	// setting maximum number of connections that this connection pool can have
	if maxOpenConnCountStr, found := pgDbConfigs["max_open_connections"]; found {
		maxOpenConnCount, err := strconv.Atoi(maxOpenConnCountStr)
		if err != nil {
			err = fmt.Errorf("error setting max open connections: %s", err)
			return db, err
		}
		db.DB().SetMaxOpenConns(maxOpenConnCount)
	}
	// setting maximum number of idle (in reserve) connections that this connection pool
	if maxIdleConnCountStr, found := pgDbConfigs["max_idle_connections"]; found {
		maxIdleConnCount, err := strconv.Atoi(maxIdleConnCountStr)
		if err != nil {
			err = fmt.Errorf("error setting max idle connections: %s", err)
			return db, err
		}
		db.DB().SetMaxIdleConns(maxIdleConnCount)
	}
	return db, nil
}
