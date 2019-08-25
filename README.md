# Kaliber

[![Golang](https://img.shields.io/badge/Language-Go-green.svg)](https://golang.org)
[![GoDoc](https://godoc.org/github.com/mwat56/kaliber?status.svg)](https://godoc.org/github.com/mwat56/kaliber/)
[![Go Report](https://goreportcard.com/badge/github.com/mwat56/kaliber)](https://goreportcard.com/report/github.com/mwat56/kaliber)
[![Issues](https://img.shields.io/github/issues/mwat56/kaliber.svg)](https://github.com/mwat56/kaliber/issues?q=is%3Aopen+is%3Aissue)
[![Size](https://img.shields.io/github/repo-size/mwat56/kaliber.svg)](https://github.com/mwat56/kaliber/)
[![Tag](https://img.shields.io/github/tag/mwat56/kaliber.svg)](https://github.com/mwat56/kaliber/releases)
[![License](https://img.shields.io/github/license/mwat56/kaliber.svg)](https://github.com/mwat56/kaliber/blob/master/LICENSE)
[![View examples](https://img.shields.io/badge/learn%20by-examples-0077b3.svg)](https://github.com/mwat56/kaliber/blob/master/app/kaliber.go)

- [Kaliber](#kaliber)
	- [Note](#note)
	- [Purpose](#purpose)
	- [Features](#features)
	- [Installation](#installation)
	- [Usage](#usage)
		- [Commandline options](#commandline-options)
		- [User/password file & handling](#userpassword-file--handling)
			- [Authentication](#authentication)
		- [INI file](#ini-file)
	- [Directory structure](#directory-structure)
	- [Caveats](#caveats)
	- [Logging](#logging)
	- [Libraries](#libraries)
	- [Licence](#licence)

## Note

> **Please _note_ that this is a work in progress.**
> Changes – even API breaking changes – can happen at any time.

## Purpose

I _love_ books.
Always have.
Since I was a kid.
Over the years several thousand books gathered in my flat; so many that I sort of ran out of space.
Relunctantly I started to investigate and then use eBooks, first on my main desktop computer, later with a dedicated eBook reader.

Soon – and again – there were so many that I started looking for some convenient way to handle, store, and retrieve them for reading.
That's when I became acquainted with [Calibre](https://calibre-ebook.com/), a great software for working with eBooks.
Of course, there were some problems but since the software is actively maintained and extended all the time those problems either went away on their own with the next update or I found some way around them.
And sure, it took some work to get all the eBooks into that library system, and since there are always coming new titles the work kind of never ends.

Another question soon became urgent: How to access my books when I'm not at home?
As it turned out, `Calibre` comes bundled with its own web-server.
After figuring out how to start it automatically when the machine gets restarted (which happens once in a while when some kernel software upgrade requires it) the server did its job.
Quite another question, however, is _how_ it does its job.
And there the problem lies.
The web-server coming with `Calibre` serves pages that are heavily dependent on JavaScript; so much so that the pages simply don't appear or work at all if you have JavaScript disabled in your browser e.g. for privacy or security reasons.
For a while I grudgingly activated JavaScript whenever I wanted to access my books remotely.
I asked the author of `Calibre` whether he'd be willing to provide a barrier-free alternative (i.e. without JavaScript) but unfortunately he declined: "Not going to happen."

Well, due to other projects I didn't find the time then to do something about it, and it took some more years before I started seriously to look for alternatives.
Therefore now there's `Kaliber`, a barrier-free web-server for your `Calibre` book collection.
It doesn't depend on – or even need – JavaScript, but just requires a web-browser capable to read plain HTML on the remote user's side.
So even privacy or security conscious people or those who depend on assisting technologies (like e.g. screen-readers) can now access their library.

As an aside:
> I never really understood the desire to multiply the work needed to be done for a web page.
>
> A _server_ is a _server_ which means it should _serve_.
> When in a restaurant ordering a meal you surely expect it to be brought to you fully prepared and ready to be consumed.
You'd probably were seriously annoyed if the waiter/server just brought you the ingredients and left it to you to prepare the meal for yourself.
>
> For some strange reason, however, that's exactly what happens on a growing number of web-presentations: Instead of delivering a ready-to-read web-page they send every user just program code and let the user's machine do the preparing of the page.
> In other words: Instead of one server doing the work and delivering it to thousands of remote users nowadays the server forces thousands (or even millions) of remote users to spend time and electricity to just see a single web-page.

Since I couldn't find a documentation of the database structure and/or API used by `Calibre` to store its meta-data I had to reverse engineer ways to access the stored book data.
The same is true – to a certain lesser degree – for the web-pages served by `Kaliber`: it's kind of a mix of `Calibre`'s normal (i.e. JavaScript based) and `mobile` pages.
The overall layout of the web-pages served by `Kaliber` is intentionally kept simple ([KISS](https://en.wikipedia.org/wiki/KISS_principle)).

## Features

* Simplicity of use;
* Barrier-free (no JavaScript required);
* Books list layout either `Cover Grid` or `Data List`;
* Easy navigation (`First`, `Prev`, `Next`, `Last` button/links);
* Fulltext search as well as datafield-based searches;
* Ordered in either `ascending` or `descending` direction;
* Selectable number of books per page;
* Sortable by `author`, `date`, `language`, `publisher`, `rating`, `series`, `size`, `tags`, or `title`;
* Anonymised access logging;
* Optional user/password based access control.

## Installation

You can use `Go` to install this package for you:

    go get -u github.com/mwat56/kaliber

	//TODO

## Usage

After downloading this package you go to its directory and compile

    go build app/kaliber.go

which should produce an executable binary.

### Commandline options

	$ ./kaliber -h
	Usage: ./kaliber [OPTIONS]

	-booksperpage int
		<number> the default number of books shown per page  (default 25)
	-certKey string
		<fileName> the name of the TLS certificate key
		(default "/home/matthias/devel/Go/src/github.com/mwat56/kaliber/certs/server.key")
	-certPem string
		<fileName> the name of the TLS certificate PEM
		(default "/home/matthias/devel/Go/src/github.com/mwat56/kaliber/certs/server.pem")
	-datadir string
		<dirName> the directory with CACHE, CSS, IMG, and VIEWS sub-directories
		(default "/home/matthias/devel/Go/src/github.com/mwat56/kaliber")
	-ini string
		<fileName> the path/filename of the INI file to use
		(default "/home/matthias/devel/Go/src/github.com/mwat56/kaliber/kaliber.ini")
	-lang string
		(optional) the default language to use  (default "de")
	-libraryname string
		Name of this Library (shown on every page)
		(default "MyBooks")
	-librarypath string
		Path name of/to the Calibre library
		(default "/var/opt/Calibre")
	-listen string
		the host's IP to listen at (default "127.0.0.1")
	-log string
		(optional) name of the logfile to write to
		(default "/dev/stdout")
	-port int
		<portNumber> the IP port to listen to  (default 8383)
	-realm string
		(optional) <hostName> name of host/domain to secure by BasicAuth
		(default "This Host")
	-sessiondir string
		<directory> (optional) the directory to store session files
		(default "/home/matthias/devel/Go/src/github.com/mwat56/kaliber/sessions")
	-sessionttl int
		<seconds> Number of seconds an unused session keeps valid (default 1200)
	-sidname string
		(optional) <name> the name of the session ID to use
		(default "sid")
	-theme string
		<name> the display theme to use ('light' or 'dark')
		(default "light")
	-ua string
		<userName> (optional) user add: add a username to the password file
	-uc string
		<userName> (optional) user check: check a username in the password file
	-ud string
		<userName> (optional) user delete: remove a username from the password file
	-uf string
		<fileName> (optional) user passwords file storing user/passwords for BasicAuth
		(default "/home/matthias/devel/Go/src/github.com/mwat56/kaliber/pwaccess.db")
	-ul
		(optional) user list: show all users in the password file
	-uu string
		<userName> (optional) user update: update a username in the password file

	Most options can be set in an INI file to keep the command-line short ;-)
	$ _

As you can see there are quite a few options available, but almost all of them are optional since they come with reasonable default values.

	//TODO

### User/password file & handling

Only usable from the commandline are the `-uXX` options, most of which need an username and the name of the password file to use.
_Note_ that whenever you're prompted to input a password this will _not_ be echoed to the console.

    $ ./kaliber -ua testuser1 -uf pwaccess.db

     password:
    repeat pw:
        added 'testuser1' to list
    $ _

The password input is not echoed to the console, therefor you don't see it.

Since we have the `passfile` setting already in our INI file [(see below)](#ini-file) we can forget the `-uf` option for the next options.

With `-uc` you can check a user's password:

    $ ./kaliber -uc testuser1

     password:
        'testuser1' password check successful
    $ _

This `-uc` you'll probably never actually use, it was just easy to implement.

If you want to remove a user the `-ud` will do the trick:

    $ ./kaliber -ud testuser1
        removed 'testuser1' from list
    $ _

When you want to know which users are stored in your password file `-ul` is your friend:

    $ ./kaliber -ul
    matthias

    $ _

Since we deleted the `testuser1` before only one entry remains.

That only leaves `-uu` to update (change) a user's password.

    $ ./kaliber -ua testuser2

     password:
    repeat pw:
        added 'testuser2' to list

    $ ./kaliber -uu testuser2

     password:
    repeat pw:
        updated user 'testuser2' in list

    $ ./kaliber -ul
    matthias
    testuser2

    $ _

First we added (`-ua`) a new user, then we updated the password (`-uu`), and finally we asked for the list of users (`-ul`).

#### Authentication

But why, you may ask, would you need an username/password file anyway?
Well, there may be several reasons one of which could be Copyright problems.
If not all your books are in the public domain and Copyright-free in most countries you may _not make them publically_ available.
In that case you're most likely the only actual remote user allowed to access the books in your library.

Whenever there's no password file given (either in the INI file or the command-line) all functionality requiring authentication will be _disabled_ which in turn means that everybody can access your library.
Depending on your country's legislation you may or may not include your family members.
If in doubt please consult a Copyright expert.

_Note_ that the password file generated and used by this system resembles the `htpasswd` used by the _Apache_ web-server, but both files are _not_ interchangeable because the actual encryption algorithms used by both are different.

### INI file

But relax!
You don't have to give all those commandline options listed above every time you want to start `Kaliber`.
There's an INI file which can take all the options (apart from the user handling options) all in one place:

	$ cat kaliber.ini
	# Default configuration file for the Kaliber server

	[Default]

	# Number of books to show per page
	booksperpage = 25

	# path-/filename of TLS certificate's private key to enable TLS/HTTPS
	# (if empty standard HTTP is used)
	# NOTE: a relative path/name will be combined with `datadir` (below).
	certKey = ./certs/server.key

	# path-/filename of TLS (server) certificate to enable TLS/HTTPS
	# (if empty standard HTTP is used)
	# NOTE: a relative path/name will be combined with `dadadir` (below).
	certPem = ./certs/server.pem

	# The directory root for CSS, FONTS, IMG, and VIEWS sub-directories.
	# NOTE: this should be an absolute path name.
	datadir = ./

	# The default language to use:
	lang = de

	# Name of this library (shown on every page)
	libraryname = "MyBooks"

	# Path of Calibre library
	# NOTE: this must be the absolute pathname of the Calibre library.
	librarypath = "/var/opt/Calibre"

	# The host's IP to listen at:
	listen = 127.0.0.1

	# The IP port to listen to:
	port = 8383

	# Name of the optional logfile to write to.
	# NOTE: a relative path/name will be combined with `datadir` (above).
	logfile = /dev/stdout

	# Password file for HTTP Basic Authentication.
	# NOTE: a relative path/name will be combined with `datadir` (above).
	passfile = ./pwaccess.db

	# Name of host/domain to secure by BasicAuth:
	realm = "This Host"

	# Name of the directory to store session files.
	# NOTE: a relative path/name will be combined with `datadir` (above).
	sessiondir = "./sessions"

	# Number of seconds an unused session keeps valid:
	sessionTTL = 1200

	# Web/display theme: `dark` or `light':
	theme = light

	# _EoF_
	$ _


	//TODO

## Directory structure

Under the directory given with the `datadir =` entry in the INI file or the `-datadir` commandline option there are several sub-directories expected:

* `css`: containing the CSS files used,
* `fonts`: containing the fonts used,
* `img`: containing the images used,
* `views`: the Go templates used to generate the pages.

All of this directories and files are part of the `Kaliber` package.
You can use them as is or customise them as you see fit to suit your needs.
However, please note: _I will not support any customisations_, you're on your own with that.

## Caveats

There are some `Calibre` features which are not available (yet) with `Kaliber` and not currently supported:

* _custom columns_ defined by the respective `Calibre` user;
* _different/multiple libraries_ of the user to switch between;
* _book uploads_ are not planned to be included.

Once I figure out how they are realised by `Calibre` I expect they find their way into `Kaliber` as well (provided I find actually time to do it).

## Logging

Like almost every other web-server `Kaliber` writes all access data to a logfile (`logfile =` in the INI file and `-log` at the commandline).
As _**privacy**_ becomes a serious concern for a growing number of people (including law makers) – the IP address is definitely to be considered as _personal data_ – the [logging facility](https://github.com/mwat56/apachelogger) _anonymises_ the requesting users by setting the host-part of the respective remote address to zero (`0`).
This option takes care of e.g. European servers who may _not without explicit consent_ of the users store personal data; this includes IP addresses in logfiles and elsewhere (eg. statistical data gathered from logfiles).
Since the generated logfile resembles that of the popular `Apache` server you can use all tools written for `Apache` logfiles to anylyse the access data.

## Libraries

The following external libraries were used building `Kaliber`:

* [ApacheLogger](https://github.com/mwat56/apachelogger)
* [Crypto](https://golang.org/x/crypto)
* [ErrorHandler](https://github.com/mwat56/errorhandler)
* [INI](https://github.com/mwat56/ini)
* [PassList](https://github.com/mwat56/passlist)
* [Resize](https://github.com/nfnt/resize)
* [Sessions](https://github.com/mwat56/sessions)
* [SQLite3](https://github.com/mattn/go-sqlite3)

## Licence

        Copyright © 2019 M.Watermann, 10247 Berlin, Germany
                        All rights reserved
                    EMail : <support@mwat.de>

> This program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation; either version 3 of the License, or (at your option) any later version.
>
> This software is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
>
> You should have received a copy of the GNU General Public License along with this program. If not, see the [GNU General Public License](http://www.gnu.org/licenses/gpl.html) for details.
