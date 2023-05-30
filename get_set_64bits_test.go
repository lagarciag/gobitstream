package gobitstream_test

import (
	"github.com/lagarciag/gobitstream"
	"github.com/lagarciag/gobitstream/tests"
	"testing"
)

func TestGetBitsFromSlice(t *testing.T) {
	_, a, _ := tests.InitTest(t)
	tests := []struct {
		name     string
		slice    []uint64
		width    uint64
		offset   uint64
		expected uint64
	}{
		{
			name:     "Case 1",
			slice:    []uint64{0x1111111122222222, 0x3333333344444444, 0x0123456789ABCDEF},
			width:    64,
			offset:   32,
			expected: 0x4444444411111111,
		},
		{
			name:     "Case 2",
			slice:    []uint64{0x1122334455667788, 0x99AABBCCDDEEFF00, 0x0123456789ABCDEF},
			width:    40,
			offset:   20,
			expected: 0x1223344556,
		},
		{
			name:     "Case 3",
			slice:    []uint64{0x0123456789abcdef, 0xfedcba9876543210},
			width:    56,
			offset:   16,
			expected: 0x100123456789ab,
		},
		{
			name:     "Case 4: Empty slice",
			slice:    []uint64{},
			width:    64,
			offset:   32,
			expected: 0,
		},
		{
			name:     "Case 5: Single element slice, max width",
			slice:    []uint64{0xFFFFFFFFFFFFFFFF},
			width:    64,
			offset:   0,
			expected: 0xFFFFFFFFFFFFFFFF,
		},
		{
			name:     "Case 6: Single element slice, min width",
			slice:    []uint64{0xFFFFFFFFFFFFFFFF},
			width:    1,
			offset:   0,
			expected: 1,
		},
		{
			name:     "Case 7: Offset at the boundary",
			slice:    []uint64{0xe544414ce0c1a0c, 0xd44c33ac9c0b945, 0xfd2c1056a83a28c},
			width:    2,
			offset:   116,
			expected: 0x0,
		},
	}
	for _, tc := range tests {
		t.Log(tc.name)
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("Extract: %X", tc.expected)

			actual, err := gobitstream.Get64BitsFieldFromSlice(nil, tc.slice, tc.width, tc.offset)
			if err != nil {
				if len(tc.slice) == 0 {
					return
				}
			}
			a.Nil(err)
			if !a.Equal(tc.expected, actual) {
				t.Logf("slice: %X", tc.slice)
				t.Logf("width: %d", tc.width)
				t.Logf("offset: %d", tc.offset)
			}
		})
	}
}

