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

func TestTSearch_getExpression(t *testing.T) {
	o1 := NewSearch(`tag:"="`)
	o2 := NewSearch(`AUTHOR:"=Spiegel"`)
	w2 := &tExpression{
		entity:  "author",
		matcher: "=",
		term:    "Spiegel",
	}
	o3 := NewSearch(`TITLE:"~Spiegel"`)
	w3 := &tExpression{
		entity:  "title",
		matcher: "~",
		term:    "Spiegel",
	}
	o4 := NewSearch(`"Der Spiegel"`)
	w4 := &tExpression{
		matcher: "~",
		term:    "Der Spiegel",
	}
	o5 := NewSearch(`title:"~Spiegel" and author:"=Spiegel"`)
	w5 := &tExpression{
		entity:  "title",
		matcher: "~",
		term:    "Spiegel",
		op:      "and",
	}
	o6 := NewSearch(`!title:"~Spiegel" and author:"=Spiegel"`)
	w6 := &tExpression{
		entity:  "title",
		matcher: "~",
		not:     true,
		term:    "Spiegel",
		op:      "and",
	}
	o7 := NewSearch(`Hallo, hallo!`)
	w7 := &tExpression{
		entity:  "",
		matcher: "~",
		not:     false,
		term:    "Hallo, hallo!",
		op:      "",
	}
	tests := []struct {
		name     string
		fields   *TSearch
		wantRExp *tExpression
	}{
		// TODO: Add test cases.
		{" 7", o7, w7},
		{" 1", o1, nil},
		{" 2", o2, w2},
		{" 3", o3, w3},
		{" 4", o4, w4},
		{" 5", o5, w5},
		{" 6", o6, w6},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			so := tt.fields
			gotRExp := so.getExpression()
			if !reflect.DeepEqual(gotRExp, tt.wantRExp) {
				t.Errorf("TSearch.getExpression() gotRExp = %v, want %v", gotRExp, tt.wantRExp)
			}
		})
	}
} // TestTSearch_getExpression()

func TestTSearch_Parse(t *testing.T) {
	o1 := NewSearch(`tag:"="`)
	w1 := &TSearch{raw: ``, where: ``, next: ``}
	o2 := NewSearch(`AUTHOR:"=Spiegel"`)
	w2 := &TSearch{
		where: ` b.id IN (SELECT ba.book FROM books_authors_link ba JOIN authors a ON(ba.author = a.id) WHERE (a.name = "Spiegel")) `,
	}
	o3 := NewSearch(`TITLE:"~Spiegel"`)
	w3 := &TSearch{
		where: ` (b.title LIKE "%Spiegel%") `,
	}
	o4 := NewSearch(`"Der Spiegel"`)
	w4 := &TSearch{
		where: ` b.id IN (SELECT ba.book FROM books_authors_link ba JOIN authors a ON(ba.author = a.id) WHERE (a.name LIKE "%Der Spiegel%")) OR b.id IN (SELECT c.book FROM comments c WHERE (c.text LIKE "%Der Spiegel%")) OR b.id IN (SELECT d.book FROM data d WHERE (d.format LIKE "%Der Spiegel%")) OR b.id IN (SELECT bl.book FROM books_languages_link bl JOIN languages l ON(bl.lang_code = l.id) WHERE (l.lang_code LIKE "%Der Spiegel%")) OR b.id IN (SELECT bp.book FROM books_publishers_link bp JOIN publishers p ON(bp.publisher = p.id) WHERE (p.name LIKE "%Der Spiegel%")) OR b.id IN (SELECT bs.book FROM books_series_link bs JOIN series s ON(bs.series = s.id) WHERE (s.name LIKE "%Der Spiegel%")) OR b.id IN (SELECT bt.book FROM books_tags_link bt JOIN tags t ON(bt.tag = t.id) WHERE (t.name LIKE "%Der Spiegel%")) OR (b.title LIKE "%Der Spiegel%") `,
	}
	o5 := NewSearch(`title:"~Spiegel" or author:"=Spiegel"`)
	w5 := &TSearch{
		where: ` (b.title LIKE "%Spiegel%") or b.id IN (SELECT ba.book FROM books_authors_link ba JOIN authors a ON(ba.author = a.id) WHERE (a.name = "Spiegel")) `,
	}
	tests := []struct {
		name   string
		fields *TSearch
		want   *TSearch
	}{
		// TODO: Add test cases.
		{" 1", o1, w1},
		{" 2", o2, w2},
		{" 3", o3, w3},
		{" 4", o4, w4},
		{" 5", o5, w5},
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

func TestTSearch_Clause(t *testing.T) {
	o0 := NewSearch(``)
	o1 := NewSearch(`tag:"="`)
	o2 := NewSearch(`AUTHOR:"=Spiegel"`)
	w2 := ` WHERE  b.id IN (SELECT ba.book FROM books_authors_link ba JOIN authors a ON(ba.author = a.id) WHERE (a.name = "Spiegel")) `
	o3 := NewSearch(`TITLE:"~Spiegel"`)
	w3 := ` WHERE  (b.title LIKE "%Spiegel%") `
	o4 := NewSearch(`"Der Spiegel"`)
	w4 := ` WHERE  b.id IN (SELECT ba.book FROM books_authors_link ba JOIN authors a ON(ba.author = a.id) WHERE (a.name LIKE "%Der Spiegel%")) OR b.id IN (SELECT c.book FROM comments c WHERE (c.text LIKE "%Der Spiegel%")) OR b.id IN (SELECT d.book FROM data d WHERE (d.format LIKE "%Der Spiegel%")) OR b.id IN (SELECT bl.book FROM books_languages_link bl JOIN languages l ON(bl.lang_code = l.id) WHERE (l.lang_code LIKE "%Der Spiegel%")) OR b.id IN (SELECT bp.book FROM books_publishers_link bp JOIN publishers p ON(bp.publisher = p.id) WHERE (p.name LIKE "%Der Spiegel%")) OR b.id IN (SELECT bs.book FROM books_series_link bs JOIN series s ON(bs.series = s.id) WHERE (s.name LIKE "%Der Spiegel%")) OR b.id IN (SELECT bt.book FROM books_tags_link bt JOIN tags t ON(bt.tag = t.id) WHERE (t.name LIKE "%Der Spiegel%")) OR (b.title LIKE "%Der Spiegel%") `
	o5 := NewSearch(`title:"~Spiegel" or author:"=Spiegel"`)
	w5 := ` WHERE  (b.title LIKE "%Spiegel%") or b.id IN (SELECT ba.book FROM books_authors_link ba JOIN authors a ON(ba.author = a.id) WHERE (a.name = "Spiegel")) `
	tests := []struct {
		name   string
		fields *TSearch
		want   string
	}{
		// TODO: Add test cases.
		{" 0", o0, ""},
		{" 1", o1, ""},
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
