package gobitstream_test

import (
	"github.com/lagarciag/gobitstream"
	"github.com/lagarciag/gobitstream/tests"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestExtractAndSetBitsSliceFromSlice(t *testing.T) {
	_, a, _ := tests.InitTest(t)
	testCases := []struct {
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
			slice:    []uint64{0x00000000FFFFFFFF, 0xFFFFFFFF00000000},
			width:    32,
			offset:   32,
			expected: 0x0,
		},
	}

	for _, tc := range testCases {
		t.Log(tc.name)
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("Extract: %X", tc.expected)
			actual, err := gobitstream.GetFieldFromSlice(tc.slice, tc.width, tc.offset)
			if err != nil {
				if len(tc.slice) == 0 {
					return
				}
			}
			a.Nil(err)
			a.Equal([]uint64{tc.expected}, actual)
		})
	}
}

func TestExtractAndSetSliceBitsFromSlice2(t *testing.T) {
	_, a, _ := tests.InitTest(t)
	slice := []uint64{0x0123456789abcdef, 0xfedcba9876543210}
	width := uint64(56)
	offset := uint64(16)
	slice2 := []uint64{0x0, 0x0}
	slice3 := []uint64{0x0, 0x0}

	expected := (slice[0] >> offset) & ((1 << width) - 1)
	remainingBits := width - (64 - offset)
	expected |= slice[1] & ((1 << remainingBits) - 1) << (64 - offset)

	expected2 := []uint64{0x123456789ab0000, 0x10}

	t.Logf("Extract: %X", expected)
	actual, err := gobitstream.GetFieldFromSlice(slice, width, offset)
	a.Nil(err)
	a.Equal(actual[0], expected)

	slice2R, err := gobitstream.Set64BitsFieldToSlice(slice2, expected, width, offset)
	a.Nil(err)

	a.Equal(expected2, slice2R)

	slice3, err = gobitstream.SetFieldToSlice(slice3, []uint64{expected}, width, offset)
	a.Equal(expected2, slice3)
}

func TestSetSliceBitsFieldToWordSlice(t *testing.T) {
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
		// add more test cases here
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Log("running test case ", i)
			var err error
			tt.destinationField, err = gobitstream.Set64BitsFieldToSlice(tt.destinationField, tt.inputField, tt.widthInBits, tt.offsetInBits)

			if tt.expectedErr != nil {
				if err == nil || err.Error() != tt.expectedErr.Error() {
					t.Errorf("expected error %v, got %v", tt.expectedErr, err)
				}
			} else if err != nil {
				t.Errorf("expected no error, got %v", err)
			} else if !a.Equal(tt.expectedOutput, tt.destinationField) {
				t.Errorf("expected output %X, got %X", tt.expectedOutput, tt.destinationField)
			}
		})
	}
}

func TestSetFieldToSlice(t *testing.T) {

	_, a, _ := tests.InitTest(t)

	t.Run("Should set field to slice without error", func(t *testing.T) {
		dstSlice := []uint64{0, 0, 0}
		field := []uint64{6, 7, 8}
		width := uint64(192) // 64 bits for each field
		offset := uint64(0)

		dstSlice, err := gobitstream.SetFieldToSlice(dstSlice, field, width, offset)

		assert.NoError(t, err)
		expectedSlice := []uint64{6, 7, 8}
		assert.Equal(t, expectedSlice, dstSlice)
	})

	t.Run("Should return error if offset is out of range", func(t *testing.T) {
		dstSlice := []uint64{1, 2, 3}
		field := []uint64{6, 7, 8}
		width := uint64(192)  // 64 bits for each field
		offset := uint64(300) // out of range

		dstSlice, err := gobitstream.SetFieldToSlice(dstSlice, field, width, offset)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "is out of range")
	})

	t.Run("Should return error if width is larger than the size of the field", func(t *testing.T) {
		dstSlice := []uint64{1, 2, 3}
		field := []uint64{6, 7, 8}
		width := uint64(300) // out of range
		offset := uint64(0)

		dstSlice, err := gobitstream.SetFieldToSlice(dstSlice, field, width, offset)

		assert.Error(t, err)
		// Here, you should check if the error returned is the one you expect.
		// Since I do not have the exact implementation of the function Set64BitsFieldToSlice, I cannot provide the expected error.
	})

	t.Run("Should handle width and offset being zero appropriately", func(t *testing.T) {
		dstSlice := []uint64{1, 2, 3}
		field := []uint64{6, 7, 8}
		width := uint64(0)
		offset := uint64(0)

		dstSlice, err := gobitstream.SetFieldToSlice(dstSlice, field, width, offset)

		if !a.NotNil(err) {
			t.Errorf("error: %+v", err)
		} else {
			t.Logf("expected error: %v", err)
		}

	})
}

