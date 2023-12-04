// Package uniquerand provides a way for generating unique random numbers.
package uniquerand

import (
	"math/rand"
)

// defRandSrc is the random generator used by default.
// it's a function takes an integer, r, and returns a random number in range [0, r).
var defRandSrc = rand.Intn

// defRange is the default range used for the zero value of the Int.
const defRange = 10

const blockSize = 32

type blockType = uint32

// Int allows returning unique random numbers within a predefined range.
// It depends on another source for randomness, keeps track of all generated
// numbers, and makes sure that the returned number is unique.
// The zero value produces unique numbers using math/rand in range [0, 10).
type Int struct {
	r   int         // range
	c   int         // count
	m   blockType   // block num 0 (default memory).
	em  []blockType // block num 1+ (extra memory).
	src func(r int) int
}

type Config struct {
	// Range is the exclusive upper limit of the unique random number that could be
	// generated, starting from 0.
	// If not passed, the default range (10) is used.
	Range int

	// Src is the source of the random numbers.
	// It's a function that takes the Range value, as r, and returns a non-negative
	// pseudo-random number in the half-open interval [0,r).
	// If not passed, math/rand.Intn is used.
	Src func(r int) int
}

// Config readies the random generator to be used, according to the Config provided.
// Each call discards any previous calls to either Config or Reset.
func (uri *Int) Config(c Config) {
	uri.Reset(c.Range)
	uri.src = c.Src
}

// Reset sets the range of the Int generator and resets all previous memory.
// If the given range is less than or equal to zero, the default range (10) is used.
// After calling Reset, the generator is ready to produce unique random numbers within
// the specified range.
// It doesn't change the randomness source.
func (uri *Int) Reset(r int) {
	if r <= 0 {
		r = defRange
	}

	// reset the default fields
	uri.r = r
	uri.c = 0
	uri.m = 0
	uri.em = nil

	// return if we don't need the extra memory
	if r <= blockSize {
		return
	}

	// allocate the extra memory
	l := r / blockSize
	if int(r%blockSize) == 0 {
		l = l - 1
	}
	if l != 0 {
		uri.em = make([]blockType, l)
	}
}

// Range returns the current range of the Int generator, which is the exclusive
// upper limit of the unique random number that could be generated, starting from 0.
// If the range has been set (via Reset or Config), Range returns it.
// If the range has not been set, or it has been set to zero or less, Range returns the
// default range (10).
//
// Example:
//
//	uri := Int{}
//	uri.Range() // returns 10 (default range)
//	uri.Reset(100)
//	uri.Range() // returns 100
func (uri *Int) Range() int {
	if uri.r > 0 {
		return uri.r
	}
	return defRange
}

// Count returns the number of used unique random numbers from the range so far.
// Unique random numbers are generated using Get, and returned using Put.
// At any time, Count() <= Range().
//
// Example:
//
//	uri := Int{}
//	uri.Count() // returns 0
//	uri.Get() // generates a unique random number
//	uri.Count() // returns 1
//	uri.Put(1) // returns the unique random number
//	uri.Count() // returns 0
func (uri *Int) Count() int {
	return uri.c
}

// Used returns true if the unique number provided is currently consumed
// from the specified range, or false otherwise.
func (uri *Int) Used(num int) (ok bool) {
	if num < 0 || num >= uri.Range() {
		return false
	}

	// Block Number, Memory Block, Target Mask, Masked Memory
	_, _, _, mm := uri.has(num)

	// num is already available (not consumed by Get)
	if mm == 0 {
		return false
	}
	return true
}

func (uri *Int) has(n int) (bn int, mb, tm, mm blockType) {
	// get the Block Number
	bn = n / blockSize

	// get the respective Memory Block
	mb = uri.m
	if bn > 0 {
		mb = uri.em[bn-1]
	}

	sv := n % blockSize     // Shift Value
	tm = blockType(1 << sv) // Target Mask
	mm = mb & tm            // Masked Memory
	return
}

// Get returns a unique random number within the specified range and true.
// It returns 0 and false if the specified range ran out of unique numbers.
// The range can be specified using either the Reset or the Config methods.
// If no range is specified, the default range (10) is used.
func (uri *Int) Get() (urn int, ok bool) {
	randSrc := defRandSrc
	if uri.src != nil {
		randSrc = uri.src
	}

	grn := randSrc(uri.Range()) // Generated Random Number

	// Block Number, Memory Block, Target Mask, Masked Memory
	bn, mb, tm, mm := uri.has(grn)

	// Generated Random Number was not generated before
	if mm == 0 {
		// update the respective Memory Block
		if bn > 0 {
			uri.em[bn-1] = mb | tm
		} else {
			uri.m = mb | tm
		}
		urn = grn // Unique Random Number
		uri.c++   // update the counter
		return urn, true
	}

	// Generated Random Number was generated before
	return uri.getSlow()
}

// getSlow is responsible for finding a unique number based on the current
// state of the memory fields (m and em).
// the number returned depends on the history of the generated numbers so far.
func (uri *Int) getSlow() (urn int, ok bool) {
	// loop over the default memory to find the first block that has a zero bit
	for j := 0; j < blockSize; j++ {
		tm := blockType(1 << j) // current block's Target Mask
		mm := uri.m & tm        // current block's Masked Memory
		if mm != 0 {
			continue // the current bit is not zero
		}
		uri.m = uri.m | tm // update the respective Memory Block
		urn = j            // calculate the Unique Random Number
		if urn < uri.Range() {
			uri.c++ // update the counter
			return urn, true
		}
		return 0, false
	}

	// loop over the extra memory to find the first block that has a zero bit
	for i, m := range uri.em {
		// if this block is all 0s, simply set it to 1 and return
		if m == 0 {
			uri.em[i] = 1       // update the respective Memory Block
			urn = i * blockSize // calculate the Unique Random Number
			urn += blockSize    // add the base default memory size
			uri.c++             // update the counter
			return urn, true
		}

		// otherwise, search for the first 0 in this block
		for j := 0; j < blockSize; j++ {
			tm := blockType(1 << j) // current block's Target Mask
			mm := m & tm            // current block's Masked Memory
			if mm != 0 {
				continue // the current bit is not zero
			}
			uri.em[i] = m | tm    // update the respective Memory Block
			urn = i*blockSize + j // calculate the Unique Random Number
			urn += blockSize      // add the base default memory size
			if urn < uri.Range() {
				uri.c++ // update the counter
				return urn, true
			}
			return 0, false
		}
	}

	return 0, false
}

// Put marks the provided number as not used, allowing a previously generated number
// to be generated again later via Get.
// It returns true if the number was generated before, or false otherwise.
func (uri *Int) Put(num int) (ok bool) {
	if num < 0 || num >= uri.Range() {
		return false
	}

	// Block Number, Memory Block, Target Mask, Masked Memory
	bn, mb, tm, mm := uri.has(num)

	// num is already available (not consumed by Get)
	if mm == 0 {
		return false
	}

	// update the respective Memory Block
	if bn > 0 {
		uri.em[bn-1] = mb &^ tm
	} else {
		uri.m = mb &^ tm
	}

	uri.c-- // update the counter
	return true
}
