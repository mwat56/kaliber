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

func Test_tExpression_buildSQL(t *testing.T) {
	ex1 := tExpression{
		entity:  "author",
		matcher: "~",
		not:     false,
		op:      "",
		term:    "Watermann",
	}
	w1 := `(b.id IN (SELECT ba.book FROM books_authors_link ba JOIN authors a ON(ba.author = a.id) WHERE (a.name LIKE "%Watermann%")))`
	ex2 := tExpression{
		entity:  "Genre",
		matcher: "~",
		not:     false,
		op:      "",
		term:    "Computer.*",
	}
	w2 := `(b.id IN (SELECT lt.book FROM books_custom_column_1_link lt JOIN custom_column_1 t ON(lt.value = t.id) WHERE (t.value LIKE "%Computer.*%")))`
	ex3 := tExpression{
		entity:  "#hyph",
		matcher: "=",
		not:     false,
		op:      "",
		term:    "yes",
	}
	w3 := `(b.id IN (SELECT lt.book FROM books_custom_column_3_link lt JOIN custom_column_3 t ON(lt.value = t.id) WHERE (t.value = "yes")))`
	tests := []struct {
		name       string
		fields     tExpression
		wantRWhere string
	}{
		// TODO: Add test cases.
		{" 1", ex1, w1},
		{" 2", ex2, w2},
		{" 3", ex3, w3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exp := &tt.fields
			if gotRWhere := exp.buildSQL(); gotRWhere != tt.wantRWhere {
				t.Errorf("tExpression.buildSQL() = %v,\nwant %v", gotRWhere, tt.wantRWhere)
			}
		})
	}
} // Test_tExpression_buildSQL()

func TestTSearch_Clause(t *testing.T) {
	o0 := NewSearch(``)
	o1 := NewSearch(`tags:"="`)
	w1 := ` WHERE (b.id IN (SELECT bt.book FROM books_tags_link bt JOIN tags t ON(bt.tag = t.id) WHERE (t.name = "")))`
	o2 := NewSearch(`AUTHORS:"=Spiegel"`)
	w2 := ` WHERE (b.id IN (SELECT ba.book FROM books_authors_link ba JOIN authors a ON(ba.author = a.id) WHERE (a.name = "Spiegel")))`
	o3 := NewSearch(`TITLE:"~Spiegel"`)
	w3 := ` WHERE (b.title LIKE "%Spiegel%")`
	o4 := NewSearch(`Der Spiegel`)
	w4 := ` WHERE (b.id IN (SELECT ba.book FROM books_authors_link ba JOIN authors a ON(ba.author = a.id) WHERE (a.name LIKE "%Der Spiegel%")))OR(b.id IN (SELECT c.book FROM comments c WHERE (c.text LIKE "%Der Spiegel%")))OR(b.id IN (SELECT d.book FROM data d WHERE (d.format LIKE "%Der Spiegel%")))OR(b.id IN (SELECT bl.book FROM books_languages_link bl JOIN languages l ON(bl.lang_code = l.id) WHERE (l.lang_code LIKE "%Der Spiegel%")))OR(b.id IN (SELECT bp.book FROM books_publishers_link bp JOIN publishers p ON(bp.publisher = p.id) WHERE (p.name LIKE "%Der Spiegel%")))OR(b.id IN (SELECT bs.book FROM books_series_link bs JOIN series s ON(bs.series = s.id) WHERE (s.name LIKE "%Der Spiegel%")))OR(b.id IN (SELECT bt.book FROM books_tags_link bt JOIN tags t ON(bt.tag = t.id) WHERE (t.name LIKE "%Der Spiegel%")))OR(b.title LIKE "%Der Spiegel%")`
	o5 := NewSearch(`title:"~Spiegel" or authors:"=Spiegel"`)
	w5 := ` WHERE (b.title LIKE "%Spiegel%")OR (b.id IN (SELECT ba.book FROM books_authors_link ba JOIN authors a ON(ba.author = a.id) WHERE (a.name = "Spiegel")))`
	tests := []struct {
		name   string
		fields *TSearch
		want   string
	}{
		// TODO: Add test cases.
		{" 0", o0, ""},
		{" 1", o1, w1},
		{" 2", o2, w2},
		{" 3", o3, w3},
		{" 4", o4, w4},
		{" 5", o5, w5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			so := tt.fields
			if got := so.Clause(); got != tt.want {
				t.Errorf("TSearch.Clause() = `%s`,\nwant `%s`", got, tt.want)
			}
		})
	}
} // TestTSearch_Clause()

