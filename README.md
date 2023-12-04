# uniquerand [![PkgGoDev](https://pkg.go.dev/badge/github.com/asmsh/uniquerand)](https://pkg.go.dev/github.com/asmsh/uniquerand) [![Go Report Card](https://goreportcard.com/badge/github.com/asmsh/uniquerand)](https://goreportcard.com/report/github.com/asmsh/uniquerand) [![Tests](https://github.com/asmsh/uniquerand/workflows/Tests/badge.svg)](https://github.com/asmsh/uniquerand/actions) [![Go Coverage](https://github.com/asmsh/uniquerand/wiki/coverage.svg)](https://raw.githack.com/wiki/asmsh/uniquerand/coverage.html)

Provides functionality to generate unique random numbers within a specified range.

## Features

- Generate unique random numbers within a specified range.
- Configure the range and the source of random numbers.
- Optimized for short ranges.
- Verify if a number is currently consumed from the specified range.
- Offers a way to reuse (re-generate) a previously generated number.
- Get the number of unique random numbers used from the range so far.
- Modular and easy-to-use API.

## Getting Started

### Prerequisites

- `go1.18` or higher.

### Installation

> go get github.com/asmsh/uniquerand

### Examples

```go
package main

import (
	"fmt"

	"github.com/asmsh/uniquerand"
)

func main() {
	uri := uniquerand.Int{}
	uri.Reset(20)

	for i, ok := uri.Get(); ok; i, ok = uri.Get() {
		// do something with 'i'...
		fmt.Println("i", i)

		// if you are interested in getting the same 'i' value later,
		// return it to the rand source, and it will be randomly returned
		// via a future call to Get.
		// uri.Put(i)
	}

	// print some statistics
	fmt.Println("Num generated:", uri.Count(), "out of", uri.Range())
}
```

## Performance

### Some benchmarks (*):

```
Benchmark_Int/default/Get
Benchmark_Int/default/Get-8     	        53634074	        21.36 ns/op	       0 B/op	       0 allocs/op
Benchmark_Int/range_32/Get-8              	58472168	        21.05 ns/op	       0 B/op	       0 allocs/op
Benchmark_Int/range_64/Get-8              	30796538	        39.14 ns/op	       0 B/op	       0 allocs/op
Benchmark_Int/range_256/Get-8             	 9452564	        124.9 ns/op	       0 B/op	       0 allocs/op
Benchmark_Int/range_1024/Get-8            	 2636010	        454.4 ns/op	       0 B/op	       0 allocs/op
Benchmark_Int/range_4096/Get-8            	  972870	        1780 ns/op	       0 B/op	       0 allocs/op

Benchmark_Int/default/Get_&_Put
Benchmark_Int/default/Get_&_Put-8         	99708832	        12.19 ns/op	       0 B/op	       0 allocs/op
Benchmark_Int/range_32/Get_&_Put-8        	98668995	        11.93 ns/op	       0 B/op	       0 allocs/op
Benchmark_Int/range_64/Get_&_Put-8        	84568580	        14.12 ns/op	       0 B/op	       0 allocs/op
Benchmark_Int/range_256/Get_&_Put-8       	100000000	        11.91 ns/op	       0 B/op	       0 allocs/op
Benchmark_Int/range_1024/Get_&_Put-8      	100000000	        11.78 ns/op	       0 B/op	       0 allocs/op
Benchmark_Int/range_4096/Get_&_Put-8      	100000000	        11.81 ns/op	       0 B/op	       0 allocs/op

Benchmark_Int/default/Rest_&_Get
Benchmark_Int/default/Rest_&_Get-8        	98733601	        12.26 ns/op	       0 B/op	       0 allocs/op
Benchmark_Int/range_32/Rest_&_Get-8       	100000000	        11.84 ns/op	       0 B/op	       0 allocs/op
Benchmark_Int/range_64/Rest_&_Get-8       	61138449	        18.84 ns/op	       4 B/op	       1 allocs/op
Benchmark_Int/range_256/Rest_&_Get-8      	52939521	        22.48 ns/op	      32 B/op	       1 allocs/op
Benchmark_Int/range_1024/Rest_&_Get-8     	41960976	        27.93 ns/op	     128 B/op	       1 allocs/op
Benchmark_Int/range_4096/Rest_&_Get-8     	16797096	        64.83 ns/op	     512 B/op	       1 allocs/op
```

(*): Benchmarks were done on an M2 Macbook Air.

## Theory:

It depends on another Random Number Generator (RNG) (the randomness source) for generating numbers at first.  
It checks for the uniqueness of the generated number against a bits-memory.  
If the generated number (by the RNG) is found to be not unique, it replaces it with the nearest unique number not used within the specified range.

**Limitation:** The algorithm for returning the nearest unique number (replacing algo) is not random, however, it could cause the `Get` method to produce sequential numbers only in the case of the RNG returning the same number over and over.  
