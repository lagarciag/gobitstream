package gobitstream_test

import (
	"github.com/juju/errors"
	"github.com/lagarciag/gobitstream"
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

	t.Logf("in: %X, %X", in, in[0])

	countBits := 0
	for _, bits := range inFieldsBits {
		countBits += bits
	}
	t.Log("countBits:", countBits)

	r, err := gobitstream.NewReaderLE(131, in)
	if !a.Nil(err) {
		t.Errorf(err.Error())
		t.Errorf(errors.ErrorStack(err))
		t.FailNow()
	}

	for i, bits := range inFieldsBits {
		inFieldsValues[i], err = r.ReadNbitsUint64(bits)
		t.Logf("field %d, width: %d: %X", i, bits, inFieldsValues[i])
		if !a.Nil(err) {
			t.Error("on step: ", i)
			t.Errorf(err.Error())
			t.Errorf(errors.ErrorStack(err))
			t.FailNow()
		}
	}

	w := gobitstream.NewWriterLE(int(131))
	t.Logf("inFeildsValues: %X", inFieldsValues)
	t.Log("inFeildsBits:", inFieldsBits)
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
	const secondWord = uint64(0xFF)

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

	t.Logf("current word: 0x%x", wr.CurrentWord())

	t.Logf("bytes: %x", wr.Bytes())

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
