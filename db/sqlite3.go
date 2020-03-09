/*
   Copyright © 2019, 2020 M.Watermann, 10247 Berlin, Germany
                  All rights reserved
               EMail : <support@mwat.de>
*/

package db

//lint:file-ignore ST1017 - I prefer Yoda conditions

import (
	"database/sql"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3" // anonymous import
)

const (
	// Name of the `Calibre` database
	quCalibreDatabaseFilename = `metadata.db`

	// Calibre's metadata/preferences store
	quCalibrePreferencesFile = `metadata_db_prefs_backup.json`
)

type (
	tDataBase struct {
		*sql.DB           // the embedded database connection
		dbFileName string // the SQLite database file
		doCheck    chan struct{}
		wasCopied  chan struct{}
	}
)

var (
	// Pathname to the cached `Calibre` database
	dbCalibreCachePath = ``

	// Pathname to the original `Calibre` database
	dbCalibreLibraryPath = ``

	// The active `tDatabase` instance initialised by `DBopen()`.
	dbSqliteDB tDataBase

	// Optional file to log all SQL queries.
	dbSQLTraceFile = ``
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
	return filepath.Join(dbCalibreLibraryPath, quCalibrePreferencesFile)
} // CalibrePreferencesFile()

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
} // CalibreLibraryPath()

/*
// CalibreDatabaseFile returns the complete path-/filename of the `Calibre` library.
func CalibreDatabaseFile() string {
	return filepath.Join(quCalibreLibraryPath, quCalibreDatabaseFilename)
} // CalibreDatabaseFile()
*/

// SQLtraceFile returns the optional file used for logging all SQL queries.
func SQLtraceFile() string {
	return dbSQLTraceFile
} // SQLtraceFile()

// SetSQLtraceFile sets the filename to use for logging SQL queries.
//
// If the provided `aFilename` is empty the SQL logging gets disabled.
//
//	`aFilename` the tracefile to use; if empty tracing is disabled.
func SetSQLtraceFile(aFilename string) {
	if 0 < len(aFilename) {
		var doOnce sync.Once
		doOnce.Do(func() {
			if path, err := filepath.Abs(aFilename); nil == err {
				dbSQLTraceFile = path
				// start the background writer:
				go goWriteSQLtrace(dbSQLTraceFile, dbSQLTraceChannel)
			}
		})

		return
	}

	dbSQLTraceFile = ``
} // SetSQLtraceFile()

/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */

// `dbReopen()` checks whether the SQLite database file has changed since
// the last access. If so, the current database connection is closed and
// a new one is established.
func (db *tDataBase) dbReopen() error {
	select {
	case _, more := <-db.wasCopied:
		if nil != db.DB {
			_ = db.DB.Close()
		}
		var err error
		if !more {
			return err // channel closed
		}
		// `cache=shared` is essential to avoid running out of file
		// handles since each query seems to hold its own file handle.
		// `loc=auto` gets time.Time with current locale.
		// `mode=ro` is self-explanatory since we don't change the
		// DB in any way.
		dsn := `file:` + db.dbFileName + `?cache=shared&case_sensitive_like=1&immutable=0&loc=auto&mode=ro&query_only=1`
		if db.DB, err = sql.Open(`sqlite3`, dsn); nil != err {
			return err
		}
		go goSQLtrace(`-- reOpened `+dsn, time.Now())
		return db.DB.Ping()

	default:
		return nil
	}
} // dbReopen()

// Query executes a query that returns rows, typically a SELECT.
// The `args` are for any placeholder parameters in the query.
func (db *tDataBase) Query(aQuery string, args ...interface{}) (*sql.Rows, error) {
	if err := db.dbReopen(); nil != err {
		return nil, err
	}
	go goSQLtrace(aQuery, time.Now())

	rows, err := db.DB.Query(aQuery, args...)
	db.doCheck <- struct{}{}

	return rows, err
} // Query()

/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */

// `copyDatabaseFile()` copies Calibre's database file to our cache directory.
func copyDatabaseFile() (bool, error) {
	sName := filepath.Join(dbCalibreLibraryPath, quCalibreDatabaseFilename)
	dName := filepath.Join(dbCalibreCachePath, quCalibreDatabaseFilename)
	var (
		err          error
		sFile, tFile *os.File
		sFI, dFI     os.FileInfo
	)
	defer func() {
		if nil != sFile {
			_ = sFile.Close()
		}
		if nil != tFile {
			_ = tFile.Close()
		}
	}()
	if sFI, err = os.Stat(sName); nil != err {
		return false, err
	}
	if dFI, err = os.Stat(dName); nil == err {
		if sFI.ModTime().Before(dFI.ModTime()) {
			return false, nil
		}
	}

	if sFile, err = os.Open(sName); /* #nosec G304 */ err != nil {
		return false, err
	}

	tName := dName + `~`
	if tFile, err = os.OpenFile(tName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600); err != nil {
		return false, err
	}

	if _, err = io.Copy(tFile, sFile); nil != err {
		return false, err
	}
	go goSQLtrace(`-- copied `+sName+` to `+dName, time.Now())

	return true, os.Rename(tName, dName)
} // copyDatabaseFile()
