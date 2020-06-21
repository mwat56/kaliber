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
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
	// sqlite "github.com/mattn/go-sqlite3"
)

/*
 * This file provides functions to sync the used database copy with the
 * original Calibre library.
 *
 * To avoid any LOCKing problems (which happened quite frequently)
 * when reading the Calibre database while it is edited by the original
 * Calibre installation here we simply copy Calibre's database file
 * into the user's cache directory.
 * This way we can use R/O access without the fear that the database might
 * be changed under our feet by other processes.
 *
 * Additionally there are functions to handle an external text file
 * for tracing all used SQL queries.
 */

var (
	// `syncCopiedChan` Signal channel for a new database copy.
	// `goSyncFile()` writes to the channel whenever the database
	// file was copied so others (`TDataBase.reOpen()`) can check.
	syncCopiedChan = make(chan struct{}, 2)

	// Guard against parallel database copies.
	syncCopyMtx = new(sync.Mutex)

	// The channel to send SQL to and read trace messages from.
	syncSQLTraceChannel = make(chan string, 127)

	// Optional file to log all SQL queries.
	syncSQLTraceFile = ``
)

// `goSyncFile()` checks in background once a minute whether the
// original database file has changed.
// If so, that file is copied to the cache directory from where it is
// read and used by the `db.TDatabase` instance.
func goSyncFile() {
	timer := time.NewTimer(time.Minute)
	defer func() {
		_ = timer.Stop()
	}()

	//lint:ignore S1000 - We won't use `range` here
	for {
		select {
		case <-timer.C:
			if copied, err := syncDatabaseFile(); copied && (nil == err) {
				syncCopiedChan <- struct{}{}
			}
			_ = timer.Reset(time.Minute)
		}
	}
} // goSyncFile()

/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */

// `goSQLtrace()` runs in background to log `aQuery`
// (if a tracefile is set).
//
//	`aQuery` The SQL query to log.
//	`aWhen` The time when the query started.
func goSQLtrace(aQuery string, aWhen time.Time) {
	if 0 == len(syncSQLTraceFile) {
		return
	}
	aQuery = strings.Replace(aQuery, "\t", ` `, -1)
	aQuery = strings.Replace(aQuery, "\n", ` `, -1)

	syncSQLTraceChannel <- aWhen.Format(`2006-01-02 15:04:05.000 `) +
		strings.Replace(aQuery, `  `, ` `, -1)
} // goSQLtrace()

const (
	// Timer interval to wait for trace file closing after inactivity.
	syncSex = time.Second << 3 // eight seconds

	// Mode of opening the logfile(s).
	syncOpenFlags = os.O_CREATE | os.O_APPEND | os.O_WRONLY | os.O_SYNC
)

// `goWriteSQLtrace()` performs the actual file writes.
//
// This function is called only once, handling all write requests
// while running in background.
func goWriteSQLtrace() {
	var (
		err        error
		traceFile  *os.File
		fileCloser *time.Timer
	)
	defer func() {
		if (nil != traceFile) && (os.Stderr != traceFile) {
			_ = traceFile.Close()
		}
		if nil != fileCloser {
			_ = fileCloser.Stop()
		}
	}()

	// Let the application initialise:
	time.Sleep(time.Second)
	fileCloser = time.NewTimer(syncSex)

	for { // wait for strings to write
		select {
		case txt, more := <-syncSQLTraceChannel:
			if !more { // channel closed
				return
			}
			if (0 < len(syncSQLTraceFile)) && (0 < len(txt)) {
				if nil == traceFile {
					if traceFile, err = os.OpenFile(syncSQLTraceFile, syncOpenFlags, 0640); /* #nosec G302 */ nil != err {
						// A last resort:
						traceFile = os.Stderr
					}
				}
				fmt.Fprintln(traceFile, txt)
			}
			fileCloser.Reset(syncSex)

		case <-fileCloser.C:
			// Make sure to close the trace file after
			// a certain time of inactivity.
			if nil != traceFile {
				if os.Stderr != traceFile {
					_ = traceFile.Close()
				}
				traceFile = nil
			}
			fileCloser.Reset(syncSex)
		}
	}
} // goWriteSQLtrace()

var (
	// Make sure the filename is set only once.
	syncFilenameOnce sync.Once
)

