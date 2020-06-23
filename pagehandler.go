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
		cacheFS  http.Handler        // cache file server (i.e. thumbnails)
		cssFS    http.Handler        // CSS file server
		docFS    http.Handler        // document file server
		staticFS http.Handler        // static file server
		usrList  *passlist.TPassList // user/password list
		viewList *TViewList          // list of template/views
	}
)

// `handleInternalError()` sends an error page to the remote browser.
//
//	`aWriter` Used by the HTTP handler to construct an HTTP response.
//	`aSender`
//	`aMessage`
func handleInternalError(aWriter http.ResponseWriter,
	aSender, aMessage string) {
	apachelogger.Err(aSender, aMessage)
	http.Error(aWriter, `unknown page arguments`,
		http.StatusInternalServerError)
} // handleInternalError()

// NewPageHandler returns a new `TPageHandler` instance.
func NewPageHandler() (*TPageHandler, error) {
	var (
		err error
	)
	result := new(TPageHandler)

	result.cacheFS = jffs.FileServer(db.CalibreCachePath())
	result.cssFS = cssfs.FileServer(AppArgs.DataDir + `/`)
	result.docFS = jffs.FileServer(db.CalibreLibraryPath())
	result.staticFS = jffs.FileServer(AppArgs.DataDir)

	if s := AppArgs.PassFile; 0 == len(s) {
		s = "missing user/password file\nAUTHENTICATION DISABLED!`"
		apachelogger.Err("NewPageHandler()", s)
	} else if result.usrList, err = passlist.LoadPasswords(s); nil != err {
		s = fmt.Sprintf("%v\nAUTHENTICATION DISABLED!", err)
		apachelogger.Err("NewPageHandler()", s)
		result.usrList = nil
	}

	if result.viewList, err = newViewList(filepath.Join(AppArgs.DataDir, `views`)); nil != err {
		return nil, err
	}

	// Update the thumbnails cache:
	go ThumbnailUpdate()

	// Avoid sessions for certain requests:
	sessions.ExcludePaths("/certs", "/css/", "/favicon", "/file/", "/fonts", "/img/", "/robots")

	return result, nil
} // NewPageHandler()

// `newViewList()` returns a list of views found in `aDirectory`
// and a possible I/O error.
//
//	`aDirectory` Where to look for template files.
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

var (
	// RegEx to find path and possible added path components
	phURLpartsRE = regexp.MustCompile(
		`(?i)^/*([\p{L}\d_.-]+)?/*([\p{L}\d_§.?!=:;/,@# -]*)?`)
	//           1111111111111     222222222222222222222222
)

// URLparts returns two parts: `rDir` holds the base-directory of
// `aURL`, `rPath` holds the remaining part of `aURL`.
//
// Depending on the actual value of `aURL` both return values may be
// empty or both may be filled; none of both will hold a leading slash.
//
//	`aURL` The address to split up.
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

