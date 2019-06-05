/*
   Copyright © 2019 M.Watermann, 10247 Berlin, Germany
              All rights reserved
          EMail : <support@mwat.de>
*/

package kaliber

import (
	"reflect"
	"testing"
)

func Test_initWSre(t *testing.T) {
	tests := []struct {
		name string
		want int
	}{
		// TODO: Add test cases.
		{" 1", 11},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := initWSre(); got != tt.want {
				t.Errorf("initWSre() = %v, want %v", got, tt.want)
			}
		})
	}
} // Test_initWSre()

func TestNewView(t *testing.T) {
	type args struct {
		aBaseDir string
		aName    string
	}
	tests := []struct {
		name     string
		args     args
		wantView bool
		wantErr  bool
	}{
		// TODO: Add test cases.
		{" 1", args{"./views/", "test1"}, false, true},
		{" 2", args{"./views/", "index"}, true, false},
		{" 3", args{"./views/", "document"}, true, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewView(tt.args.aBaseDir, tt.args.aName)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewView() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if (got != nil) != tt.wantView {
				t.Errorf("NewView() = %v, want %v", got, tt.wantView)
			}
		})
	}
} // TestNewView()

func TestNewDataList(t *testing.T) {
	dl := TDataList{}
	tests := []struct {
		name string
		want *TDataList
	}{
		// TODO: Add test cases.
		{" 1", &dl},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewDataList(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewDataList() = %v, want %v", got, tt.want)
			}
		})
	}
} // TestNewDataList()

func TestTDataList_Set(t *testing.T) {
	dl := NewDataList()
	rl := NewDataList()
	(*rl)["Title"] = "Testing"
	type args struct {
		aKey   string
		aValue interface{}
	}
	tests := []struct {
		name string
		d    *TDataList
		args args
		want *TDataList
	}{
		// TODO: Add test cases.
		{" 1", dl, args{"Title", "Testing"}, rl},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.d.Set(tt.args.aKey, tt.args.aValue); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TDataList.Add() = %v, want %v", got, tt.want)
			}
		})
	}
} // TestTDataList_Set()

func TestNewViewList(t *testing.T) {
	vl := TViewList{}
	tests := []struct {
		name string
		want *TViewList
	}{
		// TODO: Add test cases.
		{" 1", &vl},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewViewList(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewViewList() = %v, want %v", got, tt.want)
			}
		})
	}
} // TestNewViewList()

func TestTViewList_Add(t *testing.T) {
	vname1 := "index"
	vw1, _ := NewView("./views/", vname1)
	vl1 := NewViewList()
	rl1 := NewViewList().Add(vw1)
	type args struct {
		aName string
		aView *TView
	}
	tests := []struct {
		name string
		vl   *TViewList
		args args
		want *TViewList
	}{
		// TODO: Add test cases.
		{" 1", vl1, args{vname1, vw1}, rl1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.vl.Add(tt.args.aView); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TViewList.Add() = %v, want %v", got, tt.want)
			}
		})
	}
} // TestTViewList_Add()
