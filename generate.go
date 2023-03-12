package main

/*
go install github.com/volatiletech/sqlboiler/v4@latest
go install github.com/volatiletech/sqlboiler/v4/drivers/sqlboiler-psql@latest
go install github.com/volatiletech/sqlboiler/v4/drivers/sqlboiler-mysql@latest
*/

//go:generate sqlboiler --config ".sqlboiler-psql.toml" --wipe --add-soft-deletes psql
//go:generate sqlboiler --config ".sqlboiler-mysql.toml" --wipe --add-soft-deletes mysql
