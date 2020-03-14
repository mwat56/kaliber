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
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
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
		*sql.DB          // the embedded database connection
		fName     string // the SQLite database file
		doCheck   chan struct{}
		wasCopied chan struct{}
	}
)

var (
	// Pathname to the cached `Calibre` database
	dbCalibreCachePath = ``

	// Pathname to the original `Calibre` database
	dbCalibreLibraryPath = ``

	// The active `tDatabase` instance initialised by `OpenDatabase()`.
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
	return filepath.Join(dbCalibreLibraryPath, dbCalibrePreferencesFile)
} // CalibrePreferencesFile()

// `copyDatabaseFile()` copies Calibre's database file to our cache directory.
func copyDatabaseFile() (bool, error) {
	sName := filepath.Join(dbCalibreLibraryPath, dbCalibreDatabaseFilename)
	dName := filepath.Join(dbCalibreCachePath, dbCalibreDatabaseFilename)
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

// `goCheckFile()` checks in background whether the original database
// file has changed. If so, that file is copied to the cache directory
// from where it is read and used by the `quSQLiteDB` instance.
func goCheckFile(aCheck <-chan struct{}, aCopied chan<- struct{}) {
	timer := time.NewTimer(time.Minute)
	defer func() { _ = timer.Stop() }()

	for {
		select {
		case _, more := <-aCheck:
			if !more {
				return // channel closed
			}
			if copied, err := copyDatabaseFile(); copied && (nil == err) {
				aCopied <- struct{}{}
			}
			_ = timer.Reset(time.Minute)

		case <-timer.C:
			if copied, err := copyDatabaseFile(); copied && (nil == err) {
				aCopied <- struct{}{}
			}
			_ = timer.Reset(time.Minute)
		}
	}
} // goCheckFile()

var (
	// The channel to send SQL to and read trace messages from
	dbSQLTraceChannel = make(chan string, 64)
)

// `goSQLtrace()` runs in background to log `aQuery` (if a tracefile is set).
//
//	`aQuery` The SQL query to log.
//	`aTime` The time at which the query was run.
func goSQLtrace(aQuery string, aTime time.Time) {
	if 0 == len(dbSQLTraceFile) {
		return
	}
	aQuery = strings.ReplaceAll(aQuery, "\t", ` `)
	aQuery = strings.ReplaceAll(aQuery, "\n", ` `)

	dbSQLTraceChannel <- aTime.Format(`2006-01-02 15:04:05 `) +
		strings.ReplaceAll(aQuery, `  `, ` `)
} // goSQLtrace()

// `goWriteSQLtrace()` performs the actual file writes.
//
// This function is called only once, handling all write requests
// in background.
//
//	`aTraceLog` The name of the logfile to write to.
//	`aSource` The source of the log messages to write.
func goWriteSQLtrace(aTraceLog string, aSource <-chan string) {
	var (
		err  error
		file *os.File
		more bool
		txt  string
	)
	defer func() {
		if (nil != file) && (os.Stderr != file) {
			_ = file.Close()
		}
	}()

	// let the application initialise:
	time.Sleep(time.Second)

	for { // wait for strings to write
		select {
		case txt, more = <-aSource:
			if !more { // channel closed
				log.Println(`queries.goWriteSQLtrace(): message channel closed`)
				return
			}
			if nil == file {
				if file, err = os.OpenFile(aTraceLog,
					os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0640); /* #nosec G302 */ nil != err {
					file = os.Stderr // a last resort
				}
			}
			fmt.Fprintln(file, txt)

		default:
			if nil == file {
				time.Sleep(time.Second)
			} else {
				if os.Stderr != file {
					_ = file.Close()
				}
				file = nil
			}
		}
	}
} // goWriteSQLtrace()

// OpenDatabase establishes a new database connection.
func OpenDatabase(aContext context.Context) error {
	var once sync.Once
	once.Do(func() {
		dbSqliteDB.fName = filepath.Join(dbCalibreCachePath, dbCalibreDatabaseFilename)
		dbSqliteDB.doCheck = make(chan struct{}, 64)
		dbSqliteDB.wasCopied = make(chan struct{}, 1)
	})

	// prepare the local database copy:
	if _, err := copyDatabaseFile(); nil != err {
		return err
	}
	// signal for `dbSqliteDB.reOpen()`:
	dbSqliteDB.wasCopied <- struct{}{}

	// start monitoring the original database file:
	go goCheckFile(dbSqliteDB.doCheck, dbSqliteDB.wasCopied)

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

// SetSQLtraceFile sets the filename to use for logging SQL queries.
//
// If the provided `aFilename` is empty the SQL logging gets disabled.
//
//	`aFilename` the tracefile to use; if empty tracing is disabled.
func SetSQLtraceFile(aFilename string) {
	if 0 < len(aFilename) {
		var once sync.Once
		once.Do(func() {
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

// SQLtraceFile returns the optional file used for logging all SQL queries.
func SQLtraceFile() string {
	return dbSQLTraceFile
} // SQLtraceFile()

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
	db.doCheck <- struct{}{}

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
	case _, more := <-db.wasCopied:
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
		// `mode=ro` is self-explanatory since we don't change the
		// DB in any way.
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
