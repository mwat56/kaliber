/*
   Copyright © 2020 M.Watermann, 10247 Berlin, Germany
                  All rights reserved
               EMail : <support@mwat.de>
*/

package db

//lint:file-ignore ST1017 - I prefer Yoda conditions

import (
	"context"
	"crypto/md5"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
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

	// _ = NewPool(mock, mock)
} // prepDBforTesting()

func TestTDBpool_Clear(t *testing.T) {
	ctx := context.Background()
	prepDBforTesting(ctx)

	conn, _ := pConnPool.get(ctx)
	_ = pConnPool.put(conn)

	tests := []struct {
		name string
		want *TDBpool
	}{
		// TODO: Add test cases.
		{" 1", pConnPool},
		{" 2", pConnPool},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := pConnPool.clear(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TDBpool.Clear() = %v, want %v", got, tt.want)
			}
		})
	}
} // TestTDBpool_Clear()

func TestTDBpool_Get(t *testing.T) {
	ctx := context.Background()
	prepDBforTesting(ctx)

	type args struct {
		aContext context.Context
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{" 1", args{ctx}, false},
		{" 2", args{ctx}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRConn, err := pConnPool.get(tt.args.aContext)
			if (err != nil) != tt.wantErr {
				t.Errorf("TDBpool.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if nil == gotRConn {
				t.Errorf("TDBpool.Get() = %v, want (!nil)", gotRConn)
			} else {
				_ = pConnPool.put(gotRConn)
			}
		})
	}
} // TestTDBpool_Get()

func TestTDBpool_Put(t *testing.T) {
	ctx := context.Background()
	prepDBforTesting(ctx)

	var conn1 *sql.DB
	conn2, _ := pConnPool.get(ctx)

	type args struct {
		aConnection *sql.DB
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		// TODO: Add test cases.
		{" 1", args{conn1}, 0},
		{" 2", args{conn2}, 1},
		{" 3", args{conn2}, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := pConnPool.put(tt.args.aConnection); got != tt.want {
				t.Errorf("TDBpool.Put() = %v, want %v", got, tt.want)
			}
		})
	}
} // TestTDBpool_Put()
