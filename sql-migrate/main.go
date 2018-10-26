package main

import (
	"fmt"

	"database/sql"

	_ "github.com/go-sql-driver/mysql"
	"github.com/rubenv/sql-migrate"
)

func main() {
	migrations := &migrate.FileMigrationSource{
		Dir: "/Users/andy/GoLang/src/Hello/migrate",
	}

	fmt.Println(migrations)
	db, err := sql.Open("mysql", "root:123456@tcp(127.0.0.1:3306)/goods?parseTime=true")
	if err != nil {
		panic(err)
	}
	n, err := migrate.Exec(db, "mysql", migrations, migrate.Up)
	if err != nil {
		return
	}
	fmt.Printf("Applied %d migrations!\n", n)

}
