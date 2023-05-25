package gobitstream

import (
	"github.hpe.com/hpe-networking/commontest"
	"reflect"
	"testing"
)

func TestShiftSliceofUint64(t *testing.T) {
	_, a := commontest.InitTestLogToStdout(t)
	// Test case 1: Shift by 1 bit

	// Test case 3: Shift by 0 bits (no shift)
	slice3 := []uint64{0x0123456789abcdef, 0xfedcba9876543210}
	expected3 := []uint64{0x0123456789abcdef, 0xfedcba9876543210}
	shiftCount3 := 0
	slice3, err := ShiftSliceOfUint64Left(slice3, shiftCount3)
	if err != nil {
		t.Errorf("ShiftSliceOfUint64Left failed: %v", err)
	}
	if !a.Equal(slice3, expected3) {
		t.Errorf("Test case 3 failed: Expected %X, but got %X", expected3, slice3)
		t.FailNow()
	}

	// Test case 4: Shift by 64 bits (full rotation)
	slice4 := []uint64{0x0123456789abcdef, 0xfedcba9876543210, 0, 0}
	expected4 := []uint64{0, 0x0123456789abcdef, 0xfedcba9876543210, 0}
	shiftCount4 := 64
	slice4, err = ShiftSliceOfUint64Left(slice4, shiftCount4)
	if err != nil {
		t.Errorf("ShiftSliceOfUint64Left failed: %v", err)
	}

	if !a.Equal(expected4, slice4) {
		t.Errorf("Test case 4 failed: Expected %X, but got %X", expected4, slice4)
		t.FailNow()
	}
}

func TestShiftSliceOfUint64Left(t *testing.T) {
	tests := []struct {
		name        string
		input       []uint64
		shiftCount  int
		expected    []uint64
		expectError bool
	}{
		{
			name:        "Test 1: Normal shift within range, no carry",
			input:       []uint64{0x1, 0x0, 0x0, 0x0},
			shiftCount:  1,
			expected:    []uint64{0x2, 0x0, 0x0, 0x0},
			expectError: false,
		},
		{
			name:        "Test 2: Normal shift within range, with carry",
			input:       []uint64{0x8000000000000000, 0x0, 0x0, 0x0},
			shiftCount:  1,
			expected:    []uint64{0x0, 0x1, 0x0, 0x0},
			expectError: false,
		},
		{
			name:        "Test 3: Shift beyond range",
			input:       []uint64{0x1, 0x2, 0x3, 0x4},
			shiftCount:  300,
			expected:    nil,
			expectError: true,
		},
		{
			name:        "Test 4: Zero shift",
			input:       []uint64{0x1, 0x2, 0x3, 0x4},
			shiftCount:  0,
			expected:    []uint64{0x1, 0x2, 0x3, 0x4},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ShiftSliceOfUint64Left(tt.input, tt.shiftCount)
			if (err != nil) != tt.expectError {
				t.Errorf("expected error: %v, got: %v", tt.expectError, err)
			}
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("expected: %v, got: %v", tt.expected, result)
			}
		})
	}
}
