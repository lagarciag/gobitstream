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
		t.Log(tc.name)
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("Extract: %X", tc.expected)
			actual, err := Get64BitsFieldFromSlice(tc.slice, tc.width, tc.offset)
			if err != nil {
				if len(tc.slice) == 0 {
					return
				}
			}
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

func TestSet64BitsFieldToWordSlice(t *testing.T) {
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
			err := Set64BitsFieldToWordSlice(tt.destinationField, tt.inputField, tt.widthInBits, tt.offsetInBits)

			if tt.expectedErr != nil {
				if err == nil || err.Error() != tt.expectedErr.Error() {
					t.Errorf("expected error %v, got %v", tt.expectedErr, err)
				}
			} else if err != nil {
				t.Errorf("expected no error, got %v", err)
			} else if !equal(tt.destinationField, tt.expectedOutput) {
				t.Errorf("expected output %X, got %X", tt.expectedOutput, tt.destinationField)
			}
		})
	}
}

func equal(a, b []uint64) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}
