//go:generate go-bindata -o bindata.go -pkg migration -prefix clickhouse  clickhouse

package migration

import (
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/clickhouse"
	bindata "github.com/golang-migrate/migrate/v4/source/go_bindata"
)

func Migrate(chDSN string) error {
	s := bindata.Resource(AssetNames(),
		func(name string) ([]byte, error) {
			return Asset(name)
		})
	d, err := bindata.WithInstance(s)
	if err != nil {
		return err
	}

	m, err := migrate.NewWithSourceInstance("go-bindata", d, chDSN)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	return nil
}
