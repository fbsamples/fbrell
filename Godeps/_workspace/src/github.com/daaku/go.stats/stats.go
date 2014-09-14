// Package stats provides a simple indirection layer to make it easy
// to add stats without locking into a specific backend.
package stats

// The abstract API provided by a backend.
type Backend interface {
	Count(name string, count int)
	Record(name string, value float64)
}
