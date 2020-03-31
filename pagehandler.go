/*
   Copyright © 2019, 2020 M.Watermann, 10247 Berlin, Germany
                  All rights reserved
               EMail : <support@mwat.de>
*/

package kaliber

//lint:file-ignore ST1017 - I prefer Yoda conditions

import (
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"path/filepath"
	"regexp"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/mwat56/apachelogger"
	"github.com/mwat56/cssfs"
	"github.com/mwat56/jffs"
	"github.com/mwat56/kaliber/db"
	"github.com/mwat56/passlist"
	"github.com/mwat56/sessions"
)

type (
	// TPageHandler provides the handling of HTTP request/response.
	TPageHandler struct {
		addr        string              // listen address ("1.2.3.4:5678")
		authAll     bool                // authenticate user for all pages and documents
		cacheFS     http.Handler        // cache file server (i.e. thumbnails)
		cssFS       http.Handler        // CSS file server
		dataDir     string              // base dir for data
		docFS       http.Handler        // document file server
		docsPerPage int                 // number of documents shown per web-page
		lang        string              // default GUI language
		libName     string              // the library's name
		logStack    bool                // log stack trace
		realm       string              // host/domain to secure by BasicAuth
		staticFS    http.Handler        // static file server
		theme       string              // `dark` or `light` display theme
		usrList     *passlist.TPassList // user/password list
		viewList    *TViewList          // list of template/views
	}
)

// NewPageHandler returns a new `TPageHandler` instance.
func NewPageHandler() (*TPageHandler, error) {
	var (
		err error
		s   string
	)
	result := new(TPageHandler)

	result.cacheFS = jffs.FileServer(db.CalibreCachePath())

	if s, err = AppArguments.Get("authAll"); nil == err {
		result.authAll = ("true" == s)
	}

	result.docsPerPage = 24
	if s, _ = AppArguments.Get("booksPerPage"); 0 < len(s) {
		if bpp, er2 := strconv.Atoi(s); nil == er2 {
			result.docsPerPage = bpp
		}
	}

	if s, err = AppArguments.Get("dataDir"); nil != err {
		return nil, err
	}
	result.dataDir = s
	result.cssFS = cssfs.FileServer(s + `/`)
	result.docFS = jffs.FileServer(db.CalibreLibraryPath())

	if s, err = AppArguments.Get("lang"); nil == err {
		result.lang = s
	}

	if s, err = AppArguments.Get("libraryName"); nil == err {
		result.libName = s
	}

	result.addr, _ = AppArguments.Get("listen")
	// an empty value means: listen on all interfaces

	if s, err = AppArguments.Get("logStack"); nil == err {
		result.logStack = ("true" == s)
	}

	if s, err = AppArguments.Get("port"); nil != err {
		return nil, err
	}
	result.addr += ":" + s

	result.staticFS = jffs.FileServer(result.dataDir)

	if s, err = AppArguments.Get("uf"); nil != err {
		s = fmt.Sprintf("%v\nAUTHENTICATION DISABLED!", err)
		apachelogger.Err("NewPageHandler()", s)
	} else if result.usrList, err = passlist.LoadPasswords(s); nil != err {
		s = fmt.Sprintf("%v\nAUTHENTICATION DISABLED!", err)
		apachelogger.Err("NewPageHandler()", s)
		result.usrList = nil
	}

	if s, err = AppArguments.Get("realm"); nil == err {
		result.realm = s
	}

	if s, err = AppArguments.Get("theme"); nil != err {
		result.theme = "dark"
	} else {
		result.theme = s
	}

	if result.viewList, err = newViewList(filepath.Join(result.dataDir, "views")); nil != err {
		return nil, err
	}

	// update the thumbnails cache:
	go ThumbnailUpdate()

	// avoid sessions for certain requests:
	sessions.ExcludePaths("/certs", "/css/", "/favicon", "/file/", "/fonts", "/img/", "/robots")

	return result, nil
} // NewPageHandler()

