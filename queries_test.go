/*
   Copyright © 2019 M.Watermann, 10247 Berlin, Germany
                 All rights reserved
              EMail : <support@mwat.de>
*/

package kaliber

import (
	"path/filepath"
	"reflect"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func openDB() {
	dir, _ := filepath.Abs("./")
	SetCalibreLibraryPath(dir)
	DBopen(CalibreDatabasePath())
	SetCalibreLibraryPath("/var/opt/Calibre/")
} // openDB()

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

func Test_queryAuthor(t *testing.T) {
	openDB()

	o1 := &TQueryOptions{
		ID:          3524,
		Descending:  false,
		LimitLength: 10,
		LimitStart:  0,
		Matching:    "",
		SortBy:      SortByTitle,
	}
	o2 := &TQueryOptions{
		ID:          0,
		Descending:  false,
		LimitLength: 25,
		LimitStart:  0,
		Matching:    "",
		SortBy:      SortByTitle,
	}
	o3 := &TQueryOptions{
		ID:          3524,
		Descending:  false,
		LimitLength: 10,
		LimitStart:  10,
		Matching:    "",
		SortBy:      SortByTitle,
	}
	type args struct {
		aOption *TQueryOptions
	}
	tests := []struct {
		name    string
		args    args
		want    int //*TDocList
		wantErr bool
	}{
		// TODO: Add test cases.
		{" 1", args{o1}, 10, false},
		{" 2", args{o2}, 0, false},
		{" 3", args{o3}, 4, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := queryAuthor(tt.args.aOption)
			if (err != nil) != tt.wantErr {
				t.Errorf("queryAuthor() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(*got) != tt.want {
				t.Errorf("queryAuthor() = %d, want %d", len(*got), tt.want)
			}
		})
	}
} // Test_queryAuthor()

func Test_queryPublisher(t *testing.T) {
	openDB()

	o1 := &TQueryOptions{
		ID:          574,
		Descending:  false,
		LimitLength: 25,
		LimitStart:  0,
		Matching:    "",
		SortBy:      SortByAuthor,
	}
	o2 := &TQueryOptions{
		ID:          574,
		Descending:  false,
		LimitLength: 25,
		LimitStart:  25,
		Matching:    "",
		SortBy:      SortByAuthor,
	}
	type args struct {
		aOption *TQueryOptions
	}
	tests := []struct {
		name    string
		args    args
		want    int //*TDocList
		wantErr bool
	}{
		// TODO: Add test cases.
		{" 1", args{o1}, 25, false},
		{" 2", args{o2}, 17, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := queryPublisher(tt.args.aOption)
			if (err != nil) != tt.wantErr {
				t.Errorf("queryPublisher() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(*got) != tt.want {
				t.Errorf("queryPublisher() = %d, want %d", len(*got), tt.want)
			}
		})
	}
} // Test_queryPublisher()

func Test_querySeries(t *testing.T) {
	openDB()

	o1 := &TQueryOptions{
		ID:          519,
		Descending:  false,
		LimitLength: 100,
		LimitStart:  0,
		Matching:    "",
		SortBy:      SortByTime,
	}
	o2 := &TQueryOptions{
		ID:          519,
		Descending:  false,
		LimitLength: 100,
		LimitStart:  100,
		Matching:    "",
		SortBy:      SortByTime,
	}
	o3 := &TQueryOptions{
		ID:          519,
		Descending:  false,
		LimitLength: 100,
		LimitStart:  200,
		Matching:    "",
		SortBy:      SortByTime,
	}
	o4 := &TQueryOptions{
		ID:          519,
		Descending:  false,
		LimitLength: 100,
		LimitStart:  300,
		Matching:    "",
		SortBy:      SortByTime,
	}
	o5 := &TQueryOptions{
		ID:          519,
		Descending:  false,
		LimitLength: 100,
		LimitStart:  400,
		Matching:    "",
		SortBy:      SortByTime,
	}
	type args struct {
		aOption *TQueryOptions
	}
	tests := []struct {
		name    string
		args    args
		want    int //*TDocList
		wantErr bool
	}{
		// TODO: Add test cases.
		{" 1", args{o1}, 100, false},
		{" 2", args{o2}, 100, false},
		{" 3", args{o3}, 100, false},
		{" 4", args{o4}, 50, false},
		{" 5", args{o5}, 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := querySeries(tt.args.aOption)
			if (err != nil) != tt.wantErr {
				t.Errorf("querySeries() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(*got) != tt.want {
				t.Errorf("querySeries() = %d, want %d", len(*got), tt.want)
			}
		})
	}
} // Test_querySeries()

func Test_queryTag(t *testing.T) {
	openDB()

	o1 := &TQueryOptions{
		ID:          60,
		Descending:  false,
		LimitLength: 100,
		LimitStart:  0,
		Matching:    "",
		SortBy:      SortByAuthor,
	}
	o2 := &TQueryOptions{
		ID:          60,
		Descending:  false,
		LimitLength: 100,
		LimitStart:  100,
		Matching:    "",
		SortBy:      SortByAuthor,
	}
	o3 := &TQueryOptions{
		ID:          60,
		Descending:  false,
		LimitLength: 100,
		LimitStart:  200,
		Matching:    "",
		SortBy:      SortByAuthor,
	}
	o4 := &TQueryOptions{
		ID:          60,
		Descending:  false,
		LimitLength: 100,
		LimitStart:  300,
		Matching:    "",
		SortBy:      SortByAuthor,
	}
	o5 := &TQueryOptions{
		ID:          60,
		Descending:  false,
		LimitLength: 100,
		LimitStart:  400,
		Matching:    "",
		SortBy:      SortByAuthor,
	}
	o6 := &TQueryOptions{
		ID:          60,
		Descending:  false,
		LimitLength: 100,
		LimitStart:  500,
		Matching:    "",
		SortBy:      SortByAuthor,
	}
	type args struct {
		aOption *TQueryOptions
	}
	tests := []struct {
		name    string
		args    args
		want    int //*TDocList
		wantErr bool
	}{
		// TODO: Add test cases.
		{" 1", args{o1}, 100, false},
		{" 2", args{o2}, 100, false},
		{" 3", args{o3}, 100, false},
		{" 4", args{o4}, 100, false},
		{" 5", args{o5}, 48, false},
		{" 6", args{o6}, 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := queryTag(tt.args.aOption)
			if (err != nil) != tt.wantErr {
				t.Errorf("queryTag() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(*got) != tt.want {
				t.Errorf("queryTag() = %d, want %d", len(*got), tt.want)
			}
		})
	}
} // Test_queryTag

func TestDBopen(t *testing.T) {
	dir, _ := filepath.Abs("./")
	SetCalibreLibraryPath(dir)
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

/*
func TestQueryAuthor(t *testing.T) {
	openDB()

	type args struct {
		aID         int
		aStart      uint
		aLength     uint
		aDescending bool
	}
	tests := []struct {
		name    string
		args    args
		want    int //*TDocList
		wantErr bool
	}{
		// TODO: Add test cases.
		{" 1", args{960, 0, 20, false}, 2, false},
		{" 2", args{1258, 0, 20, false}, 3, false},
		{" 3", args{3758, 0, 20, false}, 6, false},
		{" 4", args{3524, 0, 10, false}, 10, false},
		{" 5", args{3524, 10, 10, false}, 4, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := QueryAuthor(tt.args.aID, tt.args.aStart, tt.args.aLength, tt.args.aDescending)
			if (err != nil) != tt.wantErr {
				t.Errorf("QueryAuthor() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(*got) != tt.want {
				t.Errorf("QueryAuthor() = %v, want %v", len(*got), tt.want)
			}
		})
	}
} // TestQueryAuthor()
*/

func Test_queryDocument(t *testing.T) {
	openDB()

	type args struct {
		aID int
	}
	tests := []struct {
		name string
		args args
		want bool // *TDocument
	}{
		// TODO: Add test cases.
		// {" 1", args{1}, true},
		{" 2", args{2}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := queryDocument(tt.args.aID); (nil != got) != tt.want {
				t.Errorf("QueryDocument() = %v, want %v", got, tt.want)
			}
		})
	}
} // Test_queryDocument()

func TestQueryLimit(t *testing.T) {
	openDB()
	type args struct {
		aStart  uint
		aLength uint
	}
	tests := []struct {
		name    string
		args    args
		want    int // *TDocList
		wantErr bool
	}{
		// TODO: Add test cases.
		// {" 1", args{0, 500}, 500, false},
		// {" 2", args{500, 500}, 500, false},
		// {" 3", args{1000, 500}, 500, false},
		// {" 4", args{1500, 500}, 500, false},
		// {" 5", args{2000, 500}, 500, false},
		// {" 6", args{2500, 500}, 500, false},
		// {" 7", args{3000, 500}, 500, false},
		// {" 7", args{3500, 500}, 500, false},
		// {" 8", args{4000, 500}, 500, false},
		// {" 9", args{4500, 500}, 500, false},
		{"10", args{5000, 500}, 438, false},
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
				t.Errorf("QueryLimit() = %v, want %v", len(*got), tt.want)
			}
		})
	}
} // TestQueryLimit()

