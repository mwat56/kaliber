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

func Test_readJSONmetaDataFile(t *testing.T) {
	var v1 tVirtLibMap
	type args struct {
		aFilename string
	}
	tests := []struct {
		name    string
		args    args
		want    *tVirtLibMap
		wantErr bool
	}{
		// TODO: Add test cases.
		{" 1", args{virtLibFile}, &v1, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := readJSONmetaDataFile(tt.args.aFilename)
			if (err != nil) != tt.wantErr {
				t.Errorf("readJSONmetaDataFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("readJSONmetaDataFile() = %v, want %v", got, tt.want)
			}
		})
	}
} // Test_readJSONmetaDataFile()