// `basicTemplateData()` returns a list of common template values.
//
//	`aRequest` The HTTP request received by the server.
//	`aOptions` The current query options to use.
func (ph *TPageHandler) basicTemplateData(aRequest *http.Request, aOptions *db.TQueryOptions) *TemplateData {
	y, m, d := time.Now().Date()

	lang := AppArgs.Lang
	switch aOptions.GuiLang {
	case db.QoLangEnglish:
		lang = `en`
	default: // case db.QoLangGerman:
		lang = `de`
	}

	theme := AppArgs.Theme
	switch aOptions.Theme {
	case db.QoThemeDark:
		theme = `dark`
	default: // case db.QoThemeLight:
		theme = `light`
	}

	if nil != aRequest {
		if l := strings.ToLower(aRequest.FormValue(`lang`)); 0 < len(l) {
			switch l {
			case `de`:
				lang = l
				aOptions.GuiLang = db.QoLangGerman
			case `en`:
				lang = l
				aOptions.GuiLang = db.QoLangEnglish
			}
		}
		if t := strings.ToLower(aRequest.FormValue(`theme`)); 0 < len(t) {
			switch t {
			case `dark`:
				theme = t
				aOptions.Theme = db.QoThemeDark
			case `light`:
				theme = t
				aOptions.Theme = db.QoThemeLight
			}
		}
	}

	return NewTemplateData().
		Set("CSS", template.HTML(`<link rel="stylesheet" type="text/css" title="mwat's styles" href="/css/stylesheet.css"><link rel="stylesheet" type="text/css" href="/css/`+theme+`.css"><link rel="stylesheet" type="text/css" href="/css/fonts.css">`)).
		Set("GUILANG", aOptions.SelectLanguageOptions()).
		Set("HasLast", false).
		Set("HasNext", false).
		Set("HasPrev", false).
		Set("IsGrid", db.QoLayoutGrid == aOptions.Layout).
		Set("Lang", lang).
		Set("LibraryName", AppArgs.LibName).
		Set("Robots", "noindex,nofollow").
		Set("SLO", aOptions.SelectLayoutOptions()).
		Set("SLL", aOptions.SelectLimitOptions()).
		Set("SOO", aOptions.SelectOrderOptions()).
		Set("SSB", aOptions.SelectSortByOptions()).
		Set("THEME", aOptions.SelectThemeOptions()).
		Set("Title", AppArgs.Realm+fmt.Sprintf(": %d-%02d-%02d", y, m, d)).
		Set("VirtLib", aOptions.SelectVirtLibOptions()) // #nosec G203
} // basicTemplateData()

