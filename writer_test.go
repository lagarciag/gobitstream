package gobitstream_test

import (
	"github.com/juju/errors"
	"github.com/lagarciag/gobitstream"
	"github.com/stretchr/testify/assert"
	"os"
	"pvsimflowtracking/tests"
	"testing"
)

func TestMain(m *testing.M) {
	// call flag.Parse() here if TestMain uses flags
	os.Exit(m.Run())
}

func TestCopyCase(t *testing.T) {
	t.Log(t.Name())
	_, a := tests.InitTest(t)
	//e0
	in := []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x00}
	inFieldsBits := []int{5, 23, 8, 8, 64, 22, 1}
	//inFieldsBits := []int{5, 23, 8, 8, 64, 22, 1}
	inFieldsValues := make([]uint64, len(inFieldsBits))

	//t.Logf("in: %X, %X", in, in[0])

	countBits := 0
	for _, bits := range inFieldsBits {
		countBits += bits
	}
	//t.Logger("countBits:", countBits)

	r, err := gobitstream.NewReaderLE(131, in)
	if !a.Nil(err) {
		t.Errorf(err.Error())
		t.Errorf(errors.ErrorStack(err))
		t.FailNow()
	}

	for i, bits := range inFieldsBits {
		inFieldsValues[i], err = r.ReadNbitsUint64(bits)
		//t.Logf("field %d, width: %d: %X", i, bits, inFieldsValues[i])
		if !a.Nil(err) {
			t.Error("on step: ", i)
			t.Errorf(err.Error())
			t.Errorf(errors.ErrorStack(err))
			t.FailNow()
		}
	}

	w := gobitstream.NewWriterLE(int(131))
	//t.Logf("inFeildsValues: %X", inFieldsValues)
	//t.Logger("inFeildsBits:", inFieldsBits)
	for i, bits := range inFieldsBits {
		err = w.WriteNbitsFromWord(bits, inFieldsValues[i])
		if !a.Nil(err) {
			t.Errorf(err.Error())
			t.Errorf(errors.ErrorStack(err))
			t.FailNow()
		}
	}
	inBytes := gobitstream.BitsToBytesSize(countBits)
	w.Flush()
	t.Logf("in    :%X", in)
	t.Logf("output:%X , %d vrs %d", w.Bytes(), len(w.Bytes()), inBytes)
}

func TestOneByte(t *testing.T) {
	t.Log(t.Name())
	_, a := tests.InitTest(t)

	w := gobitstream.NewWriterBE(int(8))

	//writers

	err := w.WriteNbitsFromWord(8, uint64(0))
	a.Nil(err)

	w.Flush()
	theBytes := w.Bytes()

	t.Logf("%X", theBytes)
}

func TestRB(t *testing.T) {
	t.Log(t.Name())
	_, a := tests.InitTest(t)
	//in := bytes.NewBuffer([]byte{})
	wr := gobitstream.NewWriterLE(68)

	const firstWord = uint64(0xFFFFFFFF)

	err := wr.WriteNbitsFromWord(2, firstWord)
	if !a.Nil(err) {
		t.FailNow()
	}
	a.Equal([]uint64{0x3, 0}, wr.CurrentWord())

	err = wr.WriteNbitsFromWord(2, firstWord)
	if !a.Nil(err) {
		t.Logf("current word: 0x%x", wr.CurrentWord())
		t.FailNow()
	}

	err = wr.WriteNbitsFromWord(32, firstWord)

	if !a.Nil(err) {
		t.Logf("current word: 0x%x", wr.CurrentWord())
		t.FailNow()
	}
	err = wr.WriteNbitsFromWord(32, firstWord)
	if !a.Nil(err) {
		t.Logf("current word: 0x%x", wr.CurrentWord())
		t.FailNow()
	}

	err = wr.Flush()

	a.Equal([]uint64{0xffffffffffffffff, 0xF}, wr.CurrentWord())

}

func TestRB3(t *testing.T) {
	t.Log(t.Name())
	_, a := tests.InitTest(t)
	//in := bytes.NewBuffer([]byte{})
	wr := gobitstream.NewWriterLE(66 + 32)

	inBytes := []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}

	err := wr.WriteNbitsFromBytes(66, inBytes)
	if !a.Nil(err) {
		t.Error(errors.ErrorStack(err))
	}

	err = wr.WriteNbitsFromBytes(32, inBytes)
	if !a.Nil(err) {
		t.Error(errors.ErrorStack(err))
	}

	err = wr.Flush()
	a.Nil(err)

	t.Logf("current word: 0x%x", wr.CurrentWord())

	t.Logf("bytes: %x", wr.Bytes())
}

