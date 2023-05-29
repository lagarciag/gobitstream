package gobitstream

import (
	"github.com/juju/errors"
	"github.com/lagarciag/gobitstream/tests"
	"math/big"
	"testing"
)

func TestExtractBitsFromSlice(t *testing.T) {
	_, a, _ := tests.InitTest(t)
	slice := []uint64{0x0123456789abcdef, 0xfedcba9876543210}
	width := uint64(24)
	offset := uint64(16)
	expected := (slice[0] >> 16) & ((1 << width) - 1)

	actual, err := Get64BitsFieldFromSlice(slice, width, offset)
	a.Nil(err)
	a.Equal(actual, expected)
}

func TestExtractBitsFromSlice2(t *testing.T) {
	_, a, _ := tests.InitTest(t)
	slice := []uint64{0x0123456789abcdef, 0xfedcba9876543210}
	width := uint64(56)
	offset := uint64(16)

	expected := (slice[0] >> offset) & ((1 << width) - 1)
	remainingBits := width - (64 - offset)
	expected |= slice[1] & ((1 << remainingBits) - 1) << (64 - offset)

	t.Logf("Extract: %X", expected)
	actual, err := Get64BitsFieldFromSlice(slice, width, offset)
	a.Nil(err)
	a.Equal(actual, expected)

	// Test an error for zero width
	width = uint64(0)
	_, err = Get64BitsFieldFromSlice(slice, width, offset)
	a.NotNil(err)
	// Test an error for width greater than 64
	width = uint64(65)
	_, err = Get64BitsFieldFromSlice(slice, width, offset)
	a.NotNil(err)

	// Test an error for out-of-range offset
	offset = uint64(128)
	_, err = Get64BitsFieldFromSlice(slice, width, offset)
	a.NotNil(err)
}

func TestExtractBitsFromSliceGreater64(t *testing.T) {
	_, a, _ := tests.InitTest(t)
	//slice := []uint64{0x0123456789abcdef, 0xfedcba9876543210, 0xfedcba9876543210}
	slice := []uint64{0xFFFFFFFFFFFFFFFF, 0xFFFFFFFFFFFFFFFF, 0xFFFFFFFFFFFFFFFF, 0xFFFFFFFFFFFFFFFF, 0xFFFFFFFFFFFFFFFF, 0xFFFFFFFFFFFFFFFF}
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

	actual, err := GetAnySizeFieldFromUint64Slice(slice, width, offset)
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
	_, a, _ := tests.InitTest(t)

	var slice []uint64
	var width uint64
	var offset uint64

	slice = []uint64{0x1122334455667788, 0x99AABBCCDDEEFF00, 0x0123456789ABCDEF}
	width = uint64(64)
	offset = uint64(64)

	out, err := GetAnySizeFieldFromUint64Slice(slice, width, offset)
	a.Nil(err)
	expected := []uint64{0x99AABBCCDDEEFF00}
	a.Equal(expected, out)

	slice = []uint64{0x1122334455667788, 0x99AABBCCDDEEFF00, 0x0123456789ABCDEF}
	width = uint64(32)
	//offset = uint64(0)
	//
	//t.Logf("Extract: %X", slice[0])
	//
	//out, err = GetAnySizeFieldFromUint64Slice(slice, width, offset)
	//a.Nil(err)
	//expected = []uint64{0x55667788}
	//a.Equal(expected, out)
	//
	//slice = []uint64{0x1111111122222222, 0x3333333344444444, 0x0123456789ABCDEF}
	//width = uint64(32)
	//offset = uint64(0)
	//
	//t.Logf("Extract: %X", slice[0])
	//
	//out, err = GetAnySizeFieldFromUint64Slice(slice, width, offset)
	//a.Nil(err)
	//expected = []uint64{0x22222222}
	//a.Equal(expected, out)

	slice = []uint64{0x1111111122222222, 0x3333333344444444, 0x0123456789ABCDEF}
	width = uint64(64)
	offset = uint64(32)

	t.Logf("Extract: %X", slice[0])

	out, err = GetAnySizeFieldFromUint64Slice(slice, width, offset)
	a.Nil(err)
	expected = []uint64{0x4444444411111111}
	a.Equal(expected, out)

	slice = []uint64{0x1122334455667788, 0x99AABBCCDDEEFF00, 0x0123456789ABCDEF}
	width = uint64(40)
	offset = uint64(20)

	out, err = GetAnySizeFieldFromUint64Slice(slice, width, offset)
	a.Nil(err)
	expected = []uint64{0x1223344556}
	a.Equal(expected, out)

}
