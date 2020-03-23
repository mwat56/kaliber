/*
   Copyright © 2019, 2020 M.Watermann, 10247 Berlin, Germany
                  All rights reserved
               EMail : <support@mwat.de>
*/

package db

//lint:file-ignore ST1017 - I prefer Yoda conditions

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3" // anonymous import
)

const (
	// Name of the `Calibre` database
	dbCalibreDatabaseFilename = `metadata.db`

	// Calibre's metadata/preferences store
	dbCalibrePreferencesFile = `metadata_db_prefs_backup.json`
)

type (
	tDataBase struct {
		*sql.DB        // the embedded database connection
		fName   string // the SQLite database file
	}
)

var (
	// Pathname to the cached `Calibre` database
	dbCalibreCachePath = ``

	// Pathname to the original `Calibre` database
	dbCalibreLibraryPath = ``

	// The active `tDatabase` instance initialised by `OpenDatabase()`.
	dbSqliteDB tDataBase
)

// CalibreCachePath returns the directory of the copied `Calibre` database.
func CalibreCachePath() string {
	return dbCalibreCachePath
} // CalibreCachePath()

// CalibreLibraryPath returns the base directory of the `Calibre` library.
func CalibreLibraryPath() string {
	return dbCalibreLibraryPath
} // CalibreLibraryPath()

// CalibrePreferencesFile returns the complete path-/filename of the
// `Calibre` library's preferences file.
func CalibrePreferencesFile() string {
	return filepath.Join(dbCalibreLibraryPath, dbCalibrePreferencesFile)
} // CalibrePreferencesFile()

// OpenDatabase establishes a new database connection.
func OpenDatabase(aContext context.Context) error {
	var once sync.Once
	once.Do(func() {
		dbSqliteDB.fName = filepath.Join(dbCalibreCachePath, dbCalibreDatabaseFilename)
	})

	// prepare the local database copy:
	if _, err := copyDatabaseFile(); nil != err {
		return err
	}
	// signal for `dbSqliteDB.reOpen()`:
	syncCopied <- struct{}{}

	// start monitoring the original database file:
	go goCheckFile(syncCheck, syncCopied)

	return dbSqliteDB.reOpen(aContext)
} // OpenDatabase()

// SetCalibreCachePath sets the directory of the `Calibre` database copy.
//
//	`aPath` is the directory path to use for caching the Calibre library.
func SetCalibreCachePath(aPath string) {
	if path, err := filepath.Abs(aPath); nil == err {
		aPath = path
	}
	if fi, err := os.Stat(aPath); (nil == err) && fi.IsDir() {
		dbCalibreCachePath = aPath
	} else if err := os.MkdirAll(aPath, os.ModeDir|0775); nil == err {
		dbCalibreCachePath = aPath
	}
} // SetCalibreCachePath()

// SetCalibreLibraryPath sets the base directory of the `Calibre` library.
//
//	`aPath` is the directory path where the Calibre library resides.
func SetCalibreLibraryPath(aPath string) string {
	if path, err := filepath.Abs(aPath); nil == err {
		aPath = path
	}
	if fi, err := os.Stat(aPath); (nil == err) && fi.IsDir() {
		dbCalibreLibraryPath = aPath
	} else {
		dbCalibreLibraryPath = ``
	}

	return dbCalibreLibraryPath
} // SetCalibreLibraryPath()

/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */

// `query()` executes a query that returns rows, typically a SELECT.
// The `args` are for any placeholder parameters in the query.
//
//	`aContext` The current request's context.
//	`aQuery` The SQL query to run.
func (db *tDataBase) query(aContext context.Context, aQuery string) (rRows *sql.Rows, rErr error) {
	if rErr = db.reOpen(aContext); nil != rErr {
		return
	}
	// syncCheck <- struct{}{}

	go goSQLtrace(aQuery, time.Now())

	rRows, rErr = db.DB.QueryContext(aContext, aQuery)

	return
} // query()

// `reOpen()` checks whether the SQLite database file has changed
// since the last access.
// If so, the current database connection is closed and a new one
// is established.
//
//	`aContext` The current request's context.
func (db *tDataBase) reOpen(aContext context.Context) error {
	select {
	case _, more := <-syncCopied:
		if nil != db.DB {
			_ = db.DB.Close()
			db.DB = nil
		}
		var err error
		if !more {
			return err // channel closed
		}

		//XXX Are there custom functions to inject?

		// `cache=shared` is essential to avoid running out of file
		// handles since each query seems to hold its own file handle.
		// `loc=auto` gets time.Time with current locale.
		// `mode=ro` is self-explanatory since we don't change the DB
		// in any way.
		dsn := `file:` + db.fName +
			`?cache=shared&case_sensitive_like=1&immutable=0&loc=auto&mode=ro&query_only=1`
		if db.DB, err = sql.Open(`sqlite3`, dsn); nil != err {
			return err
		}
		// db.Exec("PRAGMA xxx=yyy")

		go goSQLtrace(`-- reOpened `+dsn, time.Now())
		return db.DB.PingContext(aContext)

	default:
		return nil
	}
} // reOpen()

/* _EoF_ */