func TestConvertBytesToWords(t *testing.T) {
	t.Log(t.Name())
	_, a := tests.InitTest(t)

	tests := []struct {
		nBits int
		val   []byte
		want  []uint64
		err   error
	}{
		// Test case 1: nBits is 8, single byte input
		{
			nBits: 8,
			val:   []byte{0xAB},
			want:  []uint64{0xAB},
			err:   nil,
		},
		// Test case 2: nBits is 16, two bytes input
		{
			nBits: 16,
			val:   []byte{0xAB, 0xCD},
			want:  []uint64{0xCDAB},
			err:   nil,
		},
		// Test case 3: nBits is 24, three bytes input
		{
			nBits: 24,
			val:   []byte{0xAB, 0xCD, 0xEF},
			want:  []uint64{0xEFCDAB},
			err:   nil,
		},
		// Test case 4: nBits is 32, four bytes input
		{
			nBits: 32,
			val:   []byte{0xAB, 0xCD, 0xEF, 0x12},
			want:  []uint64{0x12EFCDAB},
			err:   nil,
		},
		// Test case 5: nBits is 64, eight bytes input
		{
			nBits: 64,
			val:   []byte{0xAB, 0xCD, 0xEF, 0x12, 0x34, 0x56, 0x78, 0x9A},
			want:  []uint64{0x9A78563412EFCDAB},
			err:   nil,
		},
		// Test case 6: nBits is 16, one byte input (error: insufficient bytes)
		{
			nBits: 16,
			val:   []byte{0xAB},
			want:  nil,
			err:   errors.New("insufficient bytes"),
		},
	}

	for _, tt := range tests {
		got, err := gobitstream.ConvertBytesToWords(tt.nBits, tt.val)

		if !a.Equal(got, tt.want) {
			t.Errorf("ConvertBytesToWords(%d, %v) = %v, want %v", tt.nBits, tt.val, got, tt.want)
		}
		if tt.err != nil {
			if !a.NotNil(err) {
				t.Errorf("ConvertBytesToWords(%d, %v) error = %v, want %v", tt.nBits, tt.val, err, tt.err)
			}
		} else {
			a.Nil(err)
		}
	}
}

func TestSetFieldToSlice(t *testing.T) {
	t.Run("Should set field to slice without error", func(t *testing.T) {
		dstSlice := []uint64{0, 0, 0}
		field := []uint64{6, 7, 8}
		width := uint64(192) // 64 bits for each field
		offset := uint64(0)

		err := gobitstream.SetFieldToSlice(dstSlice, field, width, offset)

		assert.NoError(t, err)
		expectedSlice := []uint64{6, 7, 8}
		assert.Equal(t, expectedSlice, dstSlice)
	})

	t.Run("Should return error if offset is out of range", func(t *testing.T) {
		dstSlice := []uint64{1, 2, 3}
		field := []uint64{6, 7, 8}
		width := uint64(192)  // 64 bits for each field
		offset := uint64(300) // out of range

		err := gobitstream.SetFieldToSlice(dstSlice, field, width, offset)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "is out of range")
	})

	t.Run("Should return error if width is larger than the size of the field", func(t *testing.T) {
		dstSlice := []uint64{1, 2, 3}
		field := []uint64{6, 7, 8}
		width := uint64(300) // out of range
		offset := uint64(0)

		err := gobitstream.SetFieldToSlice(dstSlice, field, width, offset)

		assert.Error(t, err)
		// Here, you should check if the error returned is the one you expect.
		// Since I do not have the exact implementation of the function Set64BitsFieldToWordSlice, I cannot provide the expected error.
	})

	t.Run("Should handle width and offset being zero appropriately", func(t *testing.T) {
		dstSlice := []uint64{1, 2, 3}
		field := []uint64{6, 7, 8}
		width := uint64(0)
		offset := uint64(0)

		err := gobitstream.SetFieldToSlice(dstSlice, field, width, offset)

		assert.NoError(t, err)
		expectedSlice := []uint64{1, 2, 3} // No changes since width is zero
		assert.Equal(t, expectedSlice, dstSlice)
	})
}
