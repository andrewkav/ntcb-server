//go:generate go-bindata -o bindata.go -pkg migration -prefix clickhouse  clickhouse

package migration

import (
	"embed"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/clickhouse"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"log"
)

//go:embed clickhouse/*.sql
var migrationsDir embed.FS

func Migrate(chDSN string) error {
	d, err := iofs.New(migrationsDir, "clickhouse")
	if err != nil {
		log.Fatal(err)
	}
	m, err := migrate.NewWithSourceInstance("iofs", d, chDSN)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	return nil
}
