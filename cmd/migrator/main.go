package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	var DSN, migrationsPath, typeTask string

	flag.StringVar(&DSN, "dsn", "", "Database DSN")
	flag.StringVar(&migrationsPath, "migrationsPath", "file://migrations", "Path to migrations")
	//flag.StringVar(&migrationTable, "migrationTable", "migrations", "Migration table name")
	flag.StringVar(&typeTask, "typeTask", "up", "what you want - up or down")
	flag.Parse()

	dir, _ := os.Getwd()
	fmt.Println("Mirgrator start - current directory:", dir)

	m, err := migrate.New(migrationsPath, DSN)
	if err != nil {
		panic(err)
	}

	if typeTask == "up" {
		err = m.Up()
	} else {
		err = m.Down()
	}
	if err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			println("no changes")
			return
		}
		panic(err)
	}
	fmt.Println("migrations applied")
}
