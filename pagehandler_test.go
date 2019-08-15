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

func Test_splitIDterm(t *testing.T) {
	type args struct {
		aTail string
	}
	tests := []struct {
		name      string
		args      args
		wantRID   TID
		wantRTerm string
	}{
		// TODO: Add test cases.
		{" 1", args{""}, 0, ""},
		{" 2", args{"abc/def"}, 0, ""},
		{" 3", args{"5256/Humour"}, 5256, "Humour"},
		{" 4", args{"460/Terry Pratchett"}, 460, "Terry Pratchett"},
		{" 5", args{"477/Harper Torch"}, 477, "Harper Torch"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRID, gotRTerm := splitIDterm(tt.args.aTail)
			if !reflect.DeepEqual(gotRID, tt.wantRID) {
				t.Errorf("splitIDterm() gotRID = %v, want %v", gotRID, tt.wantRID)
			}
			if gotRTerm != tt.wantRTerm {
				t.Errorf("splitIDterm() gotRTerm = %v, want %v", gotRTerm, tt.wantRTerm)
			}
		})
	}
} // Test_splitIDterm()
