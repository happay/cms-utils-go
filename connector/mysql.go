package connector

import (
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/happay/cms-utils-go/v2/logger"
	"github.com/happay/cms-utils-go/v2/util"
	"github.com/jinzhu/gorm"
	"gopkg.in/yaml.v2"

	"github.com/go-sql-driver/mysql"
)

var mySqlDb *gorm.DB
var myConn sync.Once

const (
	MySqlConfigKey = "mysql"
)

func GetMySqlConn(dbCredPath string, configKey string) *gorm.DB {
	myConn.Do(func() {
		logger.GetLoggerV3().Info("initiating db connection", dbCredPath, configKey)
		var bytes []byte
		var err error
		// read database yaml configuration file
		if bytes, err = os.ReadFile(dbCredPath); err != nil {
			err = fmt.Errorf("file read error %s: %s", dbCredPath, err)
			logger.GetLoggerV3().Error(err.Error())
			return
		}

		// map the yaml file dbConfigs map object
		dbConfigs := make(map[string]map[string]string)
		if err = yaml.Unmarshal(bytes, &dbConfigs); err != nil {
			err = fmt.Errorf("error while parsing the database configuration: %s", err)
			logger.GetLoggerV3().Error(err.Error())
			return
		}

		if mySQLDbConfigs, found := dbConfigs[configKey]; !found {
			err = fmt.Errorf("%s database config not found on yaml file: %s", configKey, dbCredPath)
			logger.GetLoggerV3().Error(err.Error())
			return
		} else {
			connStr := formatConnString(mySQLDbConfigs)
			if mySqlDb, err = gorm.Open("mysql", connStr); err != nil {
				err = fmt.Errorf("initialize db connection failed: %s", err)
				logger.GetLoggerV3().Error(err.Error())
				return
			}
			if mySqlDb, err = setConnLimits(mySqlDb, mySQLDbConfigs); err != nil {
				err = fmt.Errorf("error setting connection limits: %s", err)
				logger.GetLoggerV3().Error(err.Error())
				return
			}

			mySqlDb.BlockGlobalUpdate(true)
			logger.GetLoggerV3().Info("db connection established")
		}
	})
	return mySqlDb
}

func formatConnString(mySQLDbConfigs map[string]string) string {
	config := mysql.Config{
		User:      util.GetConfigValue(mySQLDbConfigs["user"]),
		Passwd:    util.GetConfigValue(mySQLDbConfigs["password"]),
		Net:       "tcp",
		Addr:      util.GetConfigValue(mySQLDbConfigs["host"]),
		DBName:    util.GetConfigValue(mySQLDbConfigs["dbname"]),
		ParseTime: true,
		Loc:       time.UTC,
	}
	return config.FormatDSN()
}

func setConnLimits(db *gorm.DB, pgDbConfigs map[string]string) (*gorm.DB, error) {
	// setting maximum number of connections that this connection pool can have
	if maxOpenConnCountStr, found := pgDbConfigs["max_open_connections"]; found {
		maxOpenConnCount, err := strconv.Atoi(util.GetConfigValue(maxOpenConnCountStr))
		if err != nil {
			err = fmt.Errorf("error setting max open connections: %s", err)
			return db, err
		}
		db.DB().SetMaxOpenConns(maxOpenConnCount)
	}
	// setting maximum number of idle (in reserve) connections that this connection pool
	if maxIdleConnCountStr, found := pgDbConfigs["max_idle_connections"]; found {
		maxIdleConnCount, err := strconv.Atoi(util.GetConfigValue(maxIdleConnCountStr))
		if err != nil {
			err = fmt.Errorf("error setting max idle connections: %s", err)
			return db, err
		}
		db.DB().SetMaxIdleConns(maxIdleConnCount)
	}
	return db, nil
}
