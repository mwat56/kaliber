/*
   Copyright © 2019, 2020 M.Watermann, 10247 Berlin, Germany
                  All rights reserved
               EMail : <support@mwat.de>
*/

package kaliber

import "testing"

//lint:file-ignore ST1017 - I prefer Yoda conditions

func TestURLparts(t *testing.T) {
	type args struct {
		aURL string
	}
	tests := []struct {
		name      string
		args      args
		wantRHead string
		wantRTail string
	}{
		// TODO: Add test cases.
		{" 1", args{"/"}, "", ""},
		{" 1a", args{""}, "", ""},
		{" 1b", args{"index/ "}, "index", ""},
		{" 2", args{"/css"}, "css", ""},
		{" 2a", args{"css"}, "css", ""},
		{" 3", args{"/css/styles.css"}, "css", "styles.css"},
		{" 3a", args{"css/styles.css"}, "css", "styles.css"},
		{" 4", args{"/?q=searchterm"}, "", "?q=searchterm"},
		{" 4a", args{"?q=searchterm"}, "", "?q=searchterm"},
		{" 5", args{"/article/abcdef1122334455"},
			"article", "abcdef1122334455"},
		{" 6", args{"/q/searchterm"}, "q", "searchterm"},
		{" 6a", args{"/q/?s=earchterm"}, "q", "?s=earchterm"},
		{" 7", args{"/q/search?s=term"}, "q", "search?s=term"},
		{" 8", args{"/static/https://github.com/"}, "static", "https://github.com/"},
		{" 9", args{"/ht/kurzerklärt"}, "ht", "kurzerklärt"},
		{"10", args{`share/https://utopia.de/ratgeber/pink-lady-das-ist-faul-an-dieser-apfelsorte/#main_content`}, `share`, `https://utopia.de/ratgeber/pink-lady-das-ist-faul-an-dieser-apfelsorte/#main_content`},
		{"11", args{"/s/search term"}, "s", "search term"},
		{"12", args{"/ml/antoni_comín"}, "ml", "antoni_comín"},
		{"13", args{"/s/Änderungen erklären"}, "s", "Änderungen erklären"},
		{"14", args{"///asterisk/admin/config.php"}, "asterisk", "admin/config.php"},
		{"15", args{"/p/15ee22f54a6f700e"}, "p", "15ee22f54a6f700e"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRHead, gotRTail := URLparts(tt.args.aURL)
			if gotRHead != tt.wantRHead {
				t.Errorf("URLpath1() gotRHead = {%v}, want {%v}", gotRHead, tt.wantRHead)
			}
			if gotRTail != tt.wantRTail {
				t.Errorf("URLpath1() gotRTail = {%v}, want {%v}", gotRTail, tt.wantRTail)
			}
		})
	}
} // TestURLparts()
