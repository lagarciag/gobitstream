package gobitstream

import (
	"github.com/lagarciag/gobitstream/tests"
	"testing"
)

func TestExtractAndSetBitsFromSlice(t *testing.T) {
	_, a := tests.InitTest(t)
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
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("Extract: %X", tc.expected)
			actual, err := Get64BitsFieldFromSlice(tc.slice, tc.width, tc.offset)
			a.Nil(err)
			a.Equal(tc.expected, actual)
		})
	}
}

func TestExtractAndSetBitsFromSlice2(t *testing.T) {
	_, a := tests.InitTest(t)
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
	actual, err := Get64BitsFieldFromSlice(slice, width, offset)
	a.Nil(err)
	a.Equal(actual, expected)

	err = Set64BitsFieldToWordSlice(slice2, expected, width, offset)
	a.Nil(err)

	a.Equal(expected2, slice2)

	err = SetFieldToSlice(slice3, []uint64{expected}, width, offset)
	a.Equal(expected2, slice3)
}
