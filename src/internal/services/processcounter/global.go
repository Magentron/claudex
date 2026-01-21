package processcounter

// DefaultCounter is the global ProcessCounter instance used throughout the application.
var DefaultCounter ProcessCounter

func init() {
	DefaultCounter = NewProcessCounter()
}
