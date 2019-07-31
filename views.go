/*
   Copyright © 2019 M.Watermann, 10247 Berlin, Germany
              All rights reserved
          EMail : <support@mwat.de>
*/

package kaliber

/*
 * This file provides some template/view related functions and methods.
 */

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	// RegEx to HREF= tag attributes
	externalURLHrefRE = regexp.MustCompile(` (href="http)`)
)

const (
	// replacement text for `hrefRE`
	externalURLHrefReplace = ` target="_extern" $1`
)

// `addExternURLtagets()` adds a TARGET attribute to HREFs.
func addExternURLtagets(aPage []byte) []byte {
	return externalURLHrefRE.ReplaceAll(aPage, []byte(externalURLHrefReplace))
} // addExternURLtagets()

/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */

func authorList(aList *TEntityList) string {
	if nil == aList {
		return ""
	}
	result := ""
	for _, author := range *aList {
		result += author.Name + ", "
	}
	if strings.HasSuffix(result, ", ") {
		result = result[:len(result)-2]
	}

	return result + ": "
} // authorList()

// `htmlSafe()` returns `aText` as template.HTML.
func htmlSafe(aText string) template.HTML {
	return template.HTML(aText)
} // htmlSafe()

// `selectOption()` returns the OPTION markup for `aValue`.
func selectOption(aMap *TStringMap, aValue string) template.HTML {
	if result, ok := (*aMap)[aValue]; ok {
		return template.HTML(result)
	}

	return ""
} // selectOption()

var (
	viewFunctionMap = template.FuncMap{
		"authorList":   authorList,   // returns a comma separated author list
		"htmlSafe":     htmlSafe,     // returns `aText` as template.HTML
		"selectOption": selectOption, // returns a Select Option
	}
)

/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */

// TView combines a template and its logical name.
type TView struct {
	// The view's symbolic name.
	name string
	// The template as returned by a `NewView()` function call.
	tpl *template.Template
}

// NewView returns a new `TView` with `aName`.
//
// `aBaseDir` is the path to the directory storing the template files.
//
// `aName` is the name of the template file providing the page's main
// body without the filename extension (i.e. w/o ".gohtml"). `aName`
// serves as both the main template's name as well as the view's name.
func NewView(aBaseDir, aName string) (*TView, error) {
	bd, err := filepath.Abs(aBaseDir)
	if nil != err {
		return nil, err
	}
	files, err := filepath.Glob(bd + "/layout/*.gohtml")
	if nil != err {
		return nil, err
	}
	files = append(files, bd+"/"+aName+".gohtml")

	templ, err := template.New(aName).
		Funcs(viewFunctionMap).
		ParseFiles(files...)
	if nil != err {
		return nil, err
	}

	return &TView{
		name: aName,
		tpl:  templ,
	}, nil
} // NewView()

// `render()` is the core of `Render()` with a slightly different API
// (`io.Writer` instead of `http.ResponseWriter`) for easier testing.
func (v *TView) render(aWriter io.Writer, aData *TemplateData) (rErr error) {
	var page []byte

	if page, rErr = v.RenderedPage(aData); nil != rErr {
		return
	}
	_, rErr = aWriter.Write(addExternURLtagets(RemoveWhiteSpace(page)))
	// _, rErr = aWriter.Write(addExternURLtagets(page))

	return
} // render()

// Render executes the template using the TView's properties.
//
// `aWriter` is a http.ResponseWriter, or e.g. `os.Stdout` in console apps.
//
// `aData` is a list of data to be injected into the template.
//
// If an error occurs executing the template or writing its output,
// execution stops, and the method returns without writing anything
// to the output `aWriter`.
func (v *TView) Render(aWriter http.ResponseWriter, aData *TemplateData) error {
	return v.render(aWriter, aData)
} // Render()