func TestTSearch_p1(t *testing.T) {
	o1 := NewSearch(`tags:"=Golang"`)
	w1 := &TSearch{
		where: `(b.id IN (SELECT bt.book FROM books_tags_link bt JOIN tags t ON(bt.tag = t.id) WHERE (t.name = "Golang")))`,
	}
	o2 := NewSearch("(title:\"~c't\" OR series:\"~c't\") AND publisher:\"~Heise\"")
	w2 := &TSearch{
		where: `((b.title LIKE "%c't%")OR (b.id IN (SELECT bs.book FROM books_series_link bs JOIN series s ON(bs.series = s.id) WHERE (s.name LIKE "%c't%")))) AND (b.id IN (SELECT bp.book FROM books_publishers_link bp JOIN publishers p ON(bp.publisher = p.id) WHERE (p.name LIKE "%Heise%")))`,
	}
	o3 := NewSearch("tags:\"=myths & legends\" OR tags:\"=Myth\" OR tags:\"=Myth History\" OR tags:\"=Mythical\" OR tags:\"=Mythical Civilizations\" OR tags:\"=Mythology\" OR tags:\"=Myths\"")
	w3 := &TSearch{
		where: `(b.id IN (SELECT bt.book FROM books_tags_link bt JOIN tags t ON(bt.tag = t.id) WHERE (t.name = "myths & legends")))OR (b.id IN (SELECT bt.book FROM books_tags_link bt JOIN tags t ON(bt.tag = t.id) WHERE (t.name = "Myth")))OR (b.id IN (SELECT bt.book FROM books_tags_link bt JOIN tags t ON(bt.tag = t.id) WHERE (t.name = "Myth History")))OR (b.id IN (SELECT bt.book FROM books_tags_link bt JOIN tags t ON(bt.tag = t.id) WHERE (t.name = "Mythical")))OR (b.id IN (SELECT bt.book FROM books_tags_link bt JOIN tags t ON(bt.tag = t.id) WHERE (t.name = "Mythical Civilizations")))OR (b.id IN (SELECT bt.book FROM books_tags_link bt JOIN tags t ON(bt.tag = t.id) WHERE (t.name = "Mythology")))OR (b.id IN (SELECT bt.book FROM books_tags_link bt JOIN tags t ON(bt.tag = t.id) WHERE (t.name = "Myths")))`,
	}
	o4 := NewSearch("tags:\"=Philosophy\" or tags:\"=Philosophy : Ethics & Moral Philosophy\" or tags:\"=Philosophy : General\" or tags:\"=Philosophy (General)\" or tags:\"=Philosophy (Specific Aspects)\" or tags:\"=Philosophy & Social Aspects\" or tags:\"=Philosophy Of Architecture\"")
	w4 := &TSearch{
		where: `(b.id IN (SELECT bt.book FROM books_tags_link bt JOIN tags t ON(bt.tag = t.id) WHERE (t.name = "Philosophy")))OR (b.id IN (SELECT bt.book FROM books_tags_link bt JOIN tags t ON(bt.tag = t.id) WHERE (t.name = "Philosophy : Ethics & Moral Philosophy")))OR (b.id IN (SELECT bt.book FROM books_tags_link bt JOIN tags t ON(bt.tag = t.id) WHERE (t.name = "Philosophy : General")))OR (b.id IN (SELECT bt.book FROM books_tags_link bt JOIN tags t ON(bt.tag = t.id) WHERE (t.name = "Philosophy (General)")))OR (b.id IN (SELECT bt.book FROM books_tags_link bt JOIN tags t ON(bt.tag = t.id) WHERE (t.name = "Philosophy (Specific Aspects)")))OR (b.id IN (SELECT bt.book FROM books_tags_link bt JOIN tags t ON(bt.tag = t.id) WHERE (t.name = "Philosophy & Social Aspects")))OR (b.id IN (SELECT bt.book FROM books_tags_link bt JOIN tags t ON(bt.tag = t.id) WHERE (t.name = "Philosophy Of Architecture")))`,
	}
	o5 := NewSearch(`tags:"=Golang" Programming`)
	w5 := &TSearch{
		where: `(b.id IN (SELECT bt.book FROM books_tags_link bt JOIN tags t ON(bt.tag = t.id) WHERE (t.name = "Golang")))OR (b.id IN (SELECT ba.book FROM books_authors_link ba JOIN authors a ON(ba.author = a.id) WHERE (a.name LIKE "%Programming%")))OR(b.id IN (SELECT c.book FROM comments c WHERE (c.text LIKE "%Programming%")))OR(b.id IN (SELECT d.book FROM data d WHERE (d.format LIKE "%Programming%")))OR(b.id IN (SELECT bl.book FROM books_languages_link bl JOIN languages l ON(bl.lang_code = l.id) WHERE (l.lang_code LIKE "%Programming%")))OR(b.id IN (SELECT bp.book FROM books_publishers_link bp JOIN publishers p ON(bp.publisher = p.id) WHERE (p.name LIKE "%Programming%")))OR(b.id IN (SELECT bs.book FROM books_series_link bs JOIN series s ON(bs.series = s.id) WHERE (s.name LIKE "%Programming%")))OR(b.id IN (SELECT bt.book FROM books_tags_link bt JOIN tags t ON(bt.tag = t.id) WHERE (t.name LIKE "%Programming%")))OR(b.title LIKE "%Programming%")`,
	}
	o6 := NewSearch(`!tags:"=Golang" and Programming`)
	w6 := &TSearch{
		where: `(b.id IN (SELECT bt.book FROM books_tags_link bt JOIN tags t ON(bt.tag = t.id) WHERE (t.name != "Golang")))AND(b.id IN (SELECT ba.book FROM books_authors_link ba JOIN authors a ON(ba.author = a.id) WHERE (a.name LIKE "%Programming%")))OR(b.id IN (SELECT c.book FROM comments c WHERE (c.text LIKE "%Programming%")))OR(b.id IN (SELECT d.book FROM data d WHERE (d.format LIKE "%Programming%")))OR(b.id IN (SELECT bl.book FROM books_languages_link bl JOIN languages l ON(bl.lang_code = l.id) WHERE (l.lang_code LIKE "%Programming%")))OR(b.id IN (SELECT bp.book FROM books_publishers_link bp JOIN publishers p ON(bp.publisher = p.id) WHERE (p.name LIKE "%Programming%")))OR(b.id IN (SELECT bs.book FROM books_series_link bs JOIN series s ON(bs.series = s.id) WHERE (s.name LIKE "%Programming%")))OR(b.id IN (SELECT bt.book FROM books_tags_link bt JOIN tags t ON(bt.tag = t.id) WHERE (t.name LIKE "%Programming%")))OR(b.title LIKE "%Programming%")`,
	}
	o7 := NewSearch("tags:\"~Magic.\" or #genre:\"~Magic.\"")
	w7 := &TSearch{
		where: `(b.id IN (SELECT bt.book FROM books_tags_link bt JOIN tags t ON(bt.tag = t.id) WHERE (t.name LIKE "%Magic.%")))OR (b.id IN (SELECT lt.book FROM books_custom_column_1_link lt JOIN custom_column_1 t ON(lt.value = t.id) WHERE (t.value LIKE "%Magic.%")))`,
	}
	o8 := NewSearch(" tags:\"~Magic.\" ")
	w8 := &TSearch{
		where: `(b.id IN (SELECT bt.book FROM books_tags_link bt JOIN tags t ON(bt.tag = t.id) WHERE (t.name LIKE "%Magic.%")))`,
	}
	tests := []struct {
		name   string
		fields *TSearch
		want   *TSearch
	}{
		// TODO: Add test cases.
		{" 8", o8, w8},
		{" 7", o7, w7},
		{" 6", o6, w6},
		{" 5", o5, w5},
		{" 4", o4, w4},
		{" 3", o3, w3},
		{" 2", o2, w2},
		{" 1", o1, w1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			so := tt.fields
			if got := so.p1(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TSearch.p1() = %v,\nwant %v", got, tt.want)
			}
		})
	}
} // TestTSearch_p1()