// `newViewList()` returns a list of views found in `aDirectory`
// and a possible I/O error.
func newViewList(aDirectory string) (*TViewList, error) {
	var v *TView
	result := NewViewList()

	files, err := filepath.Glob(aDirectory + "/*.gohtml")
	if err != nil {
		return nil, err
	}

	for _, fName := range files {
		fName = filepath.Base(fName[:len(fName)-7]) // remove extension
		if v, err = NewView(aDirectory, fName); nil != err {
			return nil, err
		}
		result = result.Add(v)
	}

	return result, nil
} // newViewList()

// `recoverPanic()` is called by `TPageHandler.ServeHTTP()` to
// recover from a panic.
func recoverPanic(doLogStack bool) {
	if err := recover(); err != nil {
		var msg string
		if doLogStack {
			msg = fmt.Sprintf("caught panic: %v – %s", err, debug.Stack())
		} else {
			msg = fmt.Sprintf("caught panic: %v", err)
		}
		apachelogger.Err("TPageHandler.ServeHTTP()", msg)
	}
} // recoverPanic()

var (
	// RegEx to find path and possible added path components
	phURLpartsRE = regexp.MustCompile(
		`(?i)^/*([\p{L}\d_.-]+)?/*([\p{L}\d_§.?!=:;/,@# -]*)?`)
	//           1111111111111     222222222222222222222222
)

// URLparts returns two parts: `rDir` holds the base-directory of `aURL`,
// `rPath` holds the remaining part of `aURL`.
//
// Depending on the actual value of `aURL` both return values may be
// empty or both may be filled; none of both will hold a leading slash.
func URLparts(aURL string) (rDir, rPath string) {
	if result, err := url.QueryUnescape(aURL); nil == err {
		aURL = result
	}
	matches := phURLpartsRE.FindStringSubmatch(aURL)
	if 2 < len(matches) {
		return matches[1], strings.TrimSpace(matches[2])
	}

	return aURL, ""
} // URLparts()

/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */

// GetErrorPage returns an error page for `aStatus`,
// implementing the `TErrorPager` interface.
func (ph *TPageHandler) GetErrorPage(aData []byte, aStatus int) []byte {
	var empty []byte
	qo := db.NewQueryOptions(ph.docsPerPage)
	pageData := ph.basicTemplateData(qo).
		Set("ShowForm", false)

	switch aStatus {
	case 404:
		if page, err := ph.viewList.RenderedPage("404", pageData); nil == err {
			return page
		}

	default:
		pageData.Set("Error", template.HTML(aData)) // #nosec G203
		if page, err := ph.viewList.RenderedPage("error", pageData); nil == err {
			return page
		}
	}

	return empty
} // GetErrorPage()

/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */

// Address returns the configured `IP:Port` address to use for listening.
func (ph *TPageHandler) Address() string {
	return ph.addr
} // Address()

// `basicTemplateData()` returns a list of common template values.
func (ph *TPageHandler) basicTemplateData(aOptions *db.TQueryOptions) *TemplateData {
	y, m, d := time.Now().Date()

	lang := ph.lang
	switch aOptions.GuiLang {
	case db.QoLangEnglish:
		lang = "en"
	case db.QoLangGerman:
		lang = "de"
	}

	theme := ph.theme
	switch aOptions.Theme {
	case db.QoThemeDark:
		theme = "dark"
	case db.QoThemeLight:
		theme = "light"
	}
	return NewTemplateData().
		Set("CSS", template.HTML(`<link rel="stylesheet" type="text/css" title="mwat's styles" href="/css/stylesheet.css"><link rel="stylesheet" type="text/css" href="/css/`+theme+`.css"><link rel="stylesheet" type="text/css" href="/css/fonts.css">`)).
		Set("GUILANG", aOptions.SelectLanguageOptions()).
		Set("HasLast", false).
		Set("HasNext", false).
		Set("HasPrev", false).
		Set("IsGrid", db.QoLayoutGrid == aOptions.Layout).
		Set("Lang", lang).
		Set("LibraryName", ph.libName).
		Set("Robots", "noindex,nofollow").
		Set("SLO", aOptions.SelectLayoutOptions()).
		Set("SLL", aOptions.SelectLimitOptions()).
		Set("SOO", aOptions.SelectOrderOptions()).
		Set("SSB", aOptions.SelectSortByOptions()).
		Set("THEME", aOptions.SelectThemeOptions()).
		Set("Title", ph.realm+fmt.Sprintf(": %d-%02d-%02d", y, m, d)).
		Set("VirtLib", aOptions.SelectVirtLibOptions()) // #nosec G203
} // basicTemplateData()

