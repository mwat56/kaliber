/*
   Copyright © 2019 M.Watermann, 10247 Berlin, Germany
              All rights reserved
          EMail : <support@mwat.de>
*/

package kaliber

import (
	"testing"
)

func Test_makeThumbDir(t *testing.T) {
	d1 := &TDocument{
		ID:   7628,
		path: calibreLibraryPath + "/Spiegel/Der Spiegel (2019-06-01) 23_2019 (7628)",
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
	SetCalibreCachePath("./cache")
	SetCalibreLibraryPath("/var/opt/Calibre/")
	d1 := &TDocument{
		ID:   7628,
		path: "/Spiegel/Der Spiegel (2019-06-01) 23_2019 (7628)",
	}
	w1 := "/home/matthias/devel/Go/src/github.com/mwat56/kaliber/cache/0076/007628.jpg"
	ThumbnailRemove(d1)
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
				t.Errorf("Thumbnail() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Thumbnail() = %v, want %v", got, tt.want)
			}
		})
	}
} // TestThumbnail()

func TestThumbnailName(t *testing.T) {
	SetCalibreCachePath("./cache")
	SetCalibreLibraryPath("/var/opt/Calibre/")
	d1 := &TDocument{
		ID:   7628,
		path: calibreLibraryPath + "/Spiegel/Der Spiegel (2019-06-01) 23_2019 (7628)",
	}
	w1 := "/home/matthias/devel/Go/src/github.com/mwat56/kaliber/cache/0076/007628.jpg"
	ThumbnailRemove(d1)
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
			if got := ThumbnailName(tt.args.aDoc); got != tt.want {
				t.Errorf("ThumbnailName() = %v,\nwant %v", got, tt.want)
			}
		})
	}
} // TestThumbnailName()

func TestThumbnailRemove(t *testing.T) {
	SetCalibreCachePath("./cache")
	SetCalibreLibraryPath("/var/opt/Calibre/")
	d1 := &TDocument{
		ID:   7628,
		path: calibreLibraryPath + "/Spiegel/Der Spiegel (2019-06-01) 23_2019 (7628)",
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ThumbnailRemove(tt.args.aDoc); (err != nil) != tt.wantErr {
				t.Errorf("ThumbnailRemove() error = %v,\nwantErr %v", err, tt.wantErr)
			}
		})
	}
} // TestThumbnailRemove
