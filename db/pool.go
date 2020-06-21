/*
   Copyright © 2020 M.Watermann, 10247 Berlin, Germany
                  All rights reserved
               EMail : <support@mwat.de>
*/

package db

//lint:file-ignore ST1017 - I prefer Yoda conditions

/*
 * This file implements a simple connection pool to recycle
 * used SQLite connections.
 */

import (
	"context"
	"database/sql"
	"path/filepath"
	"sync"
	"time"
)

type (
	// A List of database connections.
	tDBlist []*sql.DB

	// // TOnNewFunc is called whenever an additional new database connection
	// // is required.
	// //
	// // If the operation could not be performned successfully the returned
	// // database pointer should be `nil` and the error must not be `nil`.
	// //
	// // In case of success the returned database should point to a valid
	// // database instance and the returned error must be `nil`.
	// //
	// //	`aContext` The current web request's context.
	// TOnNewFunc func(aContext context.Context) (*sql.DB, error)

	// TDBpool The list of database connections.
	TDBpool struct {
		pList tDBlist     // The actual list of available connections
		pMtx  *sync.Mutex // A guard against concurrent write accesses
		// pOnNew TOnNewFunc  // The object creating the connections
	}
)

var (
	// The list of database connections.
	//
	// NOTE: This variable as such must be considered R/O.
	pConnPool *TDBpool

	// Guard for repetitive calls to `NewPool()`.
	pInitPoolOnce sync.Once
)

// `goMonitorPool()` checks the size of the connection pool.
func goMonitorPool() {
	chkInterval := time.Minute << 2 // four minutes
	chkTimer := time.NewTimer(chkInterval)
	defer chkTimer.Stop()

	//lint:ignore S1000 - We won't use `range` here
	for {
		select {
		case <-chkTimer.C:
			pLen := pConnPool.put(nil)
			if 63 < pLen {
				pConnPool.clear()
			}
			chkTimer.Reset(chkInterval)
		}
	}
} // goMonitorPool()

// `newPool()` returns the list of database connections.
//
// To retrieve or store a certain connection use the return value's
// `Get()` and `Put()` methods respectively.
//
//	`aCreator` The object that's supposed to create database connections.
func newPool( /* aCreator TOnNewFunc */ ) *TDBpool {
	pInitPoolOnce.Do(func() {
		pConnPool = &TDBpool{
			pList: make(tDBlist, 0, 127),
			pMtx:  new(sync.Mutex),
			// pOnNew: aCreator,
		}
		go goMonitorPool()
	})

	return pConnPool
} // newPool()

/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */

// `clear()` empties the list.
//
// All connections are closed.
func (p *TDBpool) clear() *TDBpool {
	p.pMtx.Lock()
	defer p.pMtx.Unlock()

	for idx, conn := range p.pList {
		if nil != conn {
			_ = conn.Close()
		}
		p.pList[idx] = nil // clear reference
	}
	p.pList = p.pList[:0] // empty the list

	return p
} // Clear()

// `get()` selects a single database connection from the list, removes it
// from the Pool, and returns it to the caller.
//
// Callers should not assume any relation between values passed to `Put()`
// and the values returned by `get()`.
//
//	`aContext` The current request's context.
func (p *TDBpool) get(aContext context.Context) (rConn *sql.DB, rErr error) {
	p.pMtx.Lock()
	defer p.pMtx.Unlock()

	// There are three cases to consider:
	//
	// (1) the list is empty,
	// (2) the list has one entry,
	// (3) the list has more than one entry.
	//
	// We unroll these cases here to handle each most efficiently.

	sLen := len(p.pList)
	if 0 == sLen { // case (1)

		//XXX Are there custom functions to inject?

		// `cache=shared` is essential to avoid running out of file
		// handles since each query seems to hold its own file handle.
		// `loc=auto` gets time.Time with current locale.
		// `mode=ro` is self-explanatory since we don't change the DB
		// in any way.
		dsn := `file:` +
			filepath.Join(dbCalibreCachePath, dbCalibreDatabaseFilename) +
			`?cache=shared&case_sensitive_like=1&immutable=0&loc=auto&mode=ro&query_only=1`
		select {
		case <-aContext.Done():
			rErr = aContext.Err()

		default:
			if rConn, rErr = sql.Open(`sqlite3`, dsn); nil == rErr {
				// rConn.Exec("PRAGMA xxx=yyy")
				go goSQLtrace(`-- opened DB`, time.Now()) //REMOVE
				rErr = rConn.PingContext(aContext)
			}
		}

		return
	}

	select {
	case <-aContext.Done():
		rErr = aContext.Err()

	default:
		rConn = p.pList[0]
		p.pList[0] = nil // remove reference

		if 1 == sLen { // case (2)
			p.pList = p.pList[:0] // empty the list
		} else { // case (3)
			p.pList = p.pList[1:] // remove first item from list
		}
	}

	return
} // Get()

// `put()` adds `aConnection` to the list returning the new number
// of elements in the Pool.
//
// To just get the current number of connections in the pool
// use `nil` as the method's argument.
//
//	`aConnection` The database connection to add to the pool.
func (p *TDBpool) put(aConnection *sql.DB) int {
	p.pMtx.Lock()
	defer p.pMtx.Unlock()

	if nil != aConnection {
		p.pList = append(p.pList, aConnection)
	}

	return len(p.pList)
} // put()

/* _EoF_ */
