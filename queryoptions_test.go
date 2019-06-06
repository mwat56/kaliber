/*
   Copyright © 2019 M.Watermann, 10247 Berlin, Germany
                  All rights reserved
               EMail : <support@mwat.de>
*/

package kaliber

import (
	"net/url"
	"reflect"
	"testing"
)

func TestTQueryOptions_CGI(t *testing.T) {
	o1 := TQueryOptions{
		ID:          3524,
		Descending:  true,
		Entity:      "author",
		LimitLength: 25,
		LimitStart:  0,
		Matching:    "",
		NextStart:   25,
		SortBy:      SortByAuthor,
	}
	w1 := `?qo="` + url.QueryEscape(`|3524|true|"author"|25|0|""|25|1|`) + `"`
	o2 := TQueryOptions{
		ID:          1,
		Descending:  false,
		Entity:      "lang",
		LimitLength: 50,
		LimitStart:  0,
		Matching:    "",
		NextStart:   50,
		SortBy:      SortByLanguage,
	}
	w2 := `?qo="` + url.QueryEscape(`|1|false|"lang"|50|0|""|50|2|`) + `"`
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
			if got := qo.CGI(); got != tt.want {
				t.Errorf("TQueryOptions.CGI() = '%v',\nwant '%v'", got, tt.want)
			}
		})
	}
} // TestTQueryOptions_CGI()

func TestTQueryOptions_Scan(t *testing.T) {
	o1 := NewQueryOptions()
	s1 := `|3524|true|"author"|25|0|""|25|1|`
	w1 := &TQueryOptions{
		ID:          3524,
		Descending:  true,
		Entity:      "author",
		LimitLength: 25,
		LimitStart:  0,
		Matching:    "",
		NextStart:   25,
		SortBy:      SortByAuthor,
	}
	o2 := NewQueryOptions()
	s2 := `|1|false|"lang"|50|0|""|50|2|`
	w2 := &TQueryOptions{
		ID:          1,
		Descending:  false,
		Entity:      "lang",
		LimitLength: 50,
		LimitStart:  0,
		Matching:    "",
		NextStart:   50,
		SortBy:      SortByLanguage,
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
		SortBy: SortByAuthor,
	}
	w1 := TStringMap{
		"author":    `SELECTED value="author"`,
		`language`:  `value="language"`,
		`publisher`: `value="publisher"`,
		`rating`:    `value="rating"`,
		`series`:    `value="series"`,
		`size`:      `value="size"`,
		`tags`:      `value="tags"`,
		`time`:      `value="time"`,
		`title`:     `value="title"`,
	}
	o2 := TQueryOptions{
		SortBy: SortByTime,
	}
	w2 := TStringMap{
		"author":    `value="author"`,
		`language`:  `value="language"`,
		`publisher`: `value="publisher"`,
		`rating`:    `value="rating"`,
		`series`:    `value="series"`,
		`size`:      `value="size"`,
		`tags`:      `value="tags"`,
		`time`:      `SELECTED value="time"`,
		`title`:     `value="title"`,
	}
	tests := []struct {
		name   string
		fields TQueryOptions
		want   TStringMap
	}{
		// TODO: Add test cases.
		{" 1", o1, w1},
		{" 2", o2, w2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qo := &tt.fields
			if got := qo.SortSelectOptions(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TQueryOptions.SortSelectOptions() = %v, want %v", got, tt.want)
			}
		})
	}
} // TestTQueryOptions_SortSelectOptions()

func TestTQueryOptions_String(t *testing.T) {
	o1 := TQueryOptions{
		ID:          3524,
		Descending:  true,
		Entity:      "author",
		LimitLength: 50,
		LimitStart:  0,
		Matching:    "",
		NextStart:   50,
		SortBy:      SortByAuthor,
	}
	w1 := `|3524|true|"author"|50|0|""|50|1|`
	o2 := TQueryOptions{
		ID:          1,
		Descending:  false,
		Entity:      "lang",
		LimitLength: 25,
		LimitStart:  0,
		Matching:    "",
		NextStart:   25,
		SortBy:      SortByLanguage,
	}
	w2 := `|1|false|"lang"|25|0|""|25|2|`
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
				t.Errorf("TQueryOptions.String() = '%s',\nwant '%s'", got, tt.want)
			}
		})
	}
} // TestTQueryOptions_String

func TestTQueryOptions_UnCGI(t *testing.T) {
	o1 := NewQueryOptions()
	c1 := url.QueryEscape(`|3524|true|"author"|50|0|""|50|1|`)
	w1 := &TQueryOptions{
		ID:          3524,
		Descending:  true,
		Entity:      "author",
		LimitLength: 50,
		LimitStart:  0,
		Matching:    "",
		NextStart:   50,
		SortBy:      SortByAuthor,
	}
	c2 := url.QueryEscape(`|1|false|"lang"|25|0|""|25|2|`)
	w2 := &TQueryOptions{
		ID:          1,
		Descending:  false,
		Entity:      "lang",
		LimitLength: 25,
		LimitStart:  0,
		Matching:    "",
		NextStart:   25,
		SortBy:      SortByLanguage,
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
				t.Errorf("TQueryOptions.UnCGI() = %v,\nwant %v", got, tt.want)
			}
		})
	}
} // TestTQueryOptions_UnCGI()
