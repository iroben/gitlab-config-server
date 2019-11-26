package main

import (
	"database/sql"
	"fmt"
	"log"
	"path/filepath"
	"gitlab-config-server/config"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/mattes/migrate"
	"github.com/mattes/migrate/database/mysql"
	_ "github.com/mattes/migrate/source/file"
)

type MLog struct {
}

func (MLog) Printf(format string, v ...interface{}) {
	fmt.Printf(format, v...)
}
func (MLog) Verbose() bool {
	return true
}
func main() {
	dbConfig := config.GetDbConfig()
	db, err := sql.Open("mysql",
		fmt.Sprintf("%s:%s@tcp(%s)/%s?multiStatements=true",
			dbConfig.User,
			dbConfig.Passwd,
			dbConfig.Host,
			dbConfig.Name,
		))
	if err != nil {
		log.Fatal(err)
		return
	}
	driver, err := mysql.WithInstance(db, &mysql.Config{
		DatabaseName: dbConfig.Name,
	})
	if err != nil {
		log.Fatal(err)
		return
	}
	realPath, err := filepath.Abs("migration")
	if err != nil {
		log.Fatal(err)
		return
	}
	realPath = strings.Replace(realPath, "\\", "/", -1)
	m, err := migrate.NewWithDatabaseInstance(
		"file://"+realPath,
		"mysql",
		driver,
	)
	if err != nil {
		log.Fatal(err)
		return
	}
	var mLog MLog
	m.Log = mLog
	m.Up()
}
