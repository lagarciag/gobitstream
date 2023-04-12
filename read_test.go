package gobitstream

import (
	"github.com/juju/errors"
	"github.com/lagarciag/gobitstream/tests"
	"math/big"
	"testing"
)

func TestExtractBitsFromSlice(t *testing.T) {
	_, a := tests.InitTest(t)
	slice := []uint64{0x0123456789abcdef, 0xfedcba9876543210}
	width := uint64(24)
	offset := uint64(16)
	expected := (slice[0] >> 16) & ((1 << width) - 1)

	actual, err := get64BitsFieldFromSlice(slice, width, offset)
	a.Nil(err)
	a.Equal(actual, expected)
}

func TestExtractBitsFromSlice2(t *testing.T) {
	_, a := tests.InitTest(t)
	slice := []uint64{0x0123456789abcdef, 0xfedcba9876543210}
	width := uint64(56)
	offset := uint64(16)

	expected := (slice[0] >> offset) & ((1 << width) - 1)
	remainingBits := width - (64 - offset)
	expected |= slice[1] & ((1 << remainingBits) - 1) << (64 - offset)

	t.Logf("Extract: %X", expected)
	actual, err := get64BitsFieldFromSlice(slice, width, offset)
	a.Nil(err)
	a.Equal(actual, expected)

	// Test an error for zero width
	width = uint64(0)
	_, err = get64BitsFieldFromSlice(slice, width, offset)
	a.NotNil(err)
	// Test an error for width greater than 64
	width = uint64(65)
	_, err = get64BitsFieldFromSlice(slice, width, offset)
	a.NotNil(err)

	// Test an error for out-of-range offset
	offset = uint64(128)
	_, err = get64BitsFieldFromSlice(slice, width, offset)
	a.NotNil(err)
}

func TestExtractBitsFromSliceGreater64(t *testing.T) {
	_, a := tests.InitTest(t)
	//slice := []uint64{0x0123456789abcdef, 0xfedcba9876543210, 0xfedcba9876543210}
	slice := []uint64{0xFFFFFFFFFFFFFFFF, 0xFFFFFFFFFFFFFFFF, 0xFFFFFFFFFFFFFFFF, 0xFFFFFFFFFFFFFFFF, 0xFFFFFFFFFFFFFFFF, 0xFFFFFFFFFFFFFFFF}
	resultBuffer := make([]uint64, 0, 7)
	width := uint64(134)
	offset := uint64(65)

	var x big.Int
	inBytes := make([]byte, len(slice)*8)

	for i := range inBytes {
		inBytes[i] = 0xFF
	}
	//t.Logf("inBytes : %x", inBytes)

	reverseSlice(inBytes)

	x.SetBytes(inBytes)
	y := x.Rsh(&x, uint(offset))
	mask := big.NewInt(1)
	mask = mask.Lsh(mask, uint(width))
	mask = mask.Add(mask, big.NewInt(-1))
	y = y.And(y, mask)

	compareBytes := y.Bytes()
	reverseSlice(compareBytes)
	size := sizeInWords(int(width))

	actual, err := getFieldFromSlice(resultBuffer, slice, width, offset)
	if !a.Nil(err) {
		t.Error(err.Error())
		t.Error(errors.ErrorStack(err))
		t.Logf("actual %X", actual)
	}

	if !a.Equal(size, len(actual)) {
		t.Logf("acutal   : %X -- %X", actual, actual[0])
		t.Logf("expected : %X", compareBytes)
		t.FailNow()
	}

}

func TestGetFieldFromSlice(t *testing.T) {
	_, a := tests.InitTest(t)
	slice := []uint64{0x1122334455667788, 0x99AABBCCDDEEFF00, 0x0123456789ABCDEF}
	var resultBuff []uint64
	width := uint64(40)
	offset := uint64(20)

	out, err := getFieldFromSlice(resultBuff, slice, width, offset)
	a.Nil(err)
	expected := []uint64{0x1223344556}
	a.Equal(expected, out)

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
	actual, err := get64BitsFieldFromSlice(slice, width, offset)
	a.Nil(err)
	a.Equal(actual, expected)

	err = set64BitsFieldToWordSlice(slice2, expected, width, offset)
	a.Nil(err)

	a.Equal(expected2, slice2)

	err = setFieldToSlice(slice3, []uint64{expected}, width, offset)
	a.Equal(expected2, slice3)
}
