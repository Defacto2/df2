package msql

// sql.go contains custom MySQL queries and statements

type Version string // Version of the MySQL in use.

func (v *Version) Query() error {
	conn, err := ConnectDB()
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
