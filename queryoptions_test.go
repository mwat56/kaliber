/*
   Copyright © 2019 M.Watermann, 10247 Berlin, Germany
                  All rights reserved
               EMail : <support@mwat.de>
*/

package kaliber

import (
	"encoding/base64"
	"reflect"
	"testing"
)

func TestTQueryOptions_CGI(t *testing.T) {
	o1 := TQueryOptions{
		ID:          3524,
		Descending:  true,
		Entity:      "author",
		Layout:      qoLayoutList,
		LimitLength: 25,
		LimitStart:  0,
		Matching:    "",
		QueryCount:  100,
		SortBy:      qoSortByAuthor,
	}
	w1 := `?qoc=` + base64.StdEncoding.EncodeToString([]byte(`|3524|true|"author"|0|25|0|""|100|1|`))
	o2 := TQueryOptions{
		ID:          1,
		Descending:  false,
		Entity:      "lang",
		Layout:      qoLayoutGrid,
		LimitLength: 50,
		LimitStart:  0,
		Matching:    "",
		QueryCount:  200,
		SortBy:      qoSortByLanguage,
	}
	w2 := `?qoc=` + base64.StdEncoding.EncodeToString([]byte(`|1|false|"lang"|1|50|0|""|200|2|`))
	o3 := TQueryOptions{
		ID:          1,
		Descending:  false,
		Entity:      "",
		Layout:      qoLayoutGrid,
		LimitLength: 50,
		LimitStart:  0,
		Matching:    `tag:"=Golang"`,
		QueryCount:  200,
		SortBy:      qoSortByLanguage,
	}
	w3 := `?qoc=` + base64.StdEncoding.EncodeToString([]byte(`|1|false|""|1|50|0|"tag:\"=Golang\""|200|2|`))
	tests := []struct {
		name   string
		fields TQueryOptions
		want   string
	}{
		// TODO: Add test cases.
		{" 1", o1, w1},
		{" 2", o2, w2},
		{" 3", o3, w3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qo := &tt.fields
			if got := qo.CGI(); got != tt.want {
				t.Errorf("\nTQueryOptions.CGI() = '%v',\nwant '%v'", got, tt.want)
			}
		})
	}
} // TestTQueryOptions_CGI()

func TestTQueryOptions_Scan(t *testing.T) {
	o1 := NewQueryOptions()
	s1 := `|3524|true|"author"|0|25|0|""|100|1|`
	w1 := &TQueryOptions{
		ID:          3524,
		Descending:  true,
		Entity:      "author",
		Layout:      qoLayoutList,
		LimitLength: 25,
		LimitStart:  0,
		Matching:    "",
		QueryCount:  100,
		SortBy:      qoSortByAuthor,
	}
	o2 := NewQueryOptions()
	s2 := `|1|false|"lang"|1|50|0|""|200|2|`
	w2 := &TQueryOptions{
		ID:          1,
		Descending:  false,
		Entity:      "lang",
		Layout:      qoLayoutGrid,
		LimitLength: 50,
		LimitStart:  0,
		Matching:    "",
		QueryCount:  200,
		SortBy:      qoSortByLanguage,
	}
	type args struct {
		aString string
	}
	tests := []struct {
		name   string
		fields TQueryOptions
		args   args
		want   *TQueryOptions
	}{
		// TODO: Add test cases.
		{" 1", *o1, args{s1}, w1},
		{" 2", *o2, args{s2}, w2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qo := &tt.fields
			if got := qo.Scan(tt.args.aString); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TQueryOptions.Scan() = %v,\nwant %v", got, tt.want)
			}
		})
	}
} // TestTQueryOptions_Scan()