/*
func TestQueryPublisher(t *testing.T) {
	openDB()
	type args struct {
		aID         int
		aStart      uint
		aLength     uint
		aDescending bool
	}
	tests := []struct {
		name    string
		args    args
		want    int //*TDocList
		wantErr bool
	}{
		// TODO: Add test cases.
		{" 1", args{71, 0, 200, false}, 200, false},
		{" 2", args{71, 200, 200, false}, 190, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := QueryPublisher(tt.args.aID, tt.args.aStart, tt.args.aLength, tt.args.aDescending)
			if (err != nil) != tt.wantErr {
				t.Errorf("QueryPublisher() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(*got) != tt.want {
				t.Errorf("QueryPublisher() = %d, want %d", len(*got), tt.want)
			}
		})
	}
} // TestQueryPublisher()

*/
/*
func TestQuerySeries(t *testing.T) {
	openDB()
	type args struct {
		aID         int
		aStart      uint
		aLength     uint
		aDescending bool
	}
	tests := []struct {
		name    string
		args    args
		want    int // *TDocList
		wantErr bool
	}{
		// TODO: Add test cases.
		{" 1", args{519, 0, 200, false}, 200, false},
		{" 2", args{519, 200, 200, false}, 150, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := QuerySeries(tt.args.aID, tt.args.aStart, tt.args.aLength, tt.args.aDescending)
			if (err != nil) != tt.wantErr {
				t.Errorf("QuerySeries() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(*got) != tt.want {
				t.Errorf("QuerySeries() = %d, want %d", len(*got), tt.want)
			}
		})
	}
} // TestQuerySeries
*/

/*
func TestQueryTag(t *testing.T) {
	openDB()
	type args struct {
		aID         int
		aStart      uint
		aLength     uint
		aDescending bool
	}
	tests := []struct {
		name    string
		args    args
		want    int // *TDocList
		wantErr bool
	}{
		// TODO: Add test cases.
		{" 1", args{109, 0, 50, false}, 50, false},
		{" 2", args{109, 50, 50, false}, 16, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := QueryTag(tt.args.aID, tt.args.aStart, tt.args.aLength, tt.args.aDescending)
			if (err != nil) != tt.wantErr {
				t.Errorf("QueryTag() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(*got) != tt.want {
				t.Errorf("QueryTag() = %d, want %d", len(*got), tt.want)
			}
		})
	}
} // TestQueryTag()
*/
