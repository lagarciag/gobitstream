package fuzz_test

import (
	"github.com/lagarciag/gobitstream"
	"testing"
)

// fuzzGetAndSetBits uses the get and set operations to fuzz the input
func FuzzGetAndSetBits(f *testing.F) {
	// Define seed corpus
	seedCorpus := []struct {
		input1        uint64
		input2        uint64
		input3        uint64
		input4        uint64
		input5        uint64
		widthInBits   uint64
		offsetInBits  uint64
		expectedField uint64
	}{
		{0x1111111122222222, 0x3333333344444444, 0x0123456789ABCDEF, 0x0000000000000000, 0xFFFFFFFFFFFFFFFF, 64, 32, 0x4444444411111111},
		{0x1122334455667788, 0x99AABBCCDDEEFF00, 0x0123456789ABCDEF, 0x0000000000000000, 0xFFFFFFFFFFFFFFFF, 40, 20, 0x1223344556},
		{0x0123456789abcdef, 0xfedcba9876543210, 0x0000000000000000, 0xFFFFFFFFFFFFFFFF, 0x0123456789ABCDEF, 56, 16, 0x100123456789ab},
		{0x00000000FFFFFFFF, 0xFFFFFFFF00000000, 0xFFFFFFFFFFFFFFFF, 0x0000000000000000, 0x0123456789ABCDEF, 32, 32, 0x0},
		{0x0000000000000000, 0x0000000000000000, 0x0000000000000000, 0xFFFFFFFFFFFFFFFF, 0x0123456789ABCDEF, 0, 0, 0x0},                 // All zeros
		{0xFFFFFFFFFFFFFFFF, 0xFFFFFFFFFFFFFFFF, 0xFFFFFFFFFFFFFFFF, 0x0000000000000000, 0x0123456789ABCDEF, 64, 0, 0xFFFFFFFFFFFFFFFF}, // All ones
	}

	for _, v := range seedCorpus {
		f.Add(v.input1, v.input2, v.input3, v.input4, v.input5, v.widthInBits, v.offsetInBits, v.expectedField)
	}

	f.Fuzz(func(t *testing.T, input1, input2, input3, input4, input5, widthInBits, offsetInBits, expectedField uint64) {
		inputSlice := []uint64{input1, input2, input3, input4, input5}
		field, err := gobitstream.Get64BitsFieldFromSlice(inputSlice, widthInBits, offsetInBits)
		if err != nil {
			if widthInBits == 0 || widthInBits > 64 || (offsetInBits+widthInBits) > 64 {
				return
			}
			t.Fatalf("unexpected error: %v", err)
		}

		//if field != expectedField {
		//	t.Fatalf("expected %v, but got %v", expectedField, field)
		//}

		inputSlice, err = gobitstream.Set64BitsFieldToSlice(inputSlice, field, widthInBits, offsetInBits)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		result, err := gobitstream.Get64BitsFieldFromSlice(inputSlice, widthInBits, offsetInBits)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result != field {
			t.Fatalf("expected %v, but got %v", field, result)
		}
	})
}
