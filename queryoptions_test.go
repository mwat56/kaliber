/*
   Copyright © 2019 M.Watermann, 10247 Berlin, Germany
                  All rights reserved
               EMail : <support@mwat.de>
*/

package kaliber

import (
	"reflect"
	"testing"
)

func TestTQueryOptions_Scan(t *testing.T) {
	o1 := NewQueryOptions()
	s1 := `|3524|true|"author"|0|25|0|""|100|1|0|`
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
		Theme:       qoThemeLight,
	}
	o2 := NewQueryOptions()
	s2 := `|1|false|"lang"|1|50|0|""|200|2|1|`
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
		Theme:       qoThemeDark,
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
		Theme:       qoThemeDark,
	}
	w1 := `|3524|true|"author"|0|50|0|""|100|1|1|`
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
		Theme:       qoThemeLight,
	}
	w2 := `|1|false|"lang"|1|25|0|""|200|2|0|`
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