/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */

// `handleGET()` processes the HTTP GET requests.
func (ph *TPageHandler) handleGET(aWriter http.ResponseWriter, aRequest *http.Request) {
	qo := db.NewQueryOptions(ph.docsPerPage) // in `queryoptions.go`
	so := sessions.GetSession(aRequest)
	if qos, ok := so.GetString("QOS"); ok {
		qo.Scan(qos)
	}

	dbHandle, err := db.OpenDatabase(aRequest.Context())
	if nil != err {
		msg := fmt.Sprintf("db.OpenDatabase(): %v", err)
		apachelogger.Err("TPageHandler.handleGET()", msg)
		return
	}
	defer dbHandle.Close()

	path, tail := URLparts(aRequest.URL.Path)
	switch path {
	case "authors", "format", "languages", "publisher", "series", "tags":
		parts := strings.Split(tail, `/`)
		qo.Entity = path
		qo.ID, _ = strconv.Atoi(parts[0])
		qo.LimitStart = 0 // it's the first page of a new selection
		if 0 < qo.ID {
			qo.Matching = path + `:"=` + parts[1] + `"`
		}
		ph.handleQuery(aWriter, aRequest, qo, so, dbHandle)

	case "back":
		qo.DecLimit()
		ph.handleQuery(aWriter, aRequest, qo, so, dbHandle)

	case "certs": // these files are handled internally
		http.Redirect(aWriter, aRequest, "/", http.StatusMovedPermanently)

	case "cover":
		var (
			id    db.TID
			dummy string
		)
		_, _ = fmt.Sscanf(tail, "%d/%s", &id, &dummy)
		doc := dbHandle.QueryDocMini(aRequest.Context(), id)
		if nil == doc {
			http.NotFound(aWriter, aRequest)
			return
		}
		file, err := doc.CoverAbs(true)
		if (nil != err) || (0 >= len(file)) {
			http.NotFound(aWriter, aRequest)
			return
		}
		aRequest.URL.Path = file
		ph.docFS.ServeHTTP(aWriter, aRequest)

	case "css":
		ph.cssFS.ServeHTTP(aWriter, aRequest)

	case "doc":
		var (
			id    db.TID
			dummy string
		)
		_, _ = fmt.Sscanf(tail, "%d/%s", &id, &dummy)
		qo.ID = id
		doc := dbHandle.QueryDocument(aRequest.Context(), id)
		if nil == doc {
			http.NotFound(aWriter, aRequest)
			return
		}
		pageData := ph.basicTemplateData(qo).
			Set("Document", doc)
		ph.handleReply("document", aWriter, so, qo, pageData)

	case "faq":
		ph.handleReply("faq", aWriter, so, qo, ph.basicTemplateData(qo))

	case "favicon.ico":
		http.Redirect(aWriter, aRequest, "/img/"+path, http.StatusMovedPermanently)

	case "file":
		parts := strings.Split(tail, `/`)
		qo.ID, _ = strconv.Atoi(parts[0])
		doc := dbHandle.QueryDocMini(aRequest.Context(), qo.ID)
		if nil == doc {
			http.NotFound(aWriter, aRequest)
			return
		}
		file := doc.Filename(parts[1])
		if 0 == len(file) {
			http.NotFound(aWriter, aRequest)
			return
		}
		aRequest.URL.Path = file
		ph.docFS.ServeHTTP(aWriter, aRequest)

	case "first", "":
		qo.LimitStart = 0
		ph.handleQuery(aWriter, aRequest, qo, so, dbHandle)

	case "fonts":
		ph.staticFS.ServeHTTP(aWriter, aRequest)

	case "help", "hilfe":
		ph.handleReply("help", aWriter, so, qo, ph.basicTemplateData(qo))

	case "img":
		ph.staticFS.ServeHTTP(aWriter, aRequest)

	case "imprint", "impressum":
		ph.handleReply("imprint", aWriter, so, qo, ph.basicTemplateData(qo))

	case "last":
		if qo.QueryCount <= qo.LimitLength {
			qo.LimitStart = 0
		} else {
			qo.LimitStart = qo.QueryCount - qo.LimitLength
		}
		ph.handleQuery(aWriter, aRequest, qo, so, dbHandle)

	case "licence", "license", "lizenz":
		ph.handleReply("licence", aWriter, so, qo, ph.basicTemplateData(qo))

	case "next":
		ph.handleQuery(aWriter, aRequest, qo, so, dbHandle)

	case "post":
		ph.handleQuery(aWriter, aRequest, qo, so, dbHandle)

	case "prev":
		// Since the current LimitStart points to the _next_ query
		// start we have to decrement the value twice to go _before_.
		qo.DecLimit().DecLimit()
		ph.handleQuery(aWriter, aRequest, qo, so, dbHandle)

	case "privacy", "datenschutz":
		ph.handleReply("privacy", aWriter, so, qo, ph.basicTemplateData(qo))

	case "qo":
		// This gets called when user requests page source of
		// a POST result page; try to handle it gracefully.
		ph.handleQuery(aWriter, aRequest, qo, so, dbHandle)

	case "robots.txt":
		ph.staticFS.ServeHTTP(aWriter, aRequest)

	case "sessions": // files are handled internally
		http.Redirect(aWriter, aRequest, "/", http.StatusMovedPermanently)

	case "thumb":
		var (
			id    db.TID
			dummy string
		)
		_, _ = fmt.Sscanf(tail, "%d/%s", &id, &dummy)
		doc := dbHandle.QueryDocMini(aRequest.Context(), id)
		if nil == doc {
			http.NotFound(aWriter, aRequest)
			return
		}
		tName, err := Thumbnail(doc)
		if nil != err {
			http.NotFound(aWriter, aRequest)
			return
		}
		file, err := filepath.Rel(db.CalibreCachePath(), tName)
		if nil != err {
			http.NotFound(aWriter, aRequest)
			return
		}
		aRequest.URL.Path = file
		ph.cacheFS.ServeHTTP(aWriter, aRequest)

	case "views": // files are handled internally
		http.Redirect(aWriter, aRequest, "/", http.StatusMovedPermanently)

	default:
		// // if nothing matched (above) reply to the request
		// // with an HTTP 404 not found error.
		//http.NotFound(aWriter, aRequest)

		// Redirect all unknown/invalid URLs to the NSA:
		http.Redirect(aWriter, aRequest, "https://www.nsa.gov/", http.StatusMovedPermanently)
	} // switch
} // handleGET()