func TestSet64BitsFieldToWordSlice(t *testing.T) {

	_, a, _ := tests.InitTest(t)

	tests := []struct {
		name             string
		destinationField []uint64
		inputField       uint64
		widthInBits      uint64
		offsetInBits     uint64
		expectedErr      error
		expectedOutput   []uint64
	}{
		{
			name:             "Case 1: Normal case within single element",
			destinationField: []uint64{0x0, 0x0, 0x0},
			inputField:       0x123456789ABCDEF,
			widthInBits:      32,
			offsetInBits:     0,
			expectedErr:      nil,
			expectedOutput:   []uint64{0x89ABCDEF, 0x0, 0x0},
		},
		{
			name:             "Case 2: Normal case across two elements",
			destinationField: []uint64{0x0, 0x0, 0x0},
			inputField:       0x123456789ABCDEF,
			widthInBits:      48,
			offsetInBits:     40,
			expectedErr:      nil,
			expectedOutput:   []uint64{0xABCDEF0000000000, 0x456789, 0x0},
		},
		{
			name:             "Case 3: Setting bits with offset more than 64",
			destinationField: []uint64{0x0, 0x0, 0x0},
			inputField:       0x123456789ABCDEF,
			widthInBits:      32,
			offsetInBits:     80,
			expectedErr:      nil,
			expectedOutput:   []uint64{0x0, 0x89ABCDEF0000, 0},
		},
		{
			name:             "Case 4: from random test",
			destinationField: []uint64{0xf089654672e1a1d, 0x15b46f84dcbba05, 0x87cf056eeec624, 0xb96c5178b740f45},
			inputField:       0x40F450087C,
			widthInBits:      40,
			offsetInBits:     172,
			expectedErr:      nil,
			expectedOutput:   []uint64{0xf089654672e1a1d, 0x15b46f84dcbba05, 0x87cf056eeec624, 0xb96c5178b740f45},
		},
		{
			name:             "Case 5: from random test",
			destinationField: []uint64{0x71ae58272875c0c, 0x819550d59ea0344, 0x9d4b3b29d0c0a12, 0xb7ab52bf87059c5},
			inputField:       0x1C509D4B3B29D0,
			widthInBits:      55,
			offsetInBits:     148,
			expectedErr:      nil,
			expectedOutput:   []uint64{0x71ae58272875c0c, 0x819550d59ea0344, 0x9d4b3b29d0c0a12, 0xb7ab52bf87059c5},
		},
		{
			name:             "Case 6: from random test",
			destinationField: []uint64{0xabe6e3db3b17058, 0x7ac13a4d750cbdf, 0xfbe87ba50a45cfb, 0x8a9a3d77d648716},
			inputField:       0x1C583EFA,
			widthInBits:      32,
			offsetInBits:     174,
			expectedErr:      nil,
			expectedOutput:   []uint64{0xabe6e3db3b17058, 0x7ac13a4d750cbdf, 0xfbe87ba50a45cfb, 0x8a9a3d77d648716},
		},
		{
			name:             "Case 7: from random test",
			destinationField: []uint64{0xe544414ce0c1a0c, 0xd44c33ac9c0b945, 0xfd2c1056a83a28c},
			inputField:       0x0,
			widthInBits:      2,
			offsetInBits:     116,
			expectedErr:      nil,
			expectedOutput:   []uint64{0xe544414ce0c1a0c, 0xd44c33ac9c0b945, 0xfd2c1056a83a28c},
		},
		// add more test cases here
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Log("running test case ", i)
			var err error
			var result []uint64
			result, err = gobitstream.Set64BitsFieldToSlice(tt.destinationField, tt.inputField, tt.widthInBits, tt.offsetInBits)

			if tt.expectedErr != nil {
				if err == nil || err.Error() != tt.expectedErr.Error() {
					t.Errorf("expected error %v, got %v", tt.expectedErr, err)
				}
			} else if err != nil {
				t.Errorf("expected no error, got %v", err)
			} else if !a.Equal(tt.expectedOutput, result) {

				t.Logf("result: %X", result)
				t.Logf("width %d", tt.widthInBits)
				t.Logf("offset %d", tt.offsetInBits)
				t.Errorf("expected output %X, got %X", tt.expectedOutput, tt.destinationField)
			}
		})
	}
}

func TestExtractAndSetSliceBitsFromSliceRandom(t *testing.T) {
	_, a, rnd := tests.InitTest(t)
	const round = 1000
	sliceSize := rnd.Intn(100) + 2
	//rnd.Seed(1685371794)
	for i := 0; i < round; i++ {

		initialSlice := make([]uint64, sliceSize)

		for i := range initialSlice {
			initialSlice[i] = uint64(rnd.Intn(0xFFFFFFFFFFFFFFF))
		}

		secondSlice := make([]uint64, len(initialSlice))

		copy(secondSlice, initialSlice)

		maxWidth := 64 - 3
		width := uint64(rnd.Intn(maxWidth + 1))
		offset := uint64(rnd.Intn(len(initialSlice)*64-int(width)) + 1)

		val1, err := gobitstream.Get64BitsFieldFromSlice(nil, initialSlice, width, offset)
		if width == 0 {
			a.NotNil(err)
			continue
		}
		a.Nil(err)

		//t.Logf("Extract: %X", val1)

		secondSliceR, err := gobitstream.Set64BitsFieldToSlice(secondSlice, val1, width, offset)

		if !(a.Equal(initialSlice, secondSliceR)) {
			t.Logf("width: %d", width)
			t.Logf("offset: %d", offset)
			t.Logf("val1: 0x%x", val1)
			t.Logf("initial Slice:   %x", initialSlice)
			t.Logf("second Slice:    %x", secondSlice)
			t.Logf("second SliceR:   %x", secondSliceR)
			t.Logf("width: %d offset %d", width, offset)

			t.FailNow()
		}
	}
}

func BenchmarkGet64BitsFieldFromSliceWithInputBuffer(b *testing.B) {
	bits := make([]uint64, 128)
	inputBuffer := make([]uint64, 0, 128)
	for i := 0; i < b.N; i++ {
		_, err := gobitstream.Get64BitsFieldFromSlice(inputBuffer, bits, 64, 32)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGet64BitsFieldFromSliceWithOutInputBuffer(b *testing.B) {
	bits := make([]uint64, 128)
	for i := 0; i < b.N; i++ {
		_, err := gobitstream.Get64BitsFieldFromSlice(nil, bits, 64, 32)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSet64BitsFieldToSlice(b *testing.B) {
	bits := make([]uint64, 128)
	for i := 0; i < b.N; i++ {
		_, err := gobitstream.Set64BitsFieldToSlice(bits, 64, 64, 32)
		if err != nil {
			b.Fatal(err)
		}
	}
}
