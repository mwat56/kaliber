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
	s1 := `|3524|true|"authors"|0|0|25|0|""|100|2|0|`
	w1 := &TQueryOptions{
		ID:          3524,
		Descending:  true,
		Entity:      "authors",
		GuiLang:     qoLangGerman,
		Layout:      qoLayoutList,
		LimitLength: 25,
		LimitStart:  0,
		Matching:    "",
		QueryCount:  100,
		SortBy:      qoSortByAuthor,
		Theme:       qoThemeLight,
	}
	o2 := NewQueryOptions()
	s2 := `|1|false|"lang"|1|1|50|0|""|200|3|1|`
	w2 := &TQueryOptions{
		ID:          1,
		Descending:  false,
		Entity:      "lang",
		GuiLang:     qoLangEnglish,
		Layout:      qoLayoutGrid,
		LimitLength: 50,
		LimitStart:  0,
		Matching:    "",
		QueryCount:  200,
		SortBy:      qoSortByLanguage,
		Theme:       qoThemeDark,
	}
	o3 := NewQueryOptions()
	s3 := `|7607|true|"tags"|0|0|25|25|" "|6|1|0|"-"|`
	w3 := &TQueryOptions{
		ID:          7607,
		Descending:  true,
		Entity:      "tags",
		GuiLang:     qoLangGerman,
		Layout:      qoLayoutList,
		LimitLength: 25,
		LimitStart:  25,
		Matching:    "",
		QueryCount:  6,
		SortBy:      qoSortByAcquisition,
		Theme:       qoThemeLight,
		VirtLib:     "",
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
		{" 3", *o3, args{s3}, w3},
		{" 2", *o2, args{s2}, w2},
		{" 1", *o1, args{s1}, w1},
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
		`acquisition`: `<option value="acquisition">`,
		`authors`:     `<option SELECTED value="authors">`,
		`language`:    `<option value="language">`,
		`publisher`:   `<option value="publisher">`,
		`rating`:      `<option value="rating">`,
		`series`:      `<option value="series">`,
		`size`:        `<option value="size">`,
		`tags`:        `<option value="tags">`,
		`time`:        `<option value="time">`,
		`title`:       `<option value="title">`,
	}
	o2 := TQueryOptions{
		SortBy: qoSortByTime,
	}
	w2 := &TStringMap{
		`acquisition`: `<option value="acquisition">`,
		`authors`:     `<option value="authors">`,
		`language`:    `<option value="language">`,
		`publisher`:   `<option value="publisher">`,
		`rating`:      `<option value="rating">`,
		`series`:      `<option value="series">`,
		`size`:        `<option value="size">`,
		`tags`:        `<option value="tags">`,
		`time`:        `<option SELECTED value="time">`,
		`title`:       `<option value="title">`,
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
		Entity:      "authors",
		GuiLang:     qoLangEnglish,
		Layout:      qoLayoutList,
		LimitLength: 50,
		LimitStart:  0,
		Matching:    "",
		QueryCount:  100,
		SortBy:      qoSortByAuthor,
		Theme:       qoThemeDark,
	}
	w1 := `|3524|true|"authors"|1|0|50|0|""|100|2|1|""|`
	o2 := TQueryOptions{
		ID:          1,
		Descending:  false,
		Entity:      "lang",
		GuiLang:     qoLangGerman,
		Layout:      qoLayoutGrid,
		LimitLength: 25,
		LimitStart:  0,
		Matching:    "",
		QueryCount:  200,
		SortBy:      qoSortByLanguage,
		Theme:       qoThemeLight,
	}
	w2 := `|1|false|"lang"|0|1|25|0|""|200|3|0|""|`
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
} // TestTQueryOptions_String()

func TestTQueryOptions_SelectLimitOptions(t *testing.T) {
	qo1 := NewQueryOptions()
	w1 := `<option value="9">9</option>\n<option SELECTED value="24">24</option>\n<option value="48">48</option>\n<option value="99">99</option>\n<option value="249">249</option>\n<option value="498">498</option>`
	tests := []struct {
		name   string
		fields *TQueryOptions
		want   string
	}{
		// TODO: Add test cases.
		{" 1", qo1, w1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qo := tt.fields
			if got := qo.SelectLimitOptions(); got != tt.want {
				t.Errorf("TQueryOptions.SelectLimitOptions() = %v,\nwant %v", got, tt.want)
			}
		})
	}
} // TestTQueryOptions_SelectLimitOptions()
