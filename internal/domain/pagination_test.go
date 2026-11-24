package domain

import (
	"slices"
	"testing"
)

func TestPageCountRoundsUpAndNeverIsZero(t *testing.T) {
	for total, want := range map[int]int{0: 1, 1: 1, 20: 1, 21: 2, 45: 3} {
		if got := PageCount(total, 20); got != want {
			t.Errorf("%d items: got %d pages, want %d", total, got, want)
		}
	}
}

func TestPageReturnsTheItemsOfThatPage(t *testing.T) {
	items := []int{1, 2, 3, 4, 5}
	if got := Page(items, 0, 2); !slices.Equal(got, []int{1, 2}) {
		t.Errorf("first page: %v", got)
	}
	if got := Page(items, 2, 2); !slices.Equal(got, []int{5}) {
		t.Errorf("incomplete last page: %v", got)
	}
}

func TestPageOutOfRangeIsClampedToTheLastOrFirst(t *testing.T) {
	items := []int{1, 2, 3}
	if got := Page(items, 9, 2); !slices.Equal(got, []int{3}) {
		t.Errorf("past the end: %v", got)
	}
	if got := Page(items, -1, 2); !slices.Equal(got, []int{1, 2}) {
		t.Errorf("before the start: %v", got)
	}
	if got := Page([]int{}, 0, 2); len(got) != 0 {
		t.Errorf("empty list: %v", got)
	}
}
