// Package proof handles files that have the section tagged as releaseproof.
package proof

import (
	"database/sql"
	"fmt"

	"github.com/Defacto2/df2/pkg/database"
	"github.com/Defacto2/df2/pkg/logs"
	"github.com/Defacto2/df2/pkg/proof/internal/record"
	"github.com/Defacto2/df2/pkg/proof/internal/stat"
)

// Request records classified with the "proof" filetype.
type Request struct {
	Overwrite   bool   // Overwrite any existing proof assets such as images.
	AllProofs   bool   // AllProofs parses all proofs, not just new uploads.
	HideMissing bool   // HideMissing ignore proofs that are missing UUID download files.
	byID        string // Id used for proofs, either a uuid or id string.
}

// Query parses a single proof with the record id or uuid.
func (request *Request) Query(id string) error {
	if err := database.CheckID(id); err != nil {
		return fmt.Errorf("request query id %q: %w", id, err)
	}
	request.byID = id
	if err := request.Queries(); err != nil {
		return fmt.Errorf("request queries: %w", err)
	}
	return nil
}

// Queries parses all proofs.
func (request Request) Queries() error { //nolint:funlen
	s := stat.Init()
	db := database.Connect()
	defer db.Close()
	rows, err := db.Query(sqlSelect(request.byID))
	if err != nil {
		return err
	}
	if rows.Err() != nil {
		return rows.Err()
	}
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		return err
	}
	values := make([]sql.RawBytes, len(columns))
	// more information: https://github.com/go-sql-driver/mysql/wiki/Examples
	scanArgs := make([]any, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}
	for rows.Next() {
		s.Total++
	}
	fmt.Print(Total(&s, request))
	rows, err = db.Query(sqlSelect(request.byID))
	if err != nil {
		return err
	}
	if rows.Err() != nil {
		return rows.Err()
	}
	defer rows.Close()
	for rows.Next() {
		if err := rows.Scan(scanArgs...); err != nil {
			return err
		}
		if request.Skip(values) {
			continue
		}
		s.Count++
		r := record.New(values, s.BasePath)
		if skip, err := record.Skip(s, r, request.HideMissing); skip {
			continue
		} else if err != nil {
			return err
		}
		s.Columns = columns
		s.Overwrite = request.Overwrite
		s.Values = &values
		if err := r.Iterate(s); err != nil {
			return err
		}
	}
	fmt.Print(s.Summary(request.byID))
	return nil
}

// Total returns the sum of the records.
func Total(s *stat.Proof, request Request) string {
	if s == nil {
		return ""
	}
	if s.Total < 1 && request.byID != "" {
		return fmt.Sprintf("file record id '%s' does not exist or is not a release proof\n", request.byID)
	}
	if s.Total > 1 {
		return fmt.Sprintln("Total records", s.Total)
	}
	return ""
}

// Skip uses argument flags to check if a record is to be ignored.
func (request Request) Skip(values []sql.RawBytes) bool {
	if request.byID != "" && request.Overwrite {
		return false
	}
	if n := database.IsProof(values); !n && !request.AllProofs {
		if request.byID != "" {
			logs.Printf("skip record id '%s', as it is not new\n", request.byID)
		}
		return true
	}
	return false
}

func sqlSelect(id string) string {
	s := "SELECT `id`,`uuid`,`deletedat`,`createdat`,`filename`,`file_zip_content`,`updatedat`,`platform`"
	w := " WHERE `section` = 'releaseproof'"
	if id != "" {
		switch {
		case database.IsUUID(id):
			w = fmt.Sprintf("%v AND `uuid`=%q", w, id)
		case database.IsID(id):
			w = fmt.Sprintf("%v AND `id`=%q", w, id)
		}
	}
	return s + " FROM `files`" + w
}
