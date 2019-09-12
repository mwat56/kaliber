/*
   Copyright © 2019 M.Watermann, 10247 Berlin, Germany
              All rights reserved
          EMail : <support@mwat.de>
*/

package kaliber

import (
	"crypto/md5"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func setup4Testing() {
	libPath := `/var/opt/Calibre`
	s := fmt.Sprintf("%x", md5.Sum([]byte(libPath))) // #nosec G401
	ucd, _ := os.UserCacheDir()
	SetCalibreCachePath(filepath.Join(ucd, "kaliber", s))
	SetCalibreLibraryPath(libPath)
} // setup4Testing

func Test_makeThumbDir(t *testing.T) {
	setup4Testing()
	d1 := &TDocument{
		ID:   7628,
		path: quCalibreLibraryPath + "/Spiegel/Der Spiegel (2019-06-01) 23_2019 (7628)",
	}
	type args struct {
		aDoc *TDocument
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{" 1", args{d1}, false},
		{" 2", args{d1}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := makeThumbDir(tt.args.aDoc); (err != nil) != tt.wantErr {
				t.Errorf("makeThumbDir() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
} // Test_makeThumbDir()

func TestThumbnail(t *testing.T) {
	setup4Testing()
	d1 := &TDocument{
		ID:   7628,
		path: "/Spiegel/Der Spiegel (2019-06-01) 23_2019 (7628)",
	}
	w1 := `/home/matthias/.cache/kaliber/abb302a1831a12171af82e2cd612b4e9/0076/007628.jpg`
	thumbnailRemove(d1)
	type args struct {
		aDoc *TDocument
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
		{" 1", args{d1}, w1, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Thumbnail(tt.args.aDoc)
			if (err != nil) != tt.wantErr {
				t.Errorf("Thumbnail() error = %v,\nwantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Thumbnail() = %v,\nwant %v", got, tt.want)
			}
		})
	}
} // TestThumbnail()

func TestThumbnailName(t *testing.T) {
	setup4Testing()
	d1 := &TDocument{
		ID:   7628,
		path: quCalibreLibraryPath + "/Spiegel/Der Spiegel (2019-06-01) 23_2019 (7628)",
	}
	w1 := `/home/matthias/.cache/kaliber/abb302a1831a12171af82e2cd612b4e9/0076/007628.jpg`
	thumbnailRemove(d1)
	type args struct {
		aDoc *TDocument
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
		{" 1", args{d1}, w1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := thumbnailName(tt.args.aDoc); got != tt.want {
				t.Errorf("ThumbnailName() = %v,\nwant %v", got, tt.want)
			}
		})
	}
} // TestThumbnailName()

func TestThumbnailRemove(t *testing.T) {
	setup4Testing()
	d1 := &TDocument{
		ID:   7628,
		path: quCalibreLibraryPath + "/Spiegel/Der Spiegel (2019-06-01) 23_2019 (7628)",
	}
	d2 := NewDocument()
	type args struct {
		aDoc *TDocument
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{" 1", args{d1}, false},
		{" 2", args{d2}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := thumbnailRemove(tt.args.aDoc); (err != nil) != tt.wantErr {
				t.Errorf("ThumbnailRemove() error = %v,\nwantErr %v", err, tt.wantErr)
			}
		})
	}
} // TestThumbnailRemove

func Test_goThumbCleanup(t *testing.T) {
	setup4Testing()
	openDBforTesting()
	tests := []struct {
		name string
	}{
		// TODO: Add test cases.
		{" 1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			goThumbCleanup()
		})
	}
} // Test_goThumbCleanup()

func Test_checkThumbFile(t *testing.T) {
	type args struct {
		aFilename string
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkThumbFile(tt.args.aFilename)
		})
	}
}
