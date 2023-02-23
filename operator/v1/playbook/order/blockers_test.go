package order

import (
	"reflect"
	"testing"
)

func TestCommonStringsInSlices(t *testing.T) {
	testCases := []struct {
		a, b, want []string
	}{
		{
			[]string{"a", "b", "c"},
			[]string{"c", "b", "a"},
			[]string{"a", "b", "c"},
		},
		{
			[]string{"a", "b", "c"},
			[]string{"d", "e", "f"},
			[]string{},
		},
		{
			[]string{"apple", "banana", "cherry"},
			[]string{"cherry", "apple"},
			[]string{"apple", "cherry"},
		},
		{
			[]string{},
			[]string{},
			[]string{},
		},
		{
			[]string{"dog", "cat", "bird"},
			[]string{},
			[]string{},
		},
	}

	for _, tc := range testCases {
		got := commonStringsInSlices(tc.a, tc.b)
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("commonStringsInSlices(%v, %v) = %v; want %v", tc.a, tc.b, got, tc.want)
		}
	}
}
