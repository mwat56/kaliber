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
	"errors"
	"fmt"
	"path/filepath"
	"sync"
	"time"
)

type (
	// A List of database connections.
	tDBlist []*sql.DB

	// tDBpool The list of database connections.
	tDBpool struct {
		pList tDBlist     // The actual list of available connections
		pMtx  *sync.Mutex // A guard against concurrent write accesses
	}
)

var (
	// The list of database connections.
	pConnPool *tDBpool

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
			if 63 < len(pConnPool.pList) {
				pConnPool.clear()
			}
			chkTimer.Reset(chkInterval)
		}
	}
} // goMonitorPool()

// `newPool()` returns a list of database connections.
//
// To retrieve or store a certain connection use the return value's
// `get()` and `put()` methods respectively.
func newPool() *tDBpool {
	pInitPoolOnce.Do(func() {
		pConnPool = &tDBpool{
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
func (p *tDBpool) clear() *tDBpool {
	if nil == p {
		return p
	}
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
} // clear()

// `get()` selects a single database connection from the list, removes it
// from the Pool, and returns it to the caller.
//
// Callers should not assume any relation between values passed to `Put()`
// and the values returned by `get()`.
//
//	`aContext` The current request's context.
func (p *tDBpool) get(aContext context.Context) (rConn *sql.DB, rErr error) {
	if nil == p {
		rErr = errors.New(`'tDBpool' object uninitialised`)
		return
	}
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
		go goSQLtrace(`-- reused DB connection`, time.Now()) //REMOVE
	}

	return
} // get()

// `put()` adds `aConnection` to the list.
//
//	`aConnection` The database connection to add to the pool.
func (p *tDBpool) put(aConnection *sql.DB) *tDBpool {
	if nil == p {
		return p
	}
	p.pMtx.Lock()
	defer p.pMtx.Unlock()

	if nil != aConnection {
		p.pList = append(p.pList, aConnection)
	}

	go goSQLtrace(fmt.Sprintf(
		"-- recycling DB connection %d", len(p.pList)),
		time.Now()) //FIXME REMOVE

	return p
} // put()

/* _EoF_ */
