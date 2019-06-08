/*
   Copyright © 2019 M.Watermann, 10247 Berlin, Germany
              All rights reserved
          EMail : <support@mwat.de>
*/

package kaliber

import (
	"regexp"
	"strings"
)

/*
 * This file provides functions and methods for fulltext search.
 */

type (
	// TSearch provides text search capabilities.
	TSearch struct {
		raw   string // the raw (unprocessed) search expression
		where string // used to build the WHERE clause
		next  string
	}

	tExpression struct {
		entity  string // the DB field to lookup
		matcher string // how to lookup
		not     bool   // flag negating the search result
		op      string // how to concat with the next expression
		term    string // what to lookup
	}
	tExpressionList []tExpression
)

/*
There are several forms to recognise:

"just a search term" -> lookup ALL book entities;
`entity:="searchterm"` -> lookup exact match of `searchterm` in `entity`;
`entity:~"searchterm"` -> lookup `searchterm` contained in `entity`.

All three expressions can be combined by AND and OR.
All three expressions can be negated by a leading `!`.
*/

// Clause returns the produced FROM/WHERE clause.
func (so *TSearch) Clause() string {
	if 0 < len(so.raw) {
		so.Parse()
	}
	if 0 < len(so.where) {
		return ` WHERE ` + so.where
	}

	return ""
} // Clause()

var (
	expressionRE = regexp.MustCompile(
		`(?i)^\s*((!?)(\w+):([=~]))?(["']([^"']*)["'])(\s+(AND|OR)\s*)?`)
	//           12   3     4       5    6            7   8

)

func (so *TSearch) getExpression() *tExpression {
	matches := expressionRE.FindStringSubmatch(so.raw)
	if (nil == matches) || (0 == len(matches)) {
		return nil
	}
	exp := &tExpression{
		entity:  strings.ToLower(matches[3]),
		matcher: matches[4],
		term:    matches[6],
		op:      matches[8],
		not:     ("!" == matches[2]),
	}
	so.raw = so.raw[len(matches[0]):]

	return exp
} // getExpression()

// Parse returns the parsed search term(s).
func (so *TSearch) Parse() *TSearch {
	if 0 == len(so.raw) {
		return so
	}
	so.where, so.next = "", ""

	for exp := so.getExpression(); nil != exp; exp = so.getExpression() {
		if 0 == len(exp.term) {
			continue
		}
		if 0 < len(exp.entity) {
			so.parsePrim(exp)
		} else {
			// lookup ALL possible entities
			exp.matcher = "~"
			exp.op = "OR"
			exp.entity = "author"
			so.parsePrim(exp)
			exp.entity = "comment"
			so.parsePrim(exp)
			exp.entity = "format"
			so.parsePrim(exp)
			exp.entity = "language"
			so.parsePrim(exp)
			exp.entity = "publisher"
			so.parsePrim(exp)
			exp.entity = "series"
			so.parsePrim(exp)
			exp.entity = "tag"
			so.parsePrim(exp)
			exp.entity = "title"
			exp.op = ""
			so.parsePrim(exp)
		}
		if 0 == len(so.raw) {
			break
		}
	}

	return so
} // Parse()

func (so *TSearch) parsePrim(aExpression *tExpression) {
	switch aExpression.entity {
	case "author":
		so.where += so.next + ` b.id IN (SELECT ba.book FROM books_authors_link ba JOIN authors a ON(ba.author = a.id) WHERE (a.name`

	case "comment":
		so.where += so.next + ` b.id IN (SELECT c.book FROM comments c WHERE (c.text`

	case "format":
		so.where += so.next + ` b.id IN (SELECT d.book FROM data d WHERE (d.format`

	case "language":
		so.where += so.next + ` b.id IN (SELECT bl.book FROM books_languages_link bl JOIN languages l ON(bl.lang_code = l.id) WHERE (l.lang_code`

	case "publisher":
		so.where += so.next + ` b.id IN (SELECT bp.book FROM books_publishers_link bp JOIN publishers p ON(bp.publisher = p.id) WHERE (p.name`

	case "series":
		so.where += so.next + ` b.id IN (SELECT bs.book FROM books_series_link bs JOIN series s ON(bs.series = s.id) WHERE (s.name`

	case "tag":
		so.where += so.next + ` b.id IN (SELECT bt.book FROM books_tags_link bt JOIN tags t ON(bt.tag = t.id) WHERE (t.name`

	case "title":
		so.where += so.next + ` (b.title`
		if "=" == aExpression.matcher {
			if aExpression.not {
				so.where += ` != `
			} else {
				so.where += ` = `
			}
			so.where += aExpression.term + `") `
		} else {
			if aExpression.not {
				so.where += ` NOT`
			}
			so.where += ` LIKE "%` + aExpression.term + `%") `
		}
		so.next = aExpression.op
		return

	default:
		return
	}
	if "=" == aExpression.matcher {
		so.where += ` = "` + aExpression.term + `")) `
	} else {
		so.where += ` LIKE "%` + aExpression.term + `%")) `
	}
	so.next = aExpression.op
} // parsePrim()

/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */

// NewSearch returns a new `TSearch` instance.
func NewSearch(aSearchTerm string) *TSearch {
	return &TSearch{raw: aSearchTerm}
} // NewSearch()

/* _EoF_ */
