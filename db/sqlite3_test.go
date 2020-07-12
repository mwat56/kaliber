/*
   Copyright © 2019, 2020 M.Watermann, 10247 Berlin, Germany
                  All rights reserved
               EMail : <support@mwat.de>
*/

package db

//lint:file-ignore ST1017 - I prefer Yoda conditions

import (
	"context"
	"crypto/md5"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func openDBforTesting(aContext context.Context) *TDataBase {
	libPath := `/var/opt/Calibre`
	s := fmt.Sprintf("%x", md5.Sum([]byte(libPath))) // #nosec G401
	ucd, _ := os.UserCacheDir()
	SetCalibreCachePath(filepath.Join(ucd, "kaliber", s))
	SetCalibreLibraryPath(libPath)
	result, err := OpenDatabase(aContext)
	if nil != err {
		log.Fatalf("OpenDatabase(): %v", err)
	}
	SetSQLtraceFile("./SQLtrace.sql")

	return result
} // openDBforTesting()

func Test_prepAuthors(t *testing.T) {
	w0 := &tAuthorList{}
	a1 := "Willy Wichtig|1"
	w1 := &tAuthorList{
		TEntity{
			ID:   1,
			Name: "Willy Wichtig",
		},
	}
	a2 := "Ayn Rand|1108, Nathaniel Branden|2270"
	w2 := &tAuthorList{
		TEntity{
			ID:   1108,
			Name: "Ayn Rand",
		},
		TEntity{
			ID:   2270,
			Name: "Nathaniel Branden",
		},
	}
	type args struct {
		aAuthor tPSVstring
	}
	tests := []struct {
		name string
		args args
		want *tAuthorList
	}{
		// TODO: Add test cases.
		{" 0", args{""}, w0},
		{" 1", args{a1}, w1},
		{" 2", args{a2}, w2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := prepAuthors(tt.args.aAuthor); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("prepAuthors() = '%v',\nwant '%v'", got, tt.want)
			}
		})
	}
} // Test_prepAuthors()

func Test_prepFormats(t *testing.T) {
	w0 := &tFormatList{}
	f1 := "EPUB|1"
	w1 := &tFormatList{
		TEntity{
			ID:   1,
			Name: "EPUB",
		},
	}
	f2 := "AZW3|3, EPUB|1, PDF|2"
	w2 := &tFormatList{
		TEntity{
			ID:   3,
			Name: "AZW3",
		},
		TEntity{
			ID:   1,
			Name: "EPUB",
		},
		TEntity{
			ID:   2,
			Name: "PDF",
		},
	}
	type args struct {
		aFormat tPSVstring
	}
	tests := []struct {
		name string
		args args
		want *tFormatList
	}{
		// TODO: Add test cases.
		{" 0", args{""}, w0},
		{" 1", args{f1}, w1},
		{" 2", args{f2}, w2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := prepFormats(tt.args.aFormat); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("prepFormats() = %v, want %v", got, tt.want)
			}
		})
	}
} // Test_prepFormats()

func Test_prepIdentifiers(t *testing.T) {
	i1 := "amazon_de|25139|1783988029, barnesnoble|25138|w/go-programming-blueprints-mat-ryer/1120876061, google|25137|op4crgEACAAJ, isbn|25136|9781783988020"
	w1 := &tIdentifierList{
		TEntity{
			ID:   25139,
			Name: "amazon_de",
			URL:  "1783988029",
		},
		TEntity{
			ID:   25138,
			Name: "barnesnoble",
			URL:  "w/go-programming-blueprints-mat-ryer/1120876061",
		},
		TEntity{
			ID:   25137,
			Name: "google",
			URL:  "op4crgEACAAJ",
		},
		TEntity{
			ID:   25136,
			Name: "isbn",
			URL:  "9781783988020",
		},
	}
	type args struct {
		aIdentifier tPSVstring
	}
	tests := []struct {
		name string
		args args
		want *tIdentifierList
	}{
		// TODO: Add test cases.
		{" 1", args{i1}, w1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := prepIdentifiers(tt.args.aIdentifier); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("prepIdentifiers() = %v,\nwant %v", got, tt.want)
			}
		})
	}
} // Test_prepIdentifiers()

