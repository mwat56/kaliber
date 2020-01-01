/*
   Copyright © 2020 M.Watermann, 10247 Berlin, Germany
                  All rights reserved
               EMail : <support@mwat.de>
*/

package kaliber

import (
	"testing"
)

func Test_removeCSSwhitespace(t *testing.T) {
	c0 := []byte(``)
	w0 := []byte(``)
	c1 := []byte(`/*
this are my css rules
*/
`)
	w1 := []byte(``)
	c2 := []byte(`
body { /* default background colour */
	background : #f9f9f3;
}
`)
	w2 := []byte(`body{background:#f9f9f3}`)
	c3 := []byte(`
@media screen {
	body {
		background : #f9f9f3;
	}
}

@media print {
	body {
		background : #fff;
	}
}

`)
	w3 := []byte(`@media screen{body{background:#f9f9f3}}@media print{body{background:#fff}}`)

	type args struct {
		aCSS []byte
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		// TODO: Add test cases.
		{" 0", args{c0}, w0},
		{" 1", args{c1}, w1},
		{" 2", args{c2}, w2},
		{" 3", args{c3}, w3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := removeCSSwhitespace(tt.args.aCSS); string(got) != string(tt.want) {
				t.Errorf("removeCSSwhitespace() = '%s',\nwant '%s'", got, tt.want)
			}
		})
	}
} // Test_removeCSSwhitespace()

func Test_createMinFile(t *testing.T) {
	c1 := `./css/stylesheet.css`
	c2 := `./css/dark.css`
	c3 := `./css/light.css`
	c4 := `./css/fonts.css`
	type args struct {
		aCSSName string
		aMinName string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{" 1", args{c1, c1 + cssNameSuffix}, false},
		{" 2", args{c2, c2 + cssNameSuffix}, false},
		{" 3", args{c3, c3 + cssNameSuffix}, false},
		{" 4", args{c4, c4 + cssNameSuffix}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := createMinFile(tt.args.aCSSName, tt.args.aMinName); (err != nil) != tt.wantErr {
				t.Errorf("createMinFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
} // Test_createMinFile()
