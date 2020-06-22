/*
   Copyright © 2020 M.Watermann, 10247 Berlin, Germany
                  All rights reserved
               EMail : <support@mwat.de>
*/

package db

import (
	"context"
	"crypto/md5"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"testing"
)

func prepDBforTesting(aContext context.Context) {
	libPath := `/var/opt/Calibre`
	s := fmt.Sprintf("%x", md5.Sum([]byte(libPath))) // #nosec G401
	ucd, _ := os.UserCacheDir()
	SetCalibreCachePath(filepath.Join(ucd, "kaliber", s))
	SetCalibreLibraryPath(libPath)
	SetSQLtraceFile("./SQLtrace.sql")
	_, _ = OpenDatabase(aContext)
} // prepDBforTesting()

func Test_tDBpool_clear(t *testing.T) {
	ctx := context.Background()
	prepDBforTesting(ctx)

	type fields struct {
		pList tDBlist
		pMtx  *sync.Mutex
	}
	tests := []struct {
		name   string
		fields *tDBpool
		want   *tDBpool
	}{
		// TODO: Add test cases.
		{" 0", nil, nil},
		{" 1", pConnPool, pConnPool},
		{" 2", pConnPool, pConnPool},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := tt.fields
			if got := p.clear(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("tDBpool.clear() = %v, want %v", got, tt.want)
			}
		})
	}
} // Test_tDBpool_clear()

func Test_tDBpool_get(t *testing.T) {
	ctx := context.Background()
	prepDBforTesting(ctx)

	type args struct {
		aContext context.Context
	}
	tests := []struct {
		name     string
		fields   *tDBpool
		args     args
		wantRnil bool
		wantErr  bool
	}{
		// TODO: Add test cases.
		{" 0", nil, args{ctx}, true, true},
		{" 1", pConnPool, args{ctx}, false, false},
		{" 2", pConnPool, args{ctx}, false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := tt.fields
			gotRConn, err := p.get(tt.args.aContext)
			if (err != nil) != tt.wantErr {
				t.Errorf("tDBpool.get() error = %v, wantErr %v",
					err, tt.wantErr)
				return
			}
			if nil != gotRConn {
				if tt.wantRnil {
					t.Errorf("tDBpool.get() = %v, want NIL", gotRConn)
				}
			} else {
				if !tt.wantRnil {
					t.Errorf("tDBpool.get() = %v, want !NIL", nil)
				}
			}
		})
	}
} // Test_tDBpool_get()

func Test_tDBpool_put(t *testing.T) {
	ctx := context.Background()
	prepDBforTesting(ctx)

	var conn1 *sql.DB
	conn2, _ := pConnPool.get(ctx)

	type args struct {
		aConnection *sql.DB
	}
	tests := []struct {
		name   string
		fields *tDBpool
		args   args
		want   *tDBpool
	}{
		// TODO: Add test cases.
		{" 0", nil, args{conn1}, nil},
		{" 1", pConnPool, args{conn1}, pConnPool},
		{" 2", pConnPool, args{conn2}, pConnPool},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := tt.fields
			if got := p.put(tt.args.aConnection); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("tDBpool.put() = %v, want %v", got, tt.want)
			}
		})
	}
}
