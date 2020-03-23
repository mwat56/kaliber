/*
   Copyright © 2019, 2020 M.Watermann, 10247 Berlin, Germany
                  All rights reserved
               EMail : <support@mwat.de>
*/

package db

//lint:file-ignore ST1017 - I prefer Yoda conditions

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

/*
 * This file provides functions to sync the used database copy
 * with the original Calibre library.
 *
 * Additionally there are functions to handle an external text file
 * for tracing all used SQL queries.
 */

var (

	// `syncCheck` Signal channel to check for database changes.
	syncCheck = make(chan struct{}, 64)

	// `syncCopied` Signal channel for a new database copy.
	syncCopied = make(chan struct{}, 1)
)

// `copyDatabaseFile()` copies Calibre's original database file
// to our cache directory.
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
// file has changed.
// If so, that file is copied to the cache directory from where it is
// read and used by the `dbSQLiteDB` instance.
//
//	`aCheck` R/O channel to check for changes.
//	`aCopied` W/O channel to signal a new database copy.
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

/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */

var (
	// The channel to send SQL to and read trace messages from.
	dbSQLTraceChannel = make(chan string, 64)

	// Optional file to log all SQL queries.
	dbSQLTraceFile = ``
)

// `goSQLtrace()` runs in background to log `aQuery`
// (if a tracefile is set).
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
//	`aSource` R/O channel to read the log messages to write.
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

	// Interval to look for new messages to arrive.
	sleepDuration := time.Second * 3

	// Let the application initialise:
	time.Sleep(sleepDuration)

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
				time.Sleep(sleepDuration)
			} else {
				if os.Stderr != file {
					_ = file.Close()
				}
				file = nil
			}
		}
	}
} // goWriteSQLtrace()

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

/* _EoF_ */
