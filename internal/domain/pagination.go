package domain

// LibraryPageSize is how many measurements the library shows on one page.
const LibraryPageSize = 20

// PageCount returns how many pages of size hold total items; an empty list still has one page.
func PageCount(total, size int) int {
	if total <= 0 {
		return 1
	}
	return (total + size - 1) / size
}

// Page returns the items on page index (from 0), clamped to the pages there are.
func Page[T any](items []T, index, size int) []T {
	index = min(max(index, 0), PageCount(len(items), size)-1)
	start := index * size
	return items[start:min(start+size, len(items))]
}
