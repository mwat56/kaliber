/*
   Copyright © 2019, 2020 M.Watermann, 10247 Berlin, Germany
                  All rights reserved
               EMail : <support@mwat.de>
*/

package main

//lint:file-ignore ST1017 - I prefer Yoda conditions

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"github.com/NYTimes/gziphandler"
	"github.com/mwat56/apachelogger"
	"github.com/mwat56/errorhandler"
	"github.com/mwat56/kaliber"
	"github.com/mwat56/sessions"
)

// `exit()` logs `aMessage` and terminates the program.
func exit(aMessage string) {
	apachelogger.Err("Kaliber/main", aMessage)
	runtime.Gosched() // let the logger write
	log.Fatalln(aMessage)
} // exit()

// `userCmdline()` checks for and executes user/password handling functions.
func userCmdline() {
	if 0 == len(kaliber.AppArgs.PassFile) {
		return // without user file nothing to do
	}

	// All the following `kaliber.UserXxx()` calls terminate the program:
	if 0 < len(kaliber.AppArgs.UserAdd) {
		kaliber.UserAdd(kaliber.AppArgs.UserAdd, kaliber.AppArgs.PassFile)
	}
	if 0 < len(kaliber.AppArgs.UserCheck) {
		kaliber.UserCheck(kaliber.AppArgs.UserCheck, kaliber.AppArgs.PassFile)
	}
	if 0 < len(kaliber.AppArgs.UserDelete) {
		kaliber.UserDelete(kaliber.AppArgs.UserDelete, kaliber.AppArgs.PassFile)
	}
	if kaliber.AppArgs.UserList {
		kaliber.ListUsers(kaliber.AppArgs.PassFile)
	}
	if 0 < len(kaliber.AppArgs.UserUpdate) {
		kaliber.UserUpdate(kaliber.AppArgs.UserUpdate, kaliber.AppArgs.PassFile)
	}
} // userCmdline()

// `setupSignals()` configures the capture of the interrupts `SIGINT`
// and `SIGTERM` to terminate the program gracefully.
//
//	`aServer` The server instance to shutdown if a signal arrives.
func setupSignals(aServer *http.Server) {
	// handle `CTRL-C` and `kill(15)`.
	c := make(chan os.Signal, 2)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for signal := range c {
			msg := fmt.Sprintf("%s captured '%v', stopping program and exiting ...", os.Args[0], signal)
			apachelogger.Err(`Kaliber/catchSignals`, msg)
			log.Println(msg)
			runtime.Gosched() // let the logger write
			if err := aServer.Shutdown(context.Background()); nil != err {
				exit(fmt.Sprintf("%s: %v", os.Args[0], err))
			}
		}
	}()
} // setupSignals()

func main() {
	var (
		err error
		ph  *kaliber.TPageHandler
	)
	Me, _ := filepath.Abs(os.Args[0])

	// Read INI file(s) and commandline options:
	kaliber.InitConfig()

	// Handle commandline user/password maintenance:
	userCmdline()

	if ph, err = kaliber.NewPageHandler(); nil != err {
		kaliber.ShowHelp()
		exit(fmt.Sprintf("%s: %v", Me, err))
	}
	// Setup the errorpage handler:
	handler := errorhandler.Wrap(ph, ph)

	// Inspect `sessiondir` config option and setup the session handler
	if 0 < len(kaliber.AppArgs.SessionDir) {
		// an empty string means: no automatic session handling
		handler = sessions.Wrap(handler, kaliber.AppArgs.SessionDir)
	}

	// Inspect `gzip` config option and setup the Gzip handler:
	if kaliber.AppArgs.GZip {
		// a FALSE means: no gzip compression
		handler = gziphandler.GzipHandler(handler)
	}

	// Use logging config options and setup the `ApacheLogger`:
	handler = apachelogger.Wrap(handler,
		kaliber.AppArgs.AccessLog, kaliber.AppArgs.ErrorLog)

	// We need a `server` reference to use it in `setupSignals()`
	// and to set some reasonable timeouts:
	server := &http.Server{
		Addr:              kaliber.AppArgs.Addr,
		Handler:           handler,
		IdleTimeout:       0,
		ReadHeaderTimeout: 20 * time.Second,
		ReadTimeout:       20 * time.Second,
		// enough time for book download with little bandwidth:
		WriteTimeout: 20 * time.Minute,
	}
	apachelogger.SetErrLog(server)
	setupSignals(server)

	if (0 < len(kaliber.AppArgs.CertKey)) && (0 < len(kaliber.AppArgs.CertPem)) {
		// see:
		// https://ssl-config.mozilla.org/#server=golang&version=1.14.1&config=old&guideline=5.4
		server.TLSConfig = &tls.Config{
			MinVersion:               tls.VersionTLS10,
			PreferServerCipherSuites: true,
			CipherSuites: []uint16{
				tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256,
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256,
				tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
				tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
				tls.TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA,
				tls.TLS_ECDHE_RSA_WITH_RC4_128_SHA,
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
				tls.TLS_ECDHE_ECDSA_WITH_RC4_128_SHA,
				tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_RSA_WITH_AES_128_CBC_SHA256,
				tls.TLS_RSA_WITH_AES_256_CBC_SHA,
				tls.TLS_RSA_WITH_AES_128_CBC_SHA,
				tls.TLS_RSA_WITH_3DES_EDE_CBC_SHA,
				tls.TLS_RSA_WITH_RC4_128_SHA,
				tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256, // #nosec G402
			},
		} // #nosec G402
		server.TLSNextProto = make(map[string]func(*http.Server, *tls.Conn, http.Handler))

		if s := fmt.Sprintf("%s listening HTTPS at %s", Me, server.Addr); 0 < len(s) {
			log.Println(s)
			apachelogger.Log("Kaliber/main", s)
		}
		exit(fmt.Sprintf("%s: %v", Me,
			server.ListenAndServeTLS(kaliber.AppArgs.CertPem, kaliber.AppArgs.CertKey)))
		return
	}

	if s := fmt.Sprintf("%s listening HTTP at %s", Me, server.Addr); 0 < len(s) {
		log.Println(s)
		apachelogger.Log("Kaliber/main", s)
	}
	exit(fmt.Sprintf("%s: %v", Me, server.ListenAndServe()))
} // main()

/* _EoF_ */