// SetSQLtraceFile sets the filename to use for logging SQL queries.
//
// If the provided `aFilename` is empty the SQL logging gets disabled.
//
// NOTE: Once the function was called with a valid `aFilename` argument
// any forther call will be ignored.
//
//	`aFilename` The tracefile to use, if empty tracing is disabled.
func SetSQLtraceFile(aFilename string) {
	if 0 < len(aFilename) {
		if path, err := filepath.Abs(aFilename); nil == err {
			syncSQLTraceFile = path
		}
		syncFilenameOnce.Do(func() {
			// start the background writer:
			go goWriteSQLtrace()
		})

		return
	}

	syncSQLTraceFile = ``
} // SetSQLtraceFile()

// SQLtraceFile returns the filename used for the optional logging
// of SQL queries.
func SQLtraceFile() string {
	return syncSQLTraceFile
} // SQLtraceFile()

// `syncDatabaseFile()` copies Calibre's original database file
// to the configured cache directory.
//
// The `bool` return value signals whether the database file was
// actually copied or not.
// The `error` return value is either `nil` in case of success or the
// error that occurred.
func syncDatabaseFile() (bool, error) {
	var (
		err              error
		srcFile, tmpFile *os.File
		srcFI, dstFI     os.FileInfo
	)
	defer func() {
		if nil != srcFile {
			_ = srcFile.Close()
		}
		if nil != tmpFile {
			_ = tmpFile.Close()
		}
	}()
	syncCopyMtx.Lock()
	defer syncCopyMtx.Unlock()

	srcName := filepath.Join(dbCalibreLibraryPath, dbCalibreDatabaseFilename)
	if srcFI, err = os.Stat(srcName); nil != err {
		return false, err
	}

	dstName := filepath.Join(dbCalibreCachePath, dbCalibreDatabaseFilename)
	if dstFI, err = os.Stat(dstName); nil == err {
		if srcFI.ModTime().Before(dstFI.ModTime()) {
			return false, nil
		}
	}

	if srcFile, err = os.OpenFile(srcName, os.O_RDONLY, 0); /* #nosec G304 */ err != nil {
		return false, err
	}

	tmpName := dstName + `~`
	if tmpFile, err = os.OpenFile(tmpName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600); err != nil {
		return false, err
	}

	if _, err = io.Copy(tmpFile, srcFile); nil != err {
		return false, err
	}
	go goSQLtrace(`-- copied `+srcName+` to `+dstName, time.Now())

	return true, os.Rename(tmpName, dstName)
} // syncDatabaseFile()

/*
func syncBackupDataBase() (bool, error) {
	syncCopyMtx.Lock()
	defer syncCopyMtx.Unlock()

	var (
		backup           sqlite.SQLiteBackup
		dstConn, srcConn *sql.DB
		destConn         *sqlite.SQLiteConn
		done             bool
		err              error
		sFI, dFI         os.FileInfo
	)
	sName := filepath.Join(dbCalibreLibraryPath, dbCalibreDatabaseFilename)
	if sFI, err = os.Stat(sName); nil != err {
		return false, err
	}

	dName := filepath.Join(dbCalibreCachePath, dbCalibreDatabaseFilename)
	if dFI, err = os.Stat(dName); nil == err {
		if sFI.ModTime().Before(dFI.ModTime()) {
			return false, nil
		}
	} // ELSE: the dest file doesn't exist yet

	if srcConn, err = sql.Open(`sqlite3`, `file:`+sName+`?cache=shared&mode=ro`); nil != err {
		return false, err
	}
	defer srcConn.Close()

	tName := dName + `~`
	if dstConn, err = sql.Open(`sqlite3`, `file:`+tName); nil != err {
		return false, err
	}
	defer dstConn.Close()

	if backup, err = dstConn.Backup(`main`, srcConn, `main`); nil != err {
		return false, err
	}
	defer backup.Close()

	if done, err = backup.Step(-1); nil != err {
		return false, err
	}

	if !done {
		return false, errors.New(`Problem backing up ` + sName + ` to ` + tName)
	}

	go goSQLtrace(`-- copied ` + sName + ` to ` + dName)

	return true, os.Rename(tName, dName)
} // syncBackupDataBase()
*/

/* _EoF_ */
