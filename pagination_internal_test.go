package ui

import (
	"slices"
	"testing"
)

func TestPageWindow(t *testing.T) {
	tests := []struct {
		page, pages int
		want        []int
	}{
		{1, 1, []int{1}},
		{3, 5, []int{1, 2, 3, 4, 5}},
		{4, 7, []int{1, 2, 3, 4, 5, 6, 7}},
		{1, 20, []int{1, 2, 3, 4, 5, 0, 20}},
		{4, 20, []int{1, 2, 3, 4, 5, 0, 20}},
		{5, 20, []int{1, 0, 4, 5, 6, 0, 20}},
		{10, 20, []int{1, 0, 9, 10, 11, 0, 20}},
		{17, 20, []int{1, 0, 16, 17, 18, 19, 20}},
		{20, 20, []int{1, 0, 16, 17, 18, 19, 20}},
		{5, 8, []int{1, 0, 4, 5, 6, 7, 8}},
	}
	for _, tc := range tests {
		got := pageWindow(tc.page, tc.pages)
		if !slices.Equal(got, tc.want) {
			t.Errorf("pageWindow(%d, %d) = %v, want %v", tc.page, tc.pages, got, tc.want)
		}
		if len(tc.want) == 7 && len(got) != 7 {
			t.Errorf("pageWindow(%d, %d) has %d slots, want 7", tc.page, tc.pages, len(got))
		}
	}
}