func Test_prepPages(t *testing.T) {
	SetCalibreLibraryPath("/var/opt/Calibre/")
	d1 := TDocument{
		path: "Spiegel/Der Spiegel (2019-06-01) 23_2019 (7628)",
	}
	d2 := TDocument{
		path: "John Scalzi/Zoe's Tale (6730)",
	}
	type args struct {
		aPath string
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		// TODO: Add test cases.
		{" 1", args{d1.path}, 130},
		{" 2", args{d2.path}, 569},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := prepPages(tt.args.aPath); got != tt.want {
				t.Errorf("prepPages() = %v, want %v", got, tt.want)
			}
		})
	}
} // Test_prepPages()

func Test_prepPublisher(t *testing.T) {
	p1 := ""
	var w1 *tPublisher
	p2 := "Imagine Publishing|1228"
	w2 := &tPublisher{
		ID:   1228,
		Name: "Imagine Publishing",
	}
	p3 := "|1228, "
	w3 := &tPublisher{
		ID: 1228}
	type args struct {
		aPublisher tPSVstring
	}
	tests := []struct {
		name string
		args args
		want *tPublisher
	}{
		// TODO: Add test cases.
		{" 1", args{p1}, w1},
		{" 2", args{p2}, w2},
		{" 3", args{p3}, w3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := prepPublisher(tt.args.aPublisher); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("prepPublisher() = %v, want %v", got, tt.want)
			}
		})
	}
} // Test_prepPublisher()

func Test_prepSeries(t *testing.T) {
	s1 := ""
	var w1 *tSeries
	s2 := "The Dresden Files|36"
	w2 := &tSeries{
		ID:   36,
		Name: "The Dresden Files",
	}
	s3 := ", The Dresden Files|36, "
	w3 := &tSeries{
		ID:   36,
		Name: "The Dresden Files",
	}
	type args struct {
		aSeries tPSVstring
	}
	tests := []struct {
		name string
		args args
		want *tSeries
	}{
		// TODO: Add test cases.
		{" 1", args{s1}, w1},
		{" 2", args{s2}, w2},
		{" 3", args{s3}, w3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := prepSeries(tt.args.aSeries); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("prepSeries() = %v, want %v", got, tt.want)
			}
		})
	}
} // Test_prepSeries()

func Test_prepTags(t *testing.T) {
	w0 := &tTagList{}
	t1 := "Atag|1234"
	w1 := &tTagList{
		TEntity{
			ID:   1234,
			Name: "Atag",
		},
	}
	t2 := "Computer|6177, eBook.Management|6193, Dtag|7890"
	w2 := &tTagList{
		TEntity{
			ID:   6177,
			Name: "Computer",
		},
		TEntity{
			ID:   7890,
			Name: "Dtag",
		},
		TEntity{
			ID:   6193,
			Name: "eBook.Management",
		},
	}
	type args struct {
		aTag tPSVstring
	}
	tests := []struct {
		name string
		args args
		want *tTagList
	}{
		// TODO: Add test cases.
		{" 0", args{""}, w0},
		{" 1", args{t1}, w1},
		{" 2", args{t2}, w2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := prepTags(tt.args.aTag); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("prepTags() = %v, want %v", got, tt.want)
			}
		})
	}
} // Test_prepTags()

func TestOpenDatabase(t *testing.T) {
	ctx := context.TODO()
	libPath := `/var/opt/Calibre`
	s := fmt.Sprintf("%x", md5.Sum([]byte(libPath))) // #nosec G401
	ucd, _ := os.UserCacheDir()
	SetCalibreCachePath(filepath.Join(ucd, "kaliber", s))
	SetCalibreLibraryPath(libPath)

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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRDB, err := OpenDatabase(tt.args.aContext)
			if (err != nil) != tt.wantErr {
				t.Errorf("OpenDatabase() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if nil == gotRDB {
				t.Errorf("OpenDatabase() = {%v},\nwant {!NIL}", gotRDB)
			}
		})
	}
} // TestOpenDatabase()

