// Code generated by SQLBoiler 4.16.1 (https://github.com/volatiletech/sqlboiler). DO NOT EDIT.
// This file is meant to be re-generated in place and/or deleted at any time.

package models

import "testing"

// This test suite runs each operation test in parallel.
// Example, if your database has 3 tables, the suite will run:
// table1, table2 and table3 Delete in parallel
// table1, table2 and table3 Insert in parallel, and so forth.
// It does NOT run each operation group in parallel.
// Separating the tests thusly grants avoidance of Postgres deadlocks.
func TestParent(t *testing.T) {
	t.Run("Files", testFiles)
}

func TestSoftDelete(t *testing.T) {
	t.Run("Files", testFilesSoftDelete)
}

func TestQuerySoftDeleteAll(t *testing.T) {
	t.Run("Files", testFilesQuerySoftDeleteAll)
}

func TestSliceSoftDeleteAll(t *testing.T) {
	t.Run("Files", testFilesSliceSoftDeleteAll)
}

func TestDelete(t *testing.T) {
	t.Run("Files", testFilesDelete)
}

func TestQueryDeleteAll(t *testing.T) {
	t.Run("Files", testFilesQueryDeleteAll)
}

func TestSliceDeleteAll(t *testing.T) {
	t.Run("Files", testFilesSliceDeleteAll)
}

func TestExists(t *testing.T) {
	t.Run("Files", testFilesExists)
}

func TestFind(t *testing.T) {
	t.Run("Files", testFilesFind)
}

func TestBind(t *testing.T) {
	t.Run("Files", testFilesBind)
}

func TestOne(t *testing.T) {
	t.Run("Files", testFilesOne)
}

func TestAll(t *testing.T) {
	t.Run("Files", testFilesAll)
}

func TestCount(t *testing.T) {
	t.Run("Files", testFilesCount)
}

func TestInsert(t *testing.T) {
	t.Run("Files", testFilesInsert)
	t.Run("Files", testFilesInsertWhitelist)
}

func TestReload(t *testing.T) {
	t.Run("Files", testFilesReload)
}

func TestReloadAll(t *testing.T) {
	t.Run("Files", testFilesReloadAll)
}

func TestSelect(t *testing.T) {
	t.Run("Files", testFilesSelect)
}

func TestUpdate(t *testing.T) {
	t.Run("Files", testFilesUpdate)
}

func TestSliceUpdateAll(t *testing.T) {
	t.Run("Files", testFilesSliceUpdateAll)
}