// `handlePOST()` process the HTTP POST requests.
func (ph *TPageHandler) handlePOST(aWriter http.ResponseWriter, aRequest *http.Request) {
	dbHandle, err := db.OpenDatabase(aRequest.Context())
	if nil != err {
		msg := fmt.Sprintf("db.OpenDatabase(): %v", err)
		apachelogger.Err("TPageHandler.handlePOST()", msg)
		return
	}
	defer dbHandle.Close()

	path, _ := URLparts(aRequest.URL.Path)
	switch path {
	case "qo": // the only valid POST destination
		qo := db.NewQueryOptions(ph.docsPerPage)
		so := sessions.GetSession(aRequest)
		if qos, ok := so.GetString("QOS"); ok {
			qo.Scan(qos)
		}
		qo.Update(aRequest)
		// Since the query options hold the LimitStart of the
		// _next_ query we have to go back here one page:
		qo.DecLimit()
		ph.handleQuery(aWriter, aRequest, qo, so, dbHandle)

	default:
		// // if nothing matched (above) reply to the request
		// // with an HTTP 404 "not found" error.
		//http.NotFound(aWriter, aRequest)

		// Redirect all invalid URLs to the NSA:
		http.Redirect(aWriter, aRequest, "https://www.nsa.gov/", http.StatusMovedPermanently)
	}
} // handlePOST()