func TestTDataBase_QueryBy(t *testing.T) {
	ctx := context.TODO()
	dbHandle := openDBforTesting(ctx)

	o0 := &TQueryOptions{
		ID:          0,
		Descending:  false,
		Entity:      "",
		Layout:      QoLayoutGrid,
		LimitLength: 1000,
		LimitStart:  0,
		Matching:    "",
		SortBy:      qoSortByAuthor,
	}
	o1 := &TQueryOptions{
		ID:          3524,
		Descending:  false,
		Entity:      "authors",
		Layout:      QoLayoutGrid,
		LimitLength: 50,
		LimitStart:  0,
		Matching:    "",
		SortBy:      qoSortByRating,
	}
	o2 := &TQueryOptions{
		ID:          1,
		Descending:  false,
		Entity:      "languages",
		Layout:      QoLayoutGrid,
		LimitLength: 50,
		LimitStart:  0,
		Matching:    "",
		SortBy:      qoSortByLanguage,
	}
	o3 := &TQueryOptions{
		ID:          574,
		Descending:  false,
		Entity:      "publisher",
		Layout:      QoLayoutGrid,
		LimitLength: 50,
		LimitStart:  0,
		Matching:    "",
		SortBy:      qoSortByPublisher,
	}
	o4 := &TQueryOptions{
		ID:          519,
		Descending:  false,
		Entity:      "series",
		Layout:      QoLayoutGrid,
		LimitLength: 50,
		LimitStart:  0,
		Matching:    "",
		SortBy:      qoSortByTime,
	}
	o5 := &TQueryOptions{
		ID:          60,
		Descending:  false,
		Entity:      "tags",
		LimitLength: 50,
		LimitStart:  0,
		Matching:    "",
		SortBy:      qoSortByTags,
	}
	o6 := o5.clone()
	o6.Layout = QoLayoutGrid
	o6.SortBy = qoSortByAcquisition

	type args struct {
		aContext context.Context
		aOptions *TQueryOptions
	}
	tests := []struct {
		name       string
		args       args
		wantRCount int
		wantRList  int // *TDocList
		wantErr    bool
	}{
		// TODO: Add test cases.
		{" 0", args{ctx, o0}, 5626, 1000, false},
		{" 1", args{ctx, o1}, 14, 14, false},
		{" 2", args{ctx, o2}, 4631, 50, false},
		{" 3", args{ctx, o3}, 42, 42, false},
		{" 4", args{ctx, o4}, 408, 50, false},
		{" 5", args{ctx, o5}, 447, 50, false},
		{" 5", args{ctx, o6}, 447, 50, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRCount, gotRList, err := dbHandle.QueryBy(tt.args.aContext, tt.args.aOptions)
			if (err != nil) != tt.wantErr {
				t.Errorf("TDataBase.QueryBy() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotRCount != tt.wantRCount {
				t.Errorf("TDataBase.QueryBy() gotRCount = %v, want %v", gotRCount, tt.wantRCount)
			}
			// if !reflect.DeepEqual(gotRList, tt.wantRList) {
			// 	t.Errorf("TDataBase.QueryBy() gotRList = %v, want %v", gotRList, tt.wantRList)
			// }
			if len(*gotRList) != tt.wantRList {
				t.Errorf("TDataBase.QueryBy() gotRList = %d, want %d", len(*gotRList), tt.wantRList)
			}
		})
	}
} // TestTDataBase_QueryBy

func TestTDataBase_QueryCustomColumns(t *testing.T) {
	ctx := context.TODO()
	dbHandle := openDBforTesting(ctx)
	w1 := &TCustomColumnList{}

	type args struct {
		aContext context.Context
	}
	tests := []struct {
		name    string
		args    args
		want    *TCustomColumnList
		wantErr bool
	}{
		// TODO: Add test cases.
		{" 1", args{ctx}, w1, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := dbHandle.QueryCustomColumns(tt.args.aContext)
			if (err != nil) != tt.wantErr {
				t.Errorf("TDataBase.QueryCustomColumns() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// if !reflect.DeepEqual(got, tt.want) {
			// 	t.Errorf("TDataBase.QueryCustomColumns() = %v, want %v", got, tt.want)
			// }
			if nil == got {
				t.Errorf("TDataBase.QueryCustomColumns() = %v,\nwant %s", got, `!=NIL`)
			}
		})
	}
} // TestTDataBase_QueryCustomColumns()

func TestTDataBase_QueryDocMini(t *testing.T) {
	ctx := context.TODO()
	dbHandle := openDBforTesting(ctx)

	type args struct {
		aContext context.Context
		aID      TID
	}
	tests := []struct {
		name     string
		args     args
		wantRDoc *TDocument
	}{
		// TODO: Add test cases.
		{" 1", args{ctx, 0}, nil},
		{" 2", args{ctx, 1}, NewDocument()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := dbHandle.QueryDocMini(tt.args.aContext, tt.args.aID)
			if (nil == got) && (nil != tt.wantRDoc) {
				t.Errorf("TDataBase.QueryDocMini() = %v,\nwant %v", got, tt.wantRDoc)
			}
			if (nil != got) && (nil == tt.wantRDoc) {
				t.Errorf("TDataBase.QueryDocMini() = %v,\nwant %v", got, tt.wantRDoc)
			}
		})
	}
} // TestTDataBase_QueryDocMini()

