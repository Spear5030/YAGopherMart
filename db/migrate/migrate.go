package migrate

import (
	"database/sql"
	"embed"
	"github.com/pressly/goose/v3"
	"io/fs"

	_ "github.com/jackc/pgx/v5/stdlib"
)

//go:embed migrations
var Migrations embed.FS

func Migrate(dsn string, path fs.FS) error {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return err
	}
	defer db.Close()
	goose.SetBaseFS(path)
	goose.Down(db, "migrations")
	return goose.Up(db, "migrations")
}