// RenderedPage returns the rendered template/page and a possible Error
// executing the template.
//
// `aData` is a list of data to be injected into the template.
func (v *TView) RenderedPage(aData *TemplateData) (rBytes []byte, rErr error) {
	buf := &bytes.Buffer{}

	if rErr = v.tpl.ExecuteTemplate(buf, v.name, aData); nil != rErr {
		return
	}

	return buf.Bytes(), nil
} // RenderedPage()

/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */

type (
	tViewList map[string]*TView

	// TViewList is a list of `TView` instances (to be used as a template pool).
	TViewList tViewList
)

// NewViewList returns a new (empty) `TViewList` instance.
func NewViewList() *TViewList {
	result := make(TViewList, 16)

	return &result
} // NewViewlist()

// Add appends `aView` to the list.
//
// `aView` is the view to add to this list.
//
// The view's name (as specified in the `NewView()` function call)
// is used as the view's key in this list.
func (vl *TViewList) Add(aView *TView) *TViewList {
	(*vl)[aView.name] = aView

	return vl
} // Add()

// Get returns the view with `aName`.
//
// `aName` is the name (key) of the `TView` object to retrieve.
//
// If `aName` doesn't exist, the return value is `nil`.
// The second value (ok) is a `bool` that is `true` if `aName`
// exists in the list, and `false` if not.
func (vl *TViewList) Get(aName string) (*TView, bool) {
	if result, ok := (*vl)[aName]; ok {
		return result, true
	}

	return nil, false
} // Get()

// `render()` is the core of `Render()` with a slightly different API
// (`io.Writer` instead of `http.ResponseWriter`) for easier testing.
func (vl *TViewList) render(aName string, aWriter io.Writer, aData *TemplateData) error {
	if view, ok := (*vl)[aName]; ok {
		return view.render(aWriter, aData)
	}

	return fmt.Errorf("template/view '%s' not found", aName)
} // render()

// Render executes the template with the key `aName`.
//
// `aName` is the name of the template/view to use.
//
// `aWriter` is a `http.ResponseWriter` to handle the executed template.
//
// `aData` is a list of data to be injected into the template.
//
// If an error occurs executing the template or writing its output,
// execution stops, and the method returns without writing anything
// to the output `aWriter`.
func (vl *TViewList) Render(aName string, aWriter http.ResponseWriter, aData *TemplateData) error {
	return vl.render(aName, aWriter, aData)
} // Render()

// RenderedPage returns the rendered template/page with the key `aName`.
//
// `aName` is the name of the template/view to use.
//
// `aData` is a list of data to be injected into the template.
func (vl *TViewList) RenderedPage(aName string, aData *TemplateData) (rBytes []byte, rErr error) {

	if view, ok := (*vl)[aName]; ok {
		return view.RenderedPage(aData)
	}

	return rBytes, fmt.Errorf("template/view '%s' not found", aName)
} // RenderedPage()

/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */

func init() {
	initWSre()
} // init()

// Initialise the `whitespaceREs` list.
func initWSre() int {
	result := 0
	for idx, re := range whitespaceREs {
		whitespaceREs[idx].regEx = regexp.MustCompile(re.search)
		result++
	}

	return result
} // initWSre()

// `trimPREmatches()` removes leading/trailing whitespace from list entries.
func trimPREmatches(aList [][]byte) [][]byte {
	for idx, hit := range aList {
		aList[idx] = bytes.TrimSpace(hit)
	}

	return aList
} // trimPREmatches()

// Internal list of regular expressions used by
// the `RemoveWhiteSpace()` function.
type (
	tReItem struct {
		search  string
		replace string
		regEx   *regexp.Regexp
	}
	tReList []tReItem
)