func TestGetBitsFromSlice2(t *testing.T) {
	_, a, _ := tests.InitTest(t)
	tests := []struct {
		name     string
		slice    []uint64
		width    uint64
		offset   uint64
		expected []uint64
	}{
		{
			name:     "Case 1",
			slice:    []uint64{0x1111111122222222, 0x3333333344444444, 0x0123456789ABCDEF},
			width:    64,
			offset:   32,
			expected: []uint64{0x4444444411111111},
		},
		{
			name:     "Case 2",
			slice:    []uint64{0x1122334455667788, 0x99AABBCCDDEEFF00, 0x0123456789ABCDEF},
			width:    40,
			offset:   20,
			expected: []uint64{0x1223344556},
		},
		{
			name:     "Case 3",
			slice:    []uint64{0x0123456789abcdef, 0xfedcba9876543210},
			width:    56,
			offset:   16,
			expected: []uint64{0x100123456789ab},
		},
		{
			name:     "Case 4: Empty slice",
			slice:    []uint64{},
			width:    64,
			offset:   32,
			expected: []uint64{0},
		},
		{
			name:     "Case 5: Single element slice, max width",
			slice:    []uint64{0xFFFFFFFFFFFFFFFF},
			width:    64,
			offset:   0,
			expected: []uint64{0xFFFFFFFFFFFFFFFF},
		},
		{
			name:     "Case 6: Single element slice, min width",
			slice:    []uint64{0xFFFFFFFFFFFFFFFF},
			width:    1,
			offset:   0,
			expected: []uint64{1},
		},
		{
			name:     "Case 7: Offset at the boundary",
			slice:    []uint64{0xe544414ce0c1a0c, 0xd44c33ac9c0b945, 0xfd2c1056a83a28c},
			width:    2,
			offset:   116,
			expected: []uint64{0x0},
		},
	}
	for _, tc := range tests {
		t.Log(tc.name)
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("Extract: %X", tc.expected)

			actual, err := gobitstream.GetFieldFromSlice(tc.slice, tc.width, tc.offset)
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

func TestSet64BitsFieldToWordSlice2(t *testing.T) {

	_, a, _ := tests.InitTest(t)

	tests := []struct {
		name             string
		destinationField []uint64
		inputField       []uint64
		widthInBits      uint64
		offsetInBits     uint64
		expectedErr      error
		expectedOutput   []uint64
	}{
		{
			name:             "Case 1: Normal case within single element",
			destinationField: []uint64{0x0, 0x0, 0x0},
			inputField:       []uint64{0x123456789ABCDEF},
			widthInBits:      32,
			offsetInBits:     0,
			expectedErr:      nil,
			expectedOutput:   []uint64{0x89ABCDEF, 0x0, 0x0},
		},
		{
			name:             "Case 2: Normal case across two elements",
			destinationField: []uint64{0x0, 0x0, 0x0},
			inputField:       []uint64{0x123456789ABCDEF},
			widthInBits:      48,
			offsetInBits:     40,
			expectedErr:      nil,
			expectedOutput:   []uint64{0xABCDEF0000000000, 0x456789, 0x0},
		},
		{
			name:             "Case 3: Setting bits with offset more than 64",
			destinationField: []uint64{0x0, 0x0, 0x0},
			inputField:       []uint64{0x123456789ABCDEF},
			widthInBits:      32,
			offsetInBits:     80,
			expectedErr:      nil,
			expectedOutput:   []uint64{0x0, 0x89ABCDEF0000, 0},
		},
		{
			name:             "Case 4: from random test",
			destinationField: []uint64{0xf089654672e1a1d, 0x15b46f84dcbba05, 0x87cf056eeec624, 0xb96c5178b740f45},
			inputField:       []uint64{0x40F450087C},
			widthInBits:      40,
			offsetInBits:     172,
			expectedErr:      nil,
			expectedOutput:   []uint64{0xf089654672e1a1d, 0x15b46f84dcbba05, 0x87cf056eeec624, 0xb96c5178b740f45},
		},
		{
			name:             "Case 5: from random test",
			destinationField: []uint64{0x71ae58272875c0c, 0x819550d59ea0344, 0x9d4b3b29d0c0a12, 0xb7ab52bf87059c5},
			inputField:       []uint64{0x1C509D4B3B29D0},
			widthInBits:      55,
			offsetInBits:     148,
			expectedErr:      nil,
			expectedOutput:   []uint64{0x71ae58272875c0c, 0x819550d59ea0344, 0x9d4b3b29d0c0a12, 0xb7ab52bf87059c5},
		},
		{
			name:             "Case 6: from random test",
			destinationField: []uint64{0xabe6e3db3b17058, 0x7ac13a4d750cbdf, 0xfbe87ba50a45cfb, 0x8a9a3d77d648716},
			inputField:       []uint64{0x1C583EFA},
			widthInBits:      32,
			offsetInBits:     174,
			expectedErr:      nil,
			expectedOutput:   []uint64{0xabe6e3db3b17058, 0x7ac13a4d750cbdf, 0xfbe87ba50a45cfb, 0x8a9a3d77d648716},
		},
		{
			name:             "Case 7: from random test",
			destinationField: []uint64{0xe544414ce0c1a0c, 0xd44c33ac9c0b945, 0xfd2c1056a83a28c},
			inputField:       []uint64{0x0},
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
			result, err = gobitstream.SetFieldToSlice(tt.destinationField, tt.inputField, tt.widthInBits, tt.offsetInBits)

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

func TestExtractAndSetSliceBitsFromSliceRandom2(t *testing.T) {
	_, a, rnd := tests.InitTest(t)
	rnd.Seed(1685376042)
	const round = 1000
	//sliceSize := rnd.Intn(1) + 2
	const sliceSize = 3
	for i := 0; i < round; i++ {

		initialSlice := make([]uint64, sliceSize)

		for i := range initialSlice {
			initialSlice[i] = uint64(rnd.Intn(0xFFFFFFFFFFFFFFF))
		}

		secondSlice := make([]uint64, len(initialSlice))
		copy(secondSlice, initialSlice)
		t.Logf("second Slice:    %x", secondSlice)
		t.Logf("init   Slice:    %x", initialSlice)

		maxWidth := len(secondSlice)*64 - 3
		width := uint64(rnd.Intn(maxWidth + 1))
		offset := uint64(rnd.Intn(len(initialSlice)*64-int(width)) + 1)

		val1, err := gobitstream.GetFieldFromSlice(initialSlice, width, offset)
		if width == 0 {
			if !a.NotNil(err) {
				t.Log("width is zero: ", width)
			}
			continue
		}
		if !a.Nil(err) {
			t.Errorf("expected no error, got %+v+", err)
		}

		t.Logf("Extract: %X", val1)

		secondSliceR, err := gobitstream.SetFieldToSlice(secondSlice, val1, width, offset)
		a.Nil(err)

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