// GetErrorPage returns an error page for `aStatus`,
// implementing the `TErrorPager` interface.
//
//	`aData` The original error text.
//	`aStatus` The number of the actual HTTP error status.
func (ph *TPageHandler) GetErrorPage(aData []byte, aStatus int) []byte {
	var empty []byte
	qo := db.NewQueryOptions(AppArgs.BooksPerPage)
	pageData := ph.basicTemplateData(nil, qo).
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

// `handleGET()` processes the HTTP GET requests.
//
//	`aWriter` Used by the HTTP handler to construct an HTTP response.
//	`aRequest` The HTTP request received by the server.
func (ph *TPageHandler) handleGET(aWriter http.ResponseWriter, aRequest *http.Request) {
	var (
		dbHandle *db.TDataBase
		dummy    string
		err      error
		id       db.TID
	)
	defer func() {
		if nil != dbHandle {
			defer dbHandle.Close()
		}
	}()

	path, tail := URLparts(aRequest.URL.Path)
	so := sessions.GetSession(aRequest)
	qo := db.NewQueryOptions(AppArgs.BooksPerPage) // in `queryoptions.go`
	if qos, ok := so.GetString("QOS"); ok {
		qo.Scan(qos)
	}

	doOpenDatabase := func() *db.TDataBase {
		if dbHandle, err = db.OpenDatabase(aRequest.Context()); nil != err {
			dbHandle = nil
			handleInternalError(aWriter,
				`TPageHandler.handleGET('`+path+`')`,
				fmt.Sprintf("db.OpenDatabase(): %v", err))
		}
		return dbHandle
	} // doOpenDatabase()

	doHandleQuery := func() {
		if nil != doOpenDatabase() {
			ph.handleQuery(aWriter, aRequest, qo, so, dbHandle)
		}
	} // doHandleQuery()

	switch path {
	case "authors", "format", "languages", "publisher", "series", "tags":
		parts := strings.Split(tail, `/`)
		qo.Entity = path
		qo.ID, _ = strconv.Atoi(parts[0])
		qo.LimitStart = 0 // it's the first page of a new selection
		if 0 < qo.ID {
			qo.Matching = path + `:"=` + parts[1] + `"`
		}
		doHandleQuery()

	case "back":
		qo.DecLimit()
		doHandleQuery()

	case "certs": // these files are handled internally
		http.Redirect(aWriter, aRequest, "/", http.StatusMovedPermanently)

	case `cover`:
		if nil == doOpenDatabase() {
			return
		}
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

	case `doc`:
		if nil == doOpenDatabase() {
			return
		}
		_, _ = fmt.Sscanf(tail, "%d/%s", &id, &dummy)
		qo.ID = id
		doc := dbHandle.QueryDocument(aRequest.Context(), id)
		if nil == doc {
			http.NotFound(aWriter, aRequest)
			return
		}
		pageData := ph.basicTemplateData(aRequest, qo).
			Set("Document", doc)
		ph.handleReply(`document`, aWriter, qo, so, pageData)

	case `faq`:
		ph.handleReply(`faq`, aWriter, qo, so, ph.basicTemplateData(aRequest, qo))

	case "favicon.ico":
		http.Redirect(aWriter, aRequest, "/img/"+path, http.StatusMovedPermanently)

	case `file`:
		if nil == doOpenDatabase() {
			return
		}
		parts := strings.Split(tail, `/`)
		id, _ = strconv.Atoi(parts[0])

		doc := dbHandle.QueryDocMini(aRequest.Context(), id)
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

	case `first`, ``:
		qo.LimitStart = 0
		if 0 == len(path) {
			path = `first`
		}
		doHandleQuery()

	case "fonts":
		ph.staticFS.ServeHTTP(aWriter, aRequest)

	case `help`, `hilfe`:
		ph.handleReply(`help`, aWriter, qo, so, ph.basicTemplateData(aRequest, qo))

	case "img":
		ph.staticFS.ServeHTTP(aWriter, aRequest)

	case `imprint`, `impressum`:
		ph.handleReply(`imprint`, aWriter, qo, so, ph.basicTemplateData(aRequest, qo))

	case "last":
		if qo.QueryCount <= qo.LimitLength {
			qo.LimitStart = 0
		} else {
			qo.LimitStart = qo.QueryCount - qo.LimitLength
		}
		doHandleQuery()

	case `licence`, `license`, `lizenz`:
		ph.handleReply(`licence`, aWriter, qo, so, ph.basicTemplateData(aRequest, qo))

	case `next`:
		doHandleQuery()

	case `post`:
		doHandleQuery()

	case "prev":
		// Since the current LimitStart points to the _next_ query
		// start we have to decrement the value twice to go _before_.
		qo.DecLimit().DecLimit()
		doHandleQuery()

	case `privacy`, `datenschutz`:
		ph.handleReply(`privacy`, aWriter, qo, so, ph.basicTemplateData(aRequest, qo))

	case "qo":
		// This gets called when user requests page source of
		// a POST result page; try to handle it gracefully.
		doHandleQuery()

	case "robots.txt":
		ph.staticFS.ServeHTTP(aWriter, aRequest)

	case "sessions": // files are handled internally
		http.Redirect(aWriter, aRequest, "/", http.StatusMovedPermanently)

	case `thumb`:
		if nil == doOpenDatabase() {
			return
		}
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
//
//	`aWriter` Used by the HTTP handler to construct an HTTP response.
//	`aRequest` The HTTP request received by the server.
func (ph *TPageHandler) handlePOST(aWriter http.ResponseWriter, aRequest *http.Request) {
	path, _ := URLparts(aRequest.URL.Path)
	switch path {
	case "qo": // the only valid POST destination
		qo := db.NewQueryOptions(AppArgs.BooksPerPage)
		so := sessions.GetSession(aRequest)
		if qos, ok := so.GetString("QOS"); ok {
			qo.Scan(qos)
		}
		qo.Update(aRequest)
		// Since the query options hold the LimitStart of the
		// _next_ query we have to go back here one page:
		qo.DecLimit()

		dbHandle, err := db.OpenDatabase(aRequest.Context())
		if nil != err {
			handleInternalError(aWriter,
				`TPageHandler.handlePOST('`+path+`')`,
				fmt.Sprintf("db.OpenDatabase(): %v", err))
			return
		}
		defer dbHandle.Close()

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
//	`aWriter` Used by the HTTP handler to construct an HTTP response.
//	`aRequest` The HTTP request received by the server.
//	`aOptions` The current query options to use.
//	`aSession` The current user session.
//	`aDB` The DB handle to access the `Calibre` database.
func (ph *TPageHandler) handleQuery(aWriter http.ResponseWriter, aRequest *http.Request, aOptions *db.TQueryOptions, aSession *sessions.TSession, aDB *db.TDataBase) {
	var (
		count   int
		doclist *db.TDocList
		err     error
	)
	if 0 < len(aOptions.Matching) {
		count, doclist, err = aDB.QuerySearch(aRequest.Context(), aOptions)
	} else {
		count, doclist, err = aDB.QueryBy(aRequest.Context(), aOptions)
	}
	if nil != err {
		msg := fmt.Sprintf("QueryBy/QuerySearch: %v", err)
		apachelogger.Err("TPageHandler.handleQuery()", msg)
	}
	if 0 < count {
		aOptions.QueryCount = uint(count)
	} else {
		aOptions.QueryCount = 0
	}
	BFirst := aOptions.LimitStart + 1 // zero-based to one-based
	BCount := aOptions.QueryCount
	BLast := aOptions.LimitStart + aOptions.LimitLength
	if BLast > BCount {
		BLast = BCount
	}
	hasFirst := 0 < aOptions.LimitStart
	hasLast := BLast < BCount
	hasNext := BCount > BLast
	hasPrev := aOptions.LimitStart >= aOptions.LimitLength
	aOptions.IncLimit()
	pageData := ph.basicTemplateData(aRequest, aOptions).
		Set("BFirst", BFirst).
		Set("BLast", BLast).
		Set("BCount", BCount).
		Set("Documents", doclist).
		Set("HasFirst", hasFirst).
		Set("HasLast", hasLast).
		Set("HasNext", hasNext).
		Set("HasPrev", hasPrev).
		Set("Matching", aOptions.Matching).
		Set("SID", aSession.ID()).
		Set("SIDNAME", sessions.SIDname()).
		Set("ShowForm", true)
	ph.handleReply("index", aWriter, aOptions, aSession, pageData)
} // handleQuery()

// `handleReply()` sends the resulting page back to the remote user.
//
//	`aPage` Name of the template/view to use.
//	`aWriter` Used by the HTTP handler to construct an HTTP response.
//	`aOptions` The current query options to use.
//	`aSession` The current user session.
//	`aPageData` List of current template values.
func (ph *TPageHandler) handleReply(aPage string, aWriter http.ResponseWriter, aOptions *db.TQueryOptions, aSession *sessions.TSession, aPageData *TemplateData) {
	// store query options in session data
	aSession.Set("QOS", aOptions.String())

	if err := ph.viewList.Render(aPage, aWriter, aPageData); nil != err {
		handleInternalError(aWriter, `TPageHandler.handleReply()`,
			fmt.Sprintf("viewList.Render(%s): %v", aPage, err))
	}
} // handleReply()

// NeedAuthentication returns `true` if authentication is needed,
// or `false` otherwise.
//
// This method implements the `passlist.TAuthDecider` interface.
//
//	`aRequest` The web request to check.
func (ph *TPageHandler) NeedAuthentication(aRequest *http.Request) bool {
	if nil == ph.usrList {
		return false
	}
	if AppArgs.AuthAll {
		return true
	}
	path, _ := URLparts(aRequest.URL.Path)
	return (`file` == path)
} // NeedAuthentication()

// ServeHTTP handles the incoming HTTP requests.
//
//	`aWriter` Used by the HTTP handler to construct an HTTP response.
//	`aRequest` The HTTP request received by the server.
func (ph *TPageHandler) ServeHTTP(aWriter http.ResponseWriter, aRequest *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			var msg string
			if AppArgs.LogStack {
				msg = fmt.Sprintf("caught panic: %v – %s", err, debug.Stack())
			} else {
				msg = fmt.Sprintf("caught panic: %v", err)
			}
			handleInternalError(aWriter, `TPageHandler.ServeHTTP()`, msg)
		}
	}()

	aWriter.Header().Set("Access-Control-Allow-Methods", "POST, GET")
	if ph.NeedAuthentication(aRequest) {
		if err := ph.usrList.IsAuthenticated(aRequest); nil != err {
			passlist.Deny(AppArgs.Realm, aWriter)
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
