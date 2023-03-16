package msql

import "github.com/Defacto2/df2/pkg/configger"

// sql.go contains custom MySQL queries and statements

type Version string // Version of the MySQL in use.

func (v *Version) Query(cfg configger.Config) error {
	conn, err := Connect(cfg)
	if err != nil {
		return err
	}
	rows, err := conn.Query("SELECT version();")
	if err != nil {
		return err
	}
	defer rows.Close()
	defer conn.Close()
	for rows.Next() {
		if err := rows.Scan(v); err != nil {
			return err
		}
	}
	return nil
}
