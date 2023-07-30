// Package proof handles files that have the section tagged as releaseproof.
package proof

import (
	"database/sql"
	"errors"
	"fmt"
	"io"

	"github.com/Defacto2/df2/pkg/conf"
	"github.com/Defacto2/df2/pkg/database"
	"github.com/Defacto2/df2/pkg/proof/internal/record"
)

var ErrPointer = errors.New("pointer value cannot be nil")

// Request records classified with the "proof" filetype.
type Request struct {
	Overwrite   bool   // Overwrite any existing proof assets such as images.
	All         bool   // All parses every proofs, not just the new uploads.
	HideMissing bool   // HideMissing ignore proofs that are missing UUID download files.
	ByID        string // ByID is the ID used by a proof, either a uuid or id string.
}

// Query parses a single proof with the record id or uuid.
func (request *Request) Query(db *sql.DB, w io.Writer, cfg conf.Config, id string) error {
	if db == nil {
		return database.ErrDB
	}
	if w == nil {
		w = io.Discard
	}
	if err := database.CheckID(id); err != nil {
		return fmt.Errorf("request query id %q: %w", id, err)
	}
	request.ByID = id
	if err := request.Queries(db, w, cfg); err != nil {
		return fmt.Errorf("request queries: %w", err)
	}
	return nil
}

// Queries parses all proofs.
func (request Request) Queries(db *sql.DB, w io.Writer, cfg conf.Config) error { //nolint:cyclop,funlen
	if db == nil {
		return database.ErrDB
	}
	if w == nil {
		w = io.Discard
	}
	s, err := record.Init(cfg)
	if err != nil {
		return err
	}
	rows, err := db.Query(Select(request.ByID))
	if err != nil {
		return err
	}
	if rows.Err() != nil {
		return rows.Err()
	}
	defer rows.Close()
	cols, err := rows.Columns()
	if err != nil {
		return err
	}
	values := make([]sql.RawBytes, len(cols))
	// more information: https://github.com/go-sql-driver/mysql/wiki/Examples
	args := make([]any, len(values))
	for i := range values {
		args[i] = &values[i]
	}
	for rows.Next() {
		s.Total++
	}
	sum, err := Total(&s, request)
	if err != nil {
		return err
	}
	fmt.Fprint(w, sum)
	rows, err = db.Query(Select(request.ByID))
	if err != nil {
		return err
	}
	if rows.Err() != nil {
		return rows.Err()
	}
	defer rows.Close()
	for rows.Next() {
		if err := rows.Scan(args...); err != nil {
			return err
		}
		b, err := request.Skip(w, values)
		if err != nil {
			return err
		}
		if b {
			continue
		}
		s.Count++
		r := record.New(values, s.BasePath)
		nw := w
		if request.HideMissing {
			nw = io.Discard
		}
		if skip, err := record.Skip(nw, s, r); skip {
			continue
		} else if err != nil {
			return err
		}
		s.Columns = cols
		s.Overwrite = request.Overwrite
		s.Values = &values
		if err := r.Iterate(db, w, cfg, s); err != nil {
			return err
		}
	}
	fmt.Fprint(w, s.Summary(request.ByID))
	return nil
}

// Skip uses argument flags to check if a record is to be ignored.
func (request Request) Skip(w io.Writer, values []sql.RawBytes) (bool, error) {
	if w == nil {
		w = io.Discard
	}
	if request.ByID != "" && request.Overwrite {
		return false, nil
	}
	n, err := database.IsUnApproved(values)
	if err != nil {
		return false, err
	}
	if !n && !request.All {
		if request.ByID != "" {
			fmt.Fprintf(w, "skip record id '%s', as it is not new\n", request.ByID)
		}
		return true, nil
	}
	return false, nil
}

func Select(id string) string {
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

// Total returns the sum of the records.
func Total(s *record.Proof, request Request) (string, error) {
	if s == nil {
		return "", fmt.Errorf("stat proof, %w", ErrPointer)
	}
	if s.Total < 1 && request.ByID != "" {
		return fmt.Sprintf("file record id '%s' does not exist or is not a release proof\n", request.ByID), nil
	}
	if s.Total > 1 {
		return fmt.Sprintf("\t  TOTAL, %d proof records\n", s.Total), nil
	}
	return "", nil
}
