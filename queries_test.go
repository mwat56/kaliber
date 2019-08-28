/*
   Copyright © 2019 M.Watermann, 10247 Berlin, Germany
                 All rights reserved
              EMail : <support@mwat.de>
*/

package kaliber

import (
	"log"
	"reflect"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func openDBforTesting() {
	SetCalibreLibraryPath("/var/opt/Calibre/")
	if err := DBopen(CalibreDatabasePath()); nil != err {
		log.Fatalf("DBopen: %v", err)
	}
	SetSQLtraceFile("./SQLtrace.sql")
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
		aAuthor tCSVstring
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
		aFormat tCSVstring
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
		aIdentifier tCSVstring
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
		aPublisher tCSVstring
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
		aSeries tCSVstring
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
		aTag tCSVstring
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

func TestDBopen(t *testing.T) {
	SetCalibreLibraryPath(`/var/opt/Calibre`)
	dbfn := CalibreDatabasePath()
	type args struct {
		aFilename string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{" 1", args{dbfn}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := DBopen(tt.args.aFilename); (err != nil) != tt.wantErr {
				t.Errorf("DBopen() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
} // TestDBopen()

func Test_queryDocument(t *testing.T) {
	openDBforTesting()

	type args struct {
		aID int
	}
	tests := []struct {
		name string
		args args
		want bool // *TDocument
	}{
		// TODO: Add test cases.
		{" 1", args{1}, true},
		{" 2", args{2}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := QueryDocument(tt.args.aID); (nil != got) != tt.want {
				t.Errorf("QueryDocument() = %v, want %v", got, tt.want)
			}
		})
	}
} // Test_queryDocument()

func TestQueryBy(t *testing.T) {
	openDBforTesting()
	o0 := &TQueryOptions{
		ID:          0,
		Descending:  false,
		Entity:      "",
		LimitLength: 1000,
		LimitStart:  0,
		Matching:    "",
		SortBy:      qoSortByAuthor,
	}
	o1 := &TQueryOptions{
		ID:          3524,
		Descending:  false,
		Entity:      "author",
		LimitLength: 50,
		LimitStart:  0,
		Matching:    "",
		SortBy:      qoSortByAuthor,
	}
	o2 := &TQueryOptions{
		ID:          1,
		Descending:  false,
		Entity:      "lang",
		LimitLength: 50,
		LimitStart:  0,
		Matching:    "",
		SortBy:      qoSortByLanguage,
	}
	o3 := &TQueryOptions{
		ID:          574,
		Descending:  false,
		Entity:      "publisher",
		LimitLength: 50,
		LimitStart:  0,
		Matching:    "",
		SortBy:      qoSortByPublisher,
	}
	o4 := &TQueryOptions{
		ID:          519,
		Descending:  false,
		Entity:      "series",
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
	type args struct {
		aOption *TQueryOptions
	}
	tests := []struct {
		name       string
		args       args
		wantRCount int
		wantRList  int //*TDocList
		wantErr    bool
	}{
		// TODO: Add test cases.
		{" 0", args{o0}, 5474, 1000, false},
		{" 1", args{o1}, 14, 14, false},
		{" 2", args{o2}, 4589, 50, false},
		{" 3", args{o3}, 42, 42, false},
		{" 4", args{o4}, 361, 50, false},
		{" 5", args{o5}, 452, 50, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRCount, gotRList, err := QueryBy(tt.args.aOption)
			if (err != nil) != tt.wantErr {
				t.Errorf("QueryBy() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotRCount != tt.wantRCount {
				t.Errorf("QueryBy() gotRCount = %v, want %v", gotRCount, tt.wantRCount)
			}
			if len(*gotRList) != tt.wantRList {
				t.Errorf("QueryBy() gotRList = %d, want %d", len(*gotRList), tt.wantRList)
			}
		})
	}
} // TestQueryBy()

/*
func TestQueryLimit(t *testing.T) {
	openDBforTesting()
	type args struct {
		aStart  uint
		aLength uint
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		// TODO: Add test cases.
		{" 1", args{0, 500}, 500, false},
		{" 2", args{500, 500}, 500, false},
		{" 3", args{1000, 500}, 500, false},
		{" 4", args{1500, 500}, 500, false},
		{" 5", args{2000, 500}, 500, false},
		{" 6", args{2500, 500}, 500, false},
		{" 7", args{3000, 500}, 500, false},
		{" 7", args{3500, 500}, 500, false},
		{" 8", args{4000, 500}, 500, false},
		{" 9", args{4500, 500}, 500, false},
		{"10", args{5000, 500}, 475, false},
		{"11", args{5500, 500}, 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := QueryLimit(tt.args.aStart, tt.args.aLength)
			if (err != nil) != tt.wantErr {
				t.Errorf("QueryLimit() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(*got) != tt.want {
				t.Errorf("QueryLimit() = %d, want %d", len(*got), tt.want)
			}
		})
	}
} // TestQueryLimit()
*/

func TestQuerySearch(t *testing.T) {
	openDBforTesting()
	qo1 := NewQueryOptions()
	qo1.Matching = `Golang`
	type args struct {
		aOption *TQueryOptions
	}
	tests := []struct {
		name       string
		args       args
		wantRCount int
		wantRList  int //*TDocList
		wantErr    bool
	}{
		// TODO: Add test cases.
		{" 1", args{qo1}, 31, 25, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRCount, gotRList, err := QuerySearch(tt.args.aOption)
			if (err != nil) != tt.wantErr {
				t.Errorf("QuerySearch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotRCount != tt.wantRCount {
				t.Errorf("QuerySearch() gotRCount = %v, want %v", gotRCount, tt.wantRCount)
			}
			if (nil != gotRList) && (len(*gotRList) != tt.wantRList) {
				t.Errorf("QuerySearch() gotRList = %d, want %d", len(*gotRList), tt.wantRList)
			}
		})
	}
} // TestQuerySearch()

func Test_escapeQuery(t *testing.T) {
	type args struct {
		source string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
		{" 1", args{""}, ""},
		{" 2", args{"Hello World!"}, "Hello World!"},
		{" 3", args{`"Hello World!"`}, `\"Hello World!\"`},
		{" 4", args{`"Rock 'n' Roll!"`}, `\"Rock 'n' Roll!\"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := escapeQuery(tt.args.source); got != tt.want {
				t.Errorf("escape() = %v, want %v", got, tt.want)
			}
		})
	}
} // Test_escapeQuery()
