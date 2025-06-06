package users

import (
	"sync"
	"testing"
)

// The value is an empty struct because the user requested an "empty struct".
// Using an empty struct means zero memory allocation for the value itself,
// which is a common pattern when only the key's presence matters.
type empty struct{}

const numElements = 10000000

// BenchmarkMap measures the time it takes to load 10,000 elements
// into a regular Go map.
// BenchmarkMap-16    	       2	 736,512,316 ns/op
func BenchmarkMap(b *testing.B) {
	// b.N is the number of iterations the benchmark will run.
	// We run the setup (map creation) inside the loop for each iteration
	// to ensure a fresh state for each benchmark run, simulating repeated operations.
	for i := 0; i < b.N; i++ {
		// Initialize the map with a suitable capacity to reduce reallocations
		// during the benchmark run.
		m := make(map[uint64]empty, numElements)

		// Stop the timer during the setup phase to ensure only the
		// actual work (loading elements) is measured.
		b.StopTimer()
		b.StartTimer()

		// Load 10,000 elements into the map.
		for j := 0; j < numElements; j++ {
			m[uint64(j)] = empty{}
		}
	}
}

// BenchmarkSyncMap measures the time it takes to load 10,000 elements
// into a sync.Map.
// cpu: AMD Ryzen 7 7840HS w/ Radeon 780M Graphics
// BenchmarkSyncMap
// BenchmarkSyncMap-16    	       1	3,909,249,515 ns/op
func BenchmarkSyncMap(b *testing.B) {
	// Similar to BenchmarkMap, we set up the sync.Map inside the loop
	// for each iteration to ensure a fresh state.
	for i := 0; i < b.N; i++ {
		var sm sync.Map

		// Stop the timer during setup.
		b.StopTimer()
		b.StartTimer()

		// Load 10,000 elements into the sync.Map.
		// For sync.Map, we use the Store method.
		for j := 0; j < numElements; j++ {
			sm.Store(uint64(j), empty{})
		}
	}
}

/*
To run this benchmark:

1. Save the code above to a file named `benchmark_test.go`.
2. Open your terminal or command prompt.
3. Navigate to the directory where you saved the file.
4. Run the benchmark using the command: `go test -bench=. -benchtime=60s`

Explanation of the command:
- `go test`: The command to run Go tests and benchmarks.
- `-bench=.`: Runs all benchmark functions in the current package.
- `-benchtime=60s`: Specifies that each benchmark function should run for at least 60 seconds.
  This allows for more accurate and stable results, as requested.
*/
