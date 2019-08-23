/*
   Copyright © 2019 M.Watermann, 10247 Berlin, Germany
                  All rights reserved
               EMail : <support@mwat.de>
*/

package kaliber

import (
	"testing"
)

func Test_virtLibReadJSONmetadata(t *testing.T) {
	var v1 tVirtLibJSON
	type args struct {
		aFilename string
	}
	tests := []struct {
		name    string
		args    args
		want    *tVirtLibJSON
		wantErr bool
	}{
		// TODO: Add test cases.
		{" 1", args{calibrePreferencesFile}, &v1, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := virtlibReadJSONmetadata(tt.args.aFilename)
			if (err != nil) != tt.wantErr {
				t.Errorf("virtLibReadJSONmetadata() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// if !reflect.DeepEqual(got, tt.want) {
			// 	t.Errorf("virtLibReadJSONmetadata() = %v, want %v", got, tt.want)
			// }
			if 0 == len(*got) {
				t.Errorf("virtLibReadJSONmetadata() = %v, want %v", len(*got), "> 0")
			}
		})
	}
} // Test_virtLibReadJSONmetadata()

func Test_virtlibGetLibDefs(t *testing.T) {
	type args struct {
		aFilename string
	}
	tests := []struct {
		name    string
		args    args
		want    *tVirtLibJSON
		wantErr bool
	}{
		// TODO: Add test cases.
		{" 1", args{calibrePreferencesFile}, nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := virtlibGetLibDefs(tt.args.aFilename)
			if (err != nil) != tt.wantErr {
				t.Errorf("virtlibGetLibDefs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// if !reflect.DeepEqual(got, tt.want) {
			// 	t.Errorf("virtlibGetLibDefs() = %v, want %v", got, tt.want)
			// }
			if 0 == len(*got) {
				t.Errorf("virtlibGetLibDefs() = %v, want %v", len(*got), "> 0")
			}
		})
	}
} // Test_virtlibGetLibDefs()

func Test_GetVirtLibList(t *testing.T) {
	wl1 := &TvirtLibMap{}
	type args struct {
		aFilename string
	}
	tests := []struct {
		name    string
		args    args
		want    *TvirtLibMap
		wantErr bool
	}{
		// TODO: Add test cases.
		{" 1", args{calibrePreferencesFile}, wl1, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetVirtLibList(tt.args.aFilename)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetVirtLibList() error = %v,\nwantErr %v", err, tt.wantErr)
				return
			}
			// if !reflect.DeepEqual(got, tt.want) {
			// 	t.Errorf("GetVirtLibList() = %v, want %v", got, tt.want)
			// }
			if 0 == len(*got) {
				t.Errorf("GetVirtLibList() = %v, want %v", len(*got), "> 0")
			}
		})
	}
} // Test_GetVirtLibList()
