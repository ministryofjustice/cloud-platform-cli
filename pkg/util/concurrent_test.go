package util_test

import (
	"reflect"
	"testing"

	"github.com/ministryofjustice/cloud-platform-cli/pkg/util"
)

func Test_Generator(t *testing.T) {
	test01 := []string{"foo", "bar", "baz"}
	test02 := []string{}

	type args struct {
		data []string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{"GIVEN some string array data THEN write the data into channel AND return a readonly channel", args{test01}, test01},
		{"GIVEN an empty string array THEN write nothing into the channel AND close the channel properly to avoid being locked", args{test02}, test02},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			done := make(chan bool)
			defer close(done)
			got := util.Generator(done, tt.args.data...)

			gotFromChannel := []string{}

			for res := range got {
				gotFromChannel = append(gotFromChannel, res)
			}

			if !reflect.DeepEqual(gotFromChannel, tt.want) {
				t.Errorf("Generator() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_FanIn(t *testing.T) {
	test01 := [][]string{{"foo"}, {"bar"}, {"baz"}}
	test01Combined := []string{"foo", "bar", "baz"}
	test02 := [][]string{}
	test02Combined := []string{}

	type args struct {
		channels [][]string
	}

	tests := []struct {
		name string
		args args
		want []string
	}{
		{"GIVEN many channels with data on THEN return a single channel with the combined data", args{test01}, test01Combined},
		{"GIVEN an empty input channels array THEN return a single channel AND close the channels properly to avoid being locked", args{test02}, test02Combined},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			done := make(chan bool)
			defer close(done)

			channels := make([]<-chan string, len(tt.args.channels))

			for i, data := range tt.args.channels {
				newChan := make(chan string, len(data))
				for _, chanData := range data {
					newChan <- chanData
				}
				close(newChan)
				channels[i] = newChan
			}

			got := util.FanIn(done, channels...)

			gotFromChannel := []string{}

			for res := range got {
				gotFromChannel = append(gotFromChannel, res)
			}

			if len(tt.want) != len(gotFromChannel) {
				t.Errorf("FanIn() = %v, want %v", got, tt.want)
			}

			for _, val := range gotFromChannel {
				if !util.Contains(tt.want, val) {
					t.Errorf("FanIn() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}