var (
	// RegEx to find PREformatted parts in an HTML page.
	preRE = regexp.MustCompile(`(?si)\s*<pre[^>]*>.*?</pre>\s*`)

	// List of regular expressions matching different sets of HTML whitespace.
	whitespaceREs = tReList{
		// comments
		{`(?s)<!--.*?-->`, ``, nil},
		// HTML and HEAD elements:
		{`(?si)\s*(</?(body|\!DOCTYPE|head|html|link|meta|script|style|title)[^>]*>)\s*`, `$1`, nil},
		// block elements:
		{`(?si)\s+(</?(article|blockquote|div|footer|h[1-6]|header|nav|p|section)[^>]*>)`, `$1`, nil},
		{`(?si)(</?(article|blockquote|div|footer|h[1-6]|header|nav|p|section)[^>]*>)\s+`, `$1`, nil},
		// lists:
		{`(?si)\s+(</?([dou]l|li|d[dt])[^>]*>)`, `$1`, nil},
		{`(?si)(</?([dou]l|li|d[dt])[^>]*>)\s+`, `$1`, nil},
		// table elements:
		{`(?si)\s+(</?(col|t(able|body|foot|head|[dhr]))[^>]*>)`, `$1`, nil},
		{`(?si)(</?(col|t(able|body|foot|head|[dhr]))[^>]*>)\s+`, `$1`, nil},
		// form elements:
		{`(?si)\s+(</?(form|fieldset|legend|opt(group|ion))[^>]*>)`, `$1`, nil},
		{`(?si)(</?(form|fieldset|legend|opt(group|ion))[^>]*>)\s+`, `$1`, nil},
		// BR / HR:
		{`(?i)\s*(<[bh]r[^>]*>)\s*`, `$1`, nil},
		// whitespace after opened anchor:
		{`(?si)(<a\s+[^>]*>)\s+`, `$1`, nil},
		// preserve empty table cells:
		{`(?i)(<td(\s+[^>]*)?>)\s+(</td>)`, `$1&#160;$3`, nil},
		// remove empty paragraphs:
		{`(?i)<(p)(\s+[^>]*)?>\s*</$1>`, ``, nil},
		// whitespace before closing GT:
		{`\s+>`, `>`, nil},
	}
)

// RemoveWhiteSpace removes HTML comments and unnecessary whitespace.
//
// This function removes all unneeded/redundant whitespace
// and HTML comments from the given <tt>aPage</tt>.
// This can reduce significantly the amount of data to send to
// the remote user agent thus saving bandwidth.
func RemoveWhiteSpace(aPage []byte) []byte {
	var repl, search string

	// (0) Check whether there are PREformatted parts:
	preMatches := preRE.FindAll(aPage, -1)
	if (nil == preMatches) || (0 >= len(preMatches)) {
		// no PRE hence only the other REs to perform
		for _, reEntry := range whitespaceREs {
			aPage = reEntry.regEx.ReplaceAll(aPage, []byte(reEntry.replace))
		}
		return aPage
	}
	preMatches = trimPREmatches(preMatches)

	// Make sure PREformatted parts remain as-is.
	// (1) replace the PRE parts with a dummy text:
	for l, cnt := len(preMatches), 0; cnt < l; cnt++ {
		search = fmt.Sprintf(`\s*%s\s*`, regexp.QuoteMeta(string(preMatches[cnt])))
		if re, err := regexp.Compile(search); nil == err {
			repl = fmt.Sprintf(`</-%d-%d-%d-%d-/>`, cnt, cnt, cnt, cnt)
			aPage = re.ReplaceAllLiteral(aPage, []byte(repl))
		}
	}

	// (2) traverse through all the whitespace REs:
	for _, re := range whitespaceREs {
		aPage = re.regEx.ReplaceAll(aPage, []byte(re.replace))
	}

	// (3) replace the PRE dummies with the real markup:
	for l, cnt := len(preMatches), 0; cnt < l; cnt++ {
		search = fmt.Sprintf(`\s*</-%d-%d-%d-%d-/>\s*`, cnt, cnt, cnt, cnt)
		if re, err := regexp.Compile(search); nil == err {
			aPage = re.ReplaceAllLiteral(aPage, preMatches[cnt])
		}
	}
	// fmt.Println("Page3:", string(aPage))

	return aPage
} // RemoveWhiteSpace()

/* _EoF_ */
