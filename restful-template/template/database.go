
package






import (

	"log"
	"github.com/jinzhu/gorm"
)

func DatabaseInit(connStr string, maxOpen, maxIdle int, debugMode bool) {
	var err error
	db, err = gorm.Open("mysql", connStr)
	if err != nil {
		logs.Critical("connect to mysql fails: %s", err.Error())
		panic("database error")
	}

	if debugMode {
		db.LogMode(true)
		db.SetLogger(log.New(logs.GetBeeLogger(), "[GORM] ", 0))
	}

	db.SingularTable(true)
	db.DB().SetMaxOpenConns(maxOpen)
	db.DB().SetMaxIdleConns(maxIdle)
}

var db *gorm.DB

func DB(t string) *gorm.DB {
	return db
}