func TestTDataBase_QueryDocument(t *testing.T) {
	ctx := context.TODO()
	dbHandle := openDBforTesting(ctx)

	type args struct {
		aContext context.Context
		aID      TID
	}
	tests := []struct {
		name string
		args args
		want bool // *TDocument
	}{
		// TODO: Add test cases.
		{" 0", args{ctx, -1}, false},
		{" 1", args{ctx, 1}, true},
		{" 2", args{ctx, 2}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := dbHandle.QueryDocument(tt.args.aContext, tt.args.aID); (nil != got) != tt.want {
				t.Errorf("TDataBase.QueryDocument() = %v, want %v", got, tt.want)
			}
		})
	}
} // TestTDataBase_QueryDocument()

func TestTDataBase_QueryIDs(t *testing.T) {
	ctx := context.TODO()
	dbHandle := openDBforTesting(ctx)

	type args struct {
		aContext context.Context
	}
	tests := []struct {
		name      string
		args      args
		wantRList int // *TDocList
		wantErr   bool
	}{
		// TODO: Add test cases.
		{" 1", args{ctx}, 5626, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRList, err := dbHandle.QueryIDs(tt.args.aContext)
			if (err != nil) != tt.wantErr {
				t.Errorf("TDataBase.QueryIDs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotLen := len(*gotRList); gotLen != tt.wantRList {
				t.Errorf("TDataBase.QueryIDs() = %v, want %v", gotLen, tt.wantRList)
			}
		})
	}
} // TestTDataBase_QueryIDs()

func TestTDataBase_QuerySearch(t *testing.T) {
	ctx := context.TODO()
	dbHandle := openDBforTesting(ctx)

	qo1 := NewQueryOptions(0)
	qo1.Layout = QoLayoutGrid
	qo1.Matching = `Golang`
	qo1.SortBy = qoSortByLanguage
	qo2 := NewQueryOptions(0)
	qo2.Layout = QoLayoutGrid
	qo2.Matching = `languages:"=eng"`
	qo2.SortBy = qoSortBySize
	qo3 := NewQueryOptions(0)
	qo3.Layout = QoLayoutGrid
	qo3.Matching = `languages:"=deu"`
	qo3.SortBy = qoSortBySeries

	type args struct {
		aContext context.Context
		aOptions *TQueryOptions
	}
	tests := []struct {
		name       string
		args       args
		wantRCount int
		wantRList  int // *TDocList
		wantErr    bool
	}{
		// TODO: Add test cases.
		{" 1", args{ctx, qo1}, 38, 24, false},
		{" 2", args{ctx, qo2}, 4631, 24, false},
		{" 3", args{ctx, qo3}, 982, 24, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRCount, gotRList, err := dbHandle.QuerySearch(tt.args.aContext, tt.args.aOptions)
			if (err != nil) != tt.wantErr {
				t.Errorf("TDataBase.QuerySearch() error = %v,\nwantErr %v", err, tt.wantErr)
				return
			}
			if gotRCount != tt.wantRCount {
				t.Errorf("TDataBase.QuerySearch() gotRCount = %v,\nwant %v", gotRCount, tt.wantRCount)
			}
			// if !reflect.DeepEqual(gotRList, tt.wantRList) {
			// 	t.Errorf("TDataBase.QuerySearch() gotRList = %v,\nwant %v", gotRList, tt.wantRList)
			// }
			if (nil != gotRList) && (len(*gotRList) != tt.wantRList) {
				t.Errorf("TDataBase.QuerySearch() gotRList = %d,\nwant %d", len(*gotRList), tt.wantRList)
			}
		})
	}
} // TestTDataBase_QuerySearch()
