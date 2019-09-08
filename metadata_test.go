/*
   Copyright © 2019 M.Watermann, 10247 Berlin, Germany
                  All rights reserved
               EMail : <support@mwat.de>
*/

package kaliber

//lint:file-ignore ST1017 - I prefer Yoda conditions

import (
	"reflect"
	"testing"
)

func Test_mdReadMetadataFile(t *testing.T) {
	SetCalibreLibraryPath("/var/opt/Calibre")
	var v1 TVirtualLibraryList
	tests := []struct {
		name    string
		want    *TVirtualLibraryList
		wantErr bool
	}{
		// TODO: Add test cases.
		{" 1", &v1, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := mdReadMetadataFile()
			if (err != nil) != tt.wantErr {
				t.Errorf("mdReadMetadataFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if 0 == len(mdMetadataDbPrefs) {
				t.Errorf("mdReadMetadataFile() = %v, want %v", len(mdMetadataDbPrefs), "> 0")
			}
		})
	}
} // Test_mdReadMetadataFile()

func Test_mdVirtualLibDefinitions(t *testing.T) {
	SetCalibreLibraryPath("/var/opt/Calibre")
	tests := []struct {
		name    string
		want    *TVirtualLibraryList
		wantErr bool
	}{
		// TODO: Add test cases.
		{" 1", nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := mdVirtualLibDefinitions()
			if (err != nil) != tt.wantErr {
				t.Errorf("mdVirtualLibDefinitions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if 0 == len(*got) {
				t.Errorf("mdVirtualLibDefinitions() = %v, want %v", len(*got), "> 0")
			}
		})
	}
} // Test_mdVirtualLibDefinitions()

func Test_GetVirtLibList(t *testing.T) {
	SetCalibreLibraryPath("/var/opt/Calibre")
	wl1 := map[string]string{}
	tests := []struct {
		name    string
		want    map[string]string
		wantErr bool
	}{
		// TODO: Add test cases.
		{" 1", wl1, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetVirtualLibraryList()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetVirtLibList() error = %v,\nwantErr %v", err, tt.wantErr)
				return
			}
			if 0 == len(got) {
				t.Errorf("GetVirtLibList() = %v, want %v", len(got), "> 0")
			}
		})
	}
} // Test_GetVirtLibList()

func Test_GetVirtLibOptions(t *testing.T) {
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
} // Test_GetVirtLibOptions()

func Test_mdReadFieldMetadata(t *testing.T) {
	SetCalibreLibraryPath("/var/opt/Calibre")
	tests := []struct {
		name    string
		wantErr bool
	}{
		// TODO: Add test cases.
		{" 1", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := mdReadFieldMetadata(); (err != nil) != tt.wantErr {
				t.Errorf("mdReadFieldMetadata() error = %v, wantErr %v", err, tt.wantErr)
			}
			if 0 == len(*mdFieldsMetadata) {
				t.Errorf("GetVirtLibList() = %v, want %v", len(*mdFieldsMetadata), "> 0")
			}
		})
	}
} // Test_mdReadFieldMetadata()

func Test_mdGetFieldData(t *testing.T) {
	SetCalibreLibraryPath("/var/opt/Calibre")
	var w1 map[string]interface{}
	type args struct {
		aKey string
	}
	tests := []struct {
		name    string
		args    args
		want    map[string]interface{}
		wantErr bool
	}{
		// TODO: Add test cases.
		{" 1", args{"size"}, w1, false},
		{" 2", args{"#genre"}, w1, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := mdGetFieldData(tt.args.aKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("mdGetFieldData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// if !reflect.DeepEqual(got, tt.want) {
			// 	t.Errorf("mdGetFieldData() = %v, want %v", got, tt.want)
			// }
			if 0 == len(got) {
				t.Errorf("mdGetFieldData() = %v, want %v", len(got), "> 0")
			}
		})
	}
} // Test_mdGetFieldData()

func Test_mdReadVirtualLibraries(t *testing.T) {
	SetCalibreLibraryPath("/var/opt/Calibre")
	tests := []struct {
		name    string
		wantErr bool
	}{
		// TODO: Add test cases.
		{" 1", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := mdReadVirtualLibraries(); (err != nil) != tt.wantErr {
				t.Errorf("mdReadVirtualLibraries() error = %v, wantErr %v", err, tt.wantErr)
			}
			if 0 == len(*mdVirtLibsRaw) {
				t.Errorf("mdReadVirtualLibraries() = %v, want %v", len(*mdVirtLibsRaw), "> 0")
			}
		})
	}
} // Test_mdReadVirtualLibraries()

func Test_GetMetaFieldValue(t *testing.T) {
	type args struct {
		aField string
		aKey   string
	}
	tests := []struct {
		name    string
		args    args
		want    interface{}
		wantErr bool
	}{
		// TODO: Add test cases.
		{" 1", args{"authors", "is_category"}, true, false},
		{" 2", args{"authors", "table"}, "authors", false},
		{" 3", args{"#genre", "is_category"}, true, false},
		{" 4", args{"#genre", "is_custom"}, true, false},
		{" 5", args{"#genre", "table"}, "custom_column_1", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetMetaFieldValue(tt.args.aField, tt.args.aKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetMetaFieldValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetMetaFieldValue() = %v, want %v", got, tt.want)
			}
		})
	}
} // Test_GetMetaFieldValue()

func Test_mdReadHiddenVirtualLibraries(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		// TODO: Add test cases.
		{" 1", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := mdReadHiddenVirtualLibraries(); (err != nil) != tt.wantErr {
				t.Errorf("mdReadHiddenVirtualLibraries() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
} // Test_mdReadHiddenVirtualLibraries()