func TestTSearch_Parse(t *testing.T) {
	o1 := NewSearch(`tags:"="`)
	w1 := &TSearch{where: `(b.id IN (SELECT bt.book FROM books_tags_link bt JOIN tags t ON(bt.tag = t.id) WHERE (t.name = "")))`}
	o2 := NewSearch(`AUTHORS:"=Spiegel"`)
	w2 := &TSearch{
		where: `(b.id IN (SELECT ba.book FROM books_authors_link ba JOIN authors a ON(ba.author = a.id) WHERE (a.name = "Spiegel")))`,
	}
	o3 := NewSearch(`TITLE:"~Spiegel"`)
	w3 := &TSearch{
		where: `(b.title LIKE "%Spiegel%")`,
	}
	o4 := NewSearch(`Der Spiegel`)
	w4 := &TSearch{
		where: `(b.id IN (SELECT ba.book FROM books_authors_link ba JOIN authors a ON(ba.author = a.id) WHERE (a.name LIKE "%Der Spiegel%")))OR(b.id IN (SELECT c.book FROM comments c WHERE (c.text LIKE "%Der Spiegel%")))OR(b.id IN (SELECT d.book FROM data d WHERE (d.format LIKE "%Der Spiegel%")))OR(b.id IN (SELECT bl.book FROM books_languages_link bl JOIN languages l ON(bl.lang_code = l.id) WHERE (l.lang_code LIKE "%Der Spiegel%")))OR(b.id IN (SELECT bp.book FROM books_publishers_link bp JOIN publishers p ON(bp.publisher = p.id) WHERE (p.name LIKE "%Der Spiegel%")))OR(b.id IN (SELECT bs.book FROM books_series_link bs JOIN series s ON(bs.series = s.id) WHERE (s.name LIKE "%Der Spiegel%")))OR(b.id IN (SELECT bt.book FROM books_tags_link bt JOIN tags t ON(bt.tag = t.id) WHERE (t.name LIKE "%Der Spiegel%")))OR(b.title LIKE "%Der Spiegel%")`,
	}
	o5 := NewSearch(`title:"~Spiegel" or authors:"=Spiegel"`)
	w5 := &TSearch{
		where: `(b.title LIKE "%Spiegel%")OR (b.id IN (SELECT ba.book FROM books_authors_link ba JOIN authors a ON(ba.author = a.id) WHERE (a.name = "Spiegel")))`,
	}
	o6 := NewSearch(`What's going on here?`)
	w6 := &TSearch{
		where: `(b.id IN (SELECT ba.book FROM books_authors_link ba JOIN authors a ON(ba.author = a.id) WHERE (a.name LIKE "%What's going on here?%")))OR(b.id IN (SELECT c.book FROM comments c WHERE (c.text LIKE "%What's going on here?%")))OR(b.id IN (SELECT d.book FROM data d WHERE (d.format LIKE "%What's going on here?%")))OR(b.id IN (SELECT bl.book FROM books_languages_link bl JOIN languages l ON(bl.lang_code = l.id) WHERE (l.lang_code LIKE "%What's going on here?%")))OR(b.id IN (SELECT bp.book FROM books_publishers_link bp JOIN publishers p ON(bp.publisher = p.id) WHERE (p.name LIKE "%What's going on here?%")))OR(b.id IN (SELECT bs.book FROM books_series_link bs JOIN series s ON(bs.series = s.id) WHERE (s.name LIKE "%What's going on here?%")))OR(b.id IN (SELECT bt.book FROM books_tags_link bt JOIN tags t ON(bt.tag = t.id) WHERE (t.name LIKE "%What's going on here?%")))OR(b.title LIKE "%What's going on here?%")`,
	}
	o7 := NewSearch(` `)
	w7 := &TSearch{
		where: ``,
	}
	o8 := NewSearch(`lang:"=eng"`)
	w8 := &TSearch{
		where: ``,
	}
	tests := []struct {
		name   string
		fields *TSearch
		want   *TSearch
	}{
		// TODO: Add test cases.
		{" 8", o8, w8},
		{" 7", o7, w7},
		{" 6", o6, w6},
		{" 5", o5, w5},
		{" 4", o4, w4},
		{" 3", o3, w3},
		{" 2", o2, w2},
		{" 1", o1, w1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			so := tt.fields
			if got := so.Parse(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TSearch.Parse() =\n'%v',\nwant\n'%v'", got, tt.want)
			}
		})
	}
} // TestTSearch_Parse()

func TestTSearch_String(t *testing.T) {
	// type fields struct {
	// 	raw   string
	// 	where string
	// 	next  string
	// }
	s1 := NewSearch("")
	w1 := `raw: '' | where: '' | next: ''`
	s2 := NewSearch("search term")
	w2 := `raw: 'search term' | where: '' | next: ''`
	s3 := NewSearch(`title:"~Spiegel"`).Parse()
	w3 := `raw: '' | where: '(b.title LIKE "%Spiegel%")' | next: ''`
	tests := []struct {
		name   string
		fields *TSearch
		want   string
	}{
		// TODO: Add test cases.
		{" 1", s1, w1},
		{" 2", s2, w2},
		{" 3", s3, w3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			so := tt.fields
			if got := so.String(); got != tt.want {
				t.Errorf("TSearch.String() = %v,\nwant %v", got, tt.want)
			}
		})
	}
} // TestTSearch_String()
