package common

import (
	"testing"
)

func TestHashIntersection(t *testing.T) {
	t.Run("Test hash intersection", func(t *testing.T) {
		first := []uint{1, 2, 3, 4}
		second := []uint{1, 4, 454, 5}

		intersection := HashIntersection(first, second)
		if len(intersection) != 2 {
			t.Error("Error in create intersections")
		}
	})
}

func Test_removeDuplicateAdjacent(t *testing.T) {
	type args struct {
		text string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "test remove adjacent", args: struct{ text string }{text: "اااهههلللاا"}, want: "اهلا"},
		{name: "test remove adjacent", args: struct{ text string }{text: "hhiiii"}, want: "hi"},
		{name: "test remove adjacent", args: struct{ text string }{text: "تتعبااانة"}, want: "تعبانة"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := removeDuplicateAdjacent(tt.args.text); got != tt.want {
				t.Errorf("removeDuplicateAdjacent() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsBadName(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "test bad names", args: struct{ name string }{name: "خول"}, want: true},
		{name: "test bad names", args: struct{ name string }{name: "خووول"}, want: true},
		{name: "test bad names", args: struct{ name string }{name: "شرموطة"}, want: true},
		{name: "test bad names", args: struct{ name string }{name: "شريف"}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsBadName(tt.args.name); got != tt.want {
				t.Errorf("IsBadName() = %v, want %v", got, tt.want)
			}
		})
	}
}
