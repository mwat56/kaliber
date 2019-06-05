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

func TestNewDataList(t *testing.T) {
	d1 := &TemplateData{}
	tests := []struct {
		name string
		want *TemplateData
	}{
		// TODO: Add test cases.
		{" 1", d1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewTemplateData(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewDataList() = %v,\nwant %v", got, tt.want)
			}
		})
	}
} // TestNewDataList()

func TestTDataList_Set(t *testing.T) {
	d1 := NewTemplateData()
	w1 := NewTemplateData()
	(*w1)["Title"] = "Testing"
	type args struct {
		aKey   string
		aValue interface{}
	}
	tests := []struct {
		name string
		d    *TemplateData
		args args
		want *TemplateData
	}{
		// TODO: Add test cases.
		{" 1", d1, args{"Title", "Testing"}, w1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.d.Set(tt.args.aKey, tt.args.aValue); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TDataList.Add() = %v, want\n%v", got, tt.want)
			}
		})
	}
} // TestTDataList_Set()