func TestTQueryOptions_SortSelectOptions(t *testing.T) {
	o1 := TQueryOptions{
		SortBy: qoSortByAuthor,
	}
	w1 := &TStringMap{
		`author`:    `<option SELECTED value="author">`,
		`language`:  `<option value="language">`,
		`publisher`: `<option value="publisher">`,
		`rating`:    `<option value="rating">`,
		`series`:    `<option value="series">`,
		`size`:      `<option value="size">`,
		`tags`:      `<option value="tags">`,
		`time`:      `<option value="time">`,
		`title`:     `<option value="title">`,
	}
	o2 := TQueryOptions{
		SortBy: qoSortByTime,
	}
	w2 := &TStringMap{
		`author`:    `<option value="author">`,
		`language`:  `<option value="language">`,
		`publisher`: `<option value="publisher">`,
		`rating`:    `<option value="rating">`,
		`series`:    `<option value="series">`,
		`size`:      `<option value="size">`,
		`tags`:      `<option value="tags">`,
		`time`:      `<option SELECTED value="time">`,
		`title`:     `<option value="title">`,
	}
	tests := []struct {
		name   string
		fields TQueryOptions
		want   *TStringMap
	}{
		// TODO: Add test cases.
		{" 1", o1, w1},
		{" 2", o2, w2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qo := &tt.fields
			if got := qo.SelectSortByOptions(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TQueryOptions.SortSelectOptions() = %v,\nwant %v", got, tt.want)
			}
		})
	}
} // TestTQueryOptions_SortSelectOptions()

func TestTQueryOptions_String(t *testing.T) {
	o1 := TQueryOptions{
		ID:          3524,
		Descending:  true,
		Entity:      "author",
		Layout:      qoLayoutList,
		LimitLength: 50,
		LimitStart:  0,
		Matching:    "",
		QueryCount:  100,
		SortBy:      qoSortByAuthor,
	}
	w1 := `|3524|true|"author"|0|50|0|""|100|1|`
	o2 := TQueryOptions{
		ID:          1,
		Descending:  false,
		Entity:      "lang",
		Layout:      qoLayoutGrid,
		LimitLength: 25,
		LimitStart:  0,
		Matching:    "",
		QueryCount:  200,
		SortBy:      qoSortByLanguage,
	}
	w2 := `|1|false|"lang"|1|25|0|""|200|2|`
	tests := []struct {
		name   string
		fields TQueryOptions
		want   string
	}{
		// TODO: Add test cases.
		{" 1", o1, w1},
		{" 2", o2, w2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qo := &tt.fields
			if got := qo.String(); got != tt.want {
				t.Errorf("\nTQueryOptions.String() = '%s',\nwant '%s'", got, tt.want)
			}
		})
	}
} // TestTQueryOptions_String

func TestTQueryOptions_UnCGI(t *testing.T) {
	o1 := NewQueryOptions()
	c1 := base64.StdEncoding.EncodeToString([]byte(`|3524|true|"author"|0|50|0|""|100|1|`))
	w1 := &TQueryOptions{
		ID:          3524,
		Descending:  true,
		Entity:      "author",
		Layout:      qoLayoutList,
		LimitLength: 50,
		LimitStart:  0,
		Matching:    "",
		QueryCount:  100,
		SortBy:      qoSortByAuthor,
	}
	c2 := base64.StdEncoding.EncodeToString([]byte(`|1|false|"lang"|1|25|0|""|200|2|`))
	w2 := &TQueryOptions{
		ID:          1,
		Descending:  false,
		Entity:      "lang",
		Layout:      qoLayoutGrid,
		LimitLength: 25,
		LimitStart:  0,
		Matching:    "",
		QueryCount:  200,
		SortBy:      qoSortByLanguage,
	}
	type args struct {
		aCGI string
	}
	tests := []struct {
		name   string
		fields TQueryOptions
		args   args
		want   *TQueryOptions
	}{
		// TODO: Add test cases.
		{" 1", *o1, args{c1}, w1},
		{" 2", *o1, args{c2}, w2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qo := &tt.fields
			if got := qo.UnCGI(tt.args.aCGI); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("\nTQueryOptions.UnCGI() = `%v`,\nwant `%v`", got, tt.want)
			}
		})
	}
} // TestTQueryOptions_UnCGI()
