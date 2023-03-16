package connector

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"sync"

	"github.com/happay/cms-utils-go/util"
	"github.com/jinzhu/gorm"
	"gopkg.in/yaml.v2"
)

var db *gorm.DB
var pgConn sync.Once

// postgres config params
const (
	PostgresHost        = "host"
	PostgresPort        = "port"
	PostgresDBName      = "dbname"
	PostgresSslMode     = "sslmode"
	PostgresDBUser      = "user"
	PostgresPassword    = "password"
	PostgresMaxOpenConn = "max_open_connections"
	PostgresMaxIdleConn = "max_idle_connections"
)

/*
PostgresConfigParams : AVOID the use of same variable
whenever using these constants. DON'T CHANGE THE BELOW ORDER as the parsing of GORM for
line argument is not by key (yeah I know, very weird thing. TODO: test with change in order)
*/
var PostgresConfigParams = []string{
	PostgresHost,
	PostgresPort,
	PostgresDBName,
	PostgresSslMode,
	PostgresDBUser,
	PostgresPassword,
	PostgresMaxOpenConn,
	PostgresMaxIdleConn,
}

// GetPgConn creates the database connection and the *gorm.Db object.
// It doesn't set the db logger. You would need to explicity set it using db.SetLogger(<logger>)
func GetPgConn(dbCredPath string, pgConfigKey string) *gorm.DB {
	pgConn.Do(func() {
		fmt.Println("initiating postgres connection", dbCredPath, pgConfigKey)
		var bytes []byte
		var err error
		//read database yaml configuration file
		if bytes, err = ioutil.ReadFile(dbCredPath); err != nil {
			err = fmt.Errorf("file read error %s: %s", dbCredPath, err)
			fmt.Println(err.Error())
			return
		}
		// map the yaml file dbConfigs map object
		dbConfigs := make(map[string]map[string]string)
		if err = yaml.Unmarshal(bytes, &dbConfigs); err != nil {
			err = fmt.Errorf("error while parsing the database configuration: %s", err)
			fmt.Println(err.Error())
			return
		}
		// get postgres database configuration from the dbConfigs
		if pgDbConfigs, found := dbConfigs[pgConfigKey]; !found {
			err = fmt.Errorf("%s database config not found on yaml file: %s", pgConfigKey, dbCredPath)
			fmt.Println(err.Error())
			return
		} else {
			pgConnStr := createPgConnString(pgDbConfigs)
			if db, err = gorm.Open("postgres", pgConnStr); err != nil {
				err = fmt.Errorf("initialize postgres db connection failed: %s", err)
				fmt.Println(err.Error())
				return
			}
			if db, err = setPgConnLimits(db, pgDbConfigs); err != nil {
				err = fmt.Errorf("error setting connection limits: %s", err)
				fmt.Println(err.Error())
				return
			}
			db.BlockGlobalUpdate(true)
			//db.SetLogger(logger.GetLogger())
			fmt.Println("postgres connection established")
		}
	})
	fmt.Println("return statement")
	return db
}

/*
createPgConnString reads the given database connection config map,
and creates the collated string argument to be used in
creating database connection pool.
*/
func createPgConnString(pgDbConfigs map[string]string) string {
	var connStrBuilder strings.Builder
	for _, val := range PostgresConfigParams {
		if val == PostgresMaxOpenConn || val == PostgresMaxIdleConn { // these will be configured later directly on the connection pool object (client)
			continue
		}
		connStrBuilder.WriteString(val)
		connStrBuilder.WriteString("=")
		connStrBuilder.WriteString(util.GetConfigValue(pgDbConfigs[val]))
		connStrBuilder.WriteString(" ") // delimiter
	}
	return connStrBuilder.String()
}

// setPgConnLimits sets the max number of possible & idle connections in the database connection pool.
func setPgConnLimits(db *gorm.DB, pgDbConfigs map[string]string) (*gorm.DB, error) {
	// setting maximum number of connections that this connection pool can have
	if maxOpenConnCountStr, found := pgDbConfigs[PostgresMaxOpenConn]; found {
		maxOpenConnCount, err := strconv.Atoi(util.GetConfigValue(maxOpenConnCountStr))
		if err != nil {
			err = fmt.Errorf("error setting max open connections: %s", err)
			// logger.GetLogger().Println(err)
			return db, err
		}
		db.DB().SetMaxOpenConns(maxOpenConnCount)
	}
	// setting maximum number of idle (in reserve) connections that this connection pool
	if maxIdleConnCountStr, found := pgDbConfigs[PostgresMaxIdleConn]; found {
		maxIdleConnCount, err := strconv.Atoi(util.GetConfigValue(maxIdleConnCountStr))
		if err != nil {
			err = fmt.Errorf("error setting max idle connections: %s", err)
			//logger.GetLogger().Println(err)
			return db, err
		}
		db.DB().SetMaxIdleConns(maxIdleConnCount)
	}
	return db, nil
}
