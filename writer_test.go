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

func TestRB(t *testing.T) {
	t.Log(t.Name())
	_, a := tests.InitTest(t)
	//dst := bytes.NewBuffer([]byte{})
	wr := gobitstream.NewWriter(68)

	const firstWord = uint64(0xFFFFFFFF)
	const secondWord = uint64(0xFF)

	err := wr.WriteNbitsOfWord(2, firstWord)
	if !a.Nil(err) {
		t.FailNow()
	}
	a.Equal([]uint64{0x3, 0}, wr.CurrentWord())

	err = wr.WriteNbitsOfWord(2, firstWord)
	if !a.Nil(err) {
		t.Logf("current word: 0x%x", wr.CurrentWord())
		t.FailNow()
	}

	err = wr.WriteNbitsOfWord(32, firstWord)

	if !a.Nil(err) {
		t.Logf("current word: 0x%x", wr.CurrentWord())
		t.FailNow()
	}
	err = wr.WriteNbitsOfWord(32, firstWord)
	if !a.Nil(err) {
		t.Logf("current word: 0x%x", wr.CurrentWord())
		t.FailNow()
	}

	err = wr.WriteBytes()

	a.Equal([]uint64{0xffffffffffffffff, 0xF}, wr.CurrentWord())

	t.Logf("current word: 0x%x", wr.CurrentWord())

	t.Logf("bytes: %x", wr.Bytes())

}

func TestRB3(t *testing.T) {
	t.Log(t.Name())
	_, a := tests.InitTest(t)
	//dst := bytes.NewBuffer([]byte{})
	wr := gobitstream.NewWriter(66 + 32)

	inBytes := []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}

	err := wr.WriteNbitsFromBytes(66, inBytes)
	if !a.Nil(err) {
		t.Error(errors.ErrorStack(err))
	}

	err = wr.WriteNbitsFromBytes(32, inBytes)
	if !a.Nil(err) {
		t.Error(errors.ErrorStack(err))
	}

	err = wr.WriteBytes()
	a.Nil(err)

	t.Logf("current word: 0x%x", wr.CurrentWord())

	t.Logf("bytes: %x", wr.Bytes())
}
