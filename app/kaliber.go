/*
   Copyright © 2019 M.Watermann, 10247 Berlin, Germany
                  All rights reserved
               EMail : <support@mwat.de>
*/

package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/mwat56/apachelogger"
	"github.com/mwat56/errorhandler"
	"github.com/mwat56/kaliber"
)

// `setupSinals()` configures the capture of the interrupts `SIGINT`,
// `SIGKILL`, and `SIGTERM` to terminate the program gracefully.
func setupSinals(aServer *http.Server) {
	// handle `CTRL-C` and `kill(9)` and `kill(15)`.
	c := make(chan os.Signal, 3)
	signal.Notify(c, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)

	go func() {
		for signal := range c {
			log.Printf("%s captured '%v', stopping program and exiting ...", os.Args[0], signal)
			if err := aServer.Shutdown(context.Background()); nil != err {
				log.Fatalf("%s: %v", os.Args[0], err)
			}
		}
	}()
} // setupSinals()

// Actually run the program …
func main() {
	var (
		err       error
		handler   http.Handler
		ph        *kaliber.TPageHandler
		ck, cp, s string
	)
	Me, _ := filepath.Abs(os.Args[0])

	if s, err = kaliber.AppArguments.Get("uf"); (nil == err) && (0 < len(s)) {
		fn := s
		if s, err = kaliber.AppArguments.Get("ua"); (nil == err) && (0 < len(s)) {
			kaliber.AddUser(s, fn)
		}
		if s, err = kaliber.AppArguments.Get("uc"); (nil == err) && (0 < len(s)) {
			kaliber.CheckUser(s, fn)
		}
		if s, err = kaliber.AppArguments.Get("ud"); (nil == err) && (0 < len(s)) {
			kaliber.DeleteUser(s, fn)
		}
		if s, err = kaliber.AppArguments.Get("ul"); (nil == err) && (0 < len(s)) {
			kaliber.ListUser(fn)
		}
		if s, err = kaliber.AppArguments.Get("uu"); (nil == err) && (0 < len(s)) {
			kaliber.UpdateUser(s, fn)
		}
	}

	if ph, err = kaliber.NewPageHandler(); nil != err {
		kaliber.ShowHelp()
		log.Fatalf("%s: %v", Me, err)
	}
	handler = errorhandler.Wrap(ph, ph)

	// inspect `logfile` commandline argument and setup the `ApacheLogger`
	if s, err = kaliber.AppArguments.Get("logfile"); (nil == err) && (0 < len(s)) {
		// we assume, an error means: no logfile
		handler = apachelogger.Wrap(handler, s)
	}
	// We need a `server` reference to use it in setupSinals() below
	server := &http.Server{Addr: ph.Address(), Handler: handler}
	setupSinals(server)

	ck, _ = kaliber.AppArguments.Get("certKey")
	cp, _ = kaliber.AppArguments.Get("certPem")

	if 0 < len(ck) && (0 < len(cp)) {
		log.Printf("%s listening HTTPS at: %s", Me, ph.Address())
		if err = server.ListenAndServeTLS(cp, ck); nil != err {
			log.Fatalf("%s: %v", Me, err)
		}
		return
	}

	log.Printf("%s listening HTTP at: %s", Me, ph.Address())
	if err = server.ListenAndServe(); nil != err {
		log.Fatalf("%s: %v", Me, err)
	}
} // main()

/* _EoF_ */
