/*
   Copyright © 2019 M.Watermann, 10247 Berlin, Germany
                  All rights reserved
               EMail : <support@mwat.de>
*/

package kaliber

//lint:file-ignore ST1017 - I prefer Yoda conditions

import (
	"testing"
)

func Test_virtLibReadJSONmetadata(t *testing.T) {
	SetCalibreLibraryPath("/var/opt/Calibre")
	var v1 tVirtLibJSON
	tests := []struct {
		name    string
		want    *tVirtLibJSON
		wantErr bool
	}{
		// TODO: Add test cases.
		{" 1", &v1, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := virtlibReadJSONmetadata()
			if (err != nil) != tt.wantErr {
				t.Errorf("virtLibReadJSONmetadata() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if 0 == len(*got) {
				t.Errorf("virtLibReadJSONmetadata() = %v, want %v", len(*got), "> 0")
			}
		})
	}
} // Test_virtLibReadJSONmetadata()

func Test_virtlibGetLibDefs(t *testing.T) {
	SetCalibreLibraryPath("/var/opt/Calibre")
	tests := []struct {
		name    string
		want    *tVirtLibJSON
		wantErr bool
	}{
		// TODO: Add test cases.
		{" 1", nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := virtlibGetLibDefs()
			if (err != nil) != tt.wantErr {
				t.Errorf("virtlibGetLibDefs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if 0 == len(*got) {
				t.Errorf("virtlibGetLibDefs() = %v, want %v", len(*got), "> 0")
			}
		})
	}
} // Test_virtlibGetLibDefs()

func Test_GetVirtLibList(t *testing.T) {
	SetCalibreLibraryPath("/var/opt/Calibre")
	wl1 := &TvirtLibMap{}
	tests := []struct {
		name    string
		want    *TvirtLibMap
		wantErr bool
	}{
		// TODO: Add test cases.
		{" 1", wl1, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetVirtLibList()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetVirtLibList() error = %v,\nwantErr %v", err, tt.wantErr)
				return
			}
			if 0 == len(*got) {
				t.Errorf("GetVirtLibList() = %v, want %v", len(*got), "> 0")
			}
		})
	}
} // Test_GetVirtLibList()

func TestGetVirtLibOptions(t *testing.T) {
	SetCalibreLibraryPath("/var/opt/Calibre")
	type args struct {
		aSelected string
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
		{" 1", args{""}},
		{" 2", args{"Warentest"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetVirtLibOptions(tt.args.aSelected); 0 == len(got) {
				t.Errorf("GetVirtLibOptions() = %v,\nwant %v", got, "> 0")
			}
		})
	}
} // TestGetVirtLibOptions()
