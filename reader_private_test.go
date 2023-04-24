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
	ShiftSliceOfUint64Left(slice3, shiftCount3)
	if !a.Equal(slice3, expected3) {
		t.Errorf("Test case 3 failed: Expected %X, but got %X", expected3, slice3)
		t.FailNow()
	}

	// Test case 4: Shift by 64 bits (full rotation)
	slice4 := []uint64{0x0123456789abcdef, 0xfedcba9876543210}
	expected4 := []uint64{0x0123456789abcdef, 0xfedcba9876543210}
	shiftCount4 := 64
	ShiftSliceOfUint64Left(slice4, shiftCount4)
	if !a.Equal(slice4, expected4) {
		t.Errorf("Test case 4 failed: Expected %X, but got %X", expected4, slice4)
		t.FailNow()
	}
}

func TestShiftSliceOfUint64Left(t *testing.T) {
	_, a := commontest.InitTestLogToStdout(t)
	// Test case 1: Empty input slice
	slice1 := []uint64{}
	shiftedSlice1 := ShiftSliceOfUint64Left(slice1, 3)
	expectedSlice1 := []uint64{}
	if !reflect.DeepEqual(shiftedSlice1, expectedSlice1) {
		t.Errorf("Test case 1 failed. Expected %v, got %v", expectedSlice1, shiftedSlice1)
	}

	// Test case 2: Shift count within slice length
	slice2 := []uint64{1, 2, 3}
	shiftedSlice2 := ShiftSliceOfUint64Left(slice2, 3)
	expectedSlice2 := []uint64{8, 16, 24}
	if !reflect.DeepEqual(shiftedSlice2, expectedSlice2) {
		t.Errorf("Test case 2 failed. Expected %v, got %v", expectedSlice2, shiftedSlice2)
	}

	// Test case 3: Shift count larger than slice length
	slice3 := []uint64{1, 2, 3}
	shiftedSlice3 := ShiftSliceOfUint64Left(slice3, 70)
	expectedSlice3 := []uint64{64, 128, 192}
	if !a.Equal(expectedSlice3, shiftedSlice3) {
		t.Errorf("Test case 3 failed. Expected %v, got %v", expectedSlice3, shiftedSlice3)
		t.FailNow()
	}

	//// Test case 4: Shift count is negative
	//slice4 := []uint64{1, 2, 3}
	//shiftedSlice4 := ShiftSliceOfUint64Left(slice4, -3)
	//expectedSlice4 := []uint64{4, 8, 12}
	//if !reflect.DeepEqual(shiftedSlice4, expectedSlice4) {
	//	t.Errorf("Test case 4 failed. Expected %v, got %v", expectedSlice4, shiftedSlice4)
	//}

	// Test case 5: Input slice with non-zero initial values
	slice5 := []uint64{0xffffffffffffffff, 0xffffffffffffffff}
	shiftedSlice5 := ShiftSliceOfUint64Left(slice5, 16)
	expectedSlice5 := []uint64{0xffffffffffff0000, 0xffffffffffffffff, 0xffff}
	if !reflect.DeepEqual(shiftedSlice5, expectedSlice5) {
		t.Errorf("Test case 5 failed. Expected %v, got %v", expectedSlice5, shiftedSlice5)
	}
}