// `handleQuery()` serves the logical web-root directory.
//
//	`aWriter`
//	`aRequest`
//	`aOption`
//	`aSession`
//	`aDB` The DB handle to access the `Calibre` database.
func (ph *TPageHandler) handleQuery(aWriter http.ResponseWriter, aRequest *http.Request, aOption *db.TQueryOptions, aSession *sessions.TSession, aDB *db.TDataBase) {
	var (
		count   int
		doclist *db.TDocList
		err     error
	)
	if 0 < len(aOption.Matching) {
		count, doclist, err = aDB.QuerySearch(aRequest.Context(), aOption)
	} else {
		count, doclist, err = aDB.QueryBy(aRequest.Context(), aOption)
	}
	if nil != err {
		msg := fmt.Sprintf("QueryBy/QuerySearch: %v", err)
		apachelogger.Err("TPageHandler.handleQuery()", msg)
	}
	if 0 < count {
		aOption.QueryCount = uint(count)
	} else {
		aOption.QueryCount = 0
	}
	BFirst := aOption.LimitStart + 1 // zero-based to one-based
	BCount := aOption.QueryCount
	BLast := aOption.LimitStart + aOption.LimitLength
	if BLast > BCount {
		BLast = BCount
	}
	hasFirst := 0 < aOption.LimitStart
	hasLast := BLast < BCount
	hasNext := BCount > BLast
	hasPrev := aOption.LimitStart >= aOption.LimitLength
	aOption.IncLimit()
	pageData := ph.basicTemplateData(aOption).
		Set("BFirst", BFirst).
		Set("BLast", BLast).
		Set("BCount", BCount).
		Set("Documents", doclist).
		Set("HasFirst", hasFirst).
		Set("HasLast", hasLast).
		Set("HasNext", hasNext).
		Set("HasPrev", hasPrev).
		Set("Matching", aOption.Matching).
		Set("SID", aSession.ID()).
		Set("SIDNAME", sessions.SIDname()).
		Set("ShowForm", true)
	ph.handleReply("index", aWriter, aSession, aOption, pageData)
} // handleQuery()

// `handleReply()` sends the resulting page back to the remote user.
func (ph *TPageHandler) handleReply(aPage string, aWriter http.ResponseWriter, aSession *sessions.TSession, aOption *db.TQueryOptions, pageData *TemplateData) {
	aSession.Set("QOS", aOption.String())
	if err := ph.viewList.Render(aPage, aWriter, pageData); nil != err {
		msg := fmt.Sprintf("viewList.Render(%s): %v", aPage, err)
		apachelogger.Err("TPageHandler.handleReply()", msg)
	}
} // handleReply

// NeedAuthentication returns `true` if authentication is needed,
// or `false` otherwise.
//
//	`aRequest` is the web request to check.
func (ph *TPageHandler) NeedAuthentication(aRequest *http.Request) bool {
	if nil == ph.usrList {
		return false
	}
	if ph.authAll {
		return true
	}
	path, _ := URLparts(aRequest.URL.Path)
	// switch path {
	// case "file":
	// 	return true
	// default:
	// 	return false
	// }
	return (`file` == path)
} // NeedAuthentication()

// ServeHTTP handles the incoming HTTP requests.
func (ph *TPageHandler) ServeHTTP(aWriter http.ResponseWriter, aRequest *http.Request) {
	defer recoverPanic(ph.logStack)

	aWriter.Header().Set("Access-Control-Allow-Methods", "POST, GET")
	if ph.NeedAuthentication(aRequest) {
		if err := ph.usrList.IsAuthenticated(aRequest); nil != err {
			passlist.Deny(ph.realm, aWriter)
			return
		}
	}

	switch aRequest.Method {
	case "GET":
		ph.handleGET(aWriter, aRequest)

	case "POST":
		ph.handlePOST(aWriter, aRequest)

	default:
		msg := fmt.Sprintf("unsupported request method: %v", aRequest.Method)
		apachelogger.Err("TPageHandler.ServeHTTP()", msg)

		http.Error(aWriter, msg, http.StatusMethodNotAllowed)
	}
} // ServeHTTP()

/* _EoF_ */
