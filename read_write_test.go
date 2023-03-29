package gobitstream_test

import (
	"github.com/juju/errors"
	"github.com/lagarciag/gobitstream"
	"math/rand"
	"pvsimflowtracking/tests"
	"testing"
)

func TestReadWrite1(t *testing.T) {
	t.Log(t.Name())
	_, a := tests.InitTest(t)

	const loops = 10000

	for i := 0; i < loops; i++ {
		sizeInBits := uint(rand.Intn(500) + 1)
		//sizeInBits := uint(322)
		in2 := tests.GenRandBytes(sizeInBits)
		rd, err := gobitstream.NewReaderLE(int(sizeInBits), in2)
		a.Nil(err)
		wr := gobitstream.NewWriterLE(int(sizeInBits))

		readBits := uint(rand.Intn(int(sizeInBits)))

		//readBits := uint(320) //239
		if readBits <= 64 {

			read, err := rd.ReadNbitsUint64(int(readBits))
			if readBits == 0 {
				a.NotNil(err)
			} else {
				if !a.Nil(err) {
					if readBits == 0 {
						continue
					}
					t.Errorf("reading bits: %d", readBits)
					t.Error(err.Error())
					t.Errorf(errors.ErrorStack(err))
					t.FailNow()
				}

				err = wr.WriteNbitsFromWord(int(readBits), read)
				if !a.Nil(err) {
					t.FailNow()
				}

				if err := wr.WriteBytes(); err != nil {
					t.FailNow()
				}

				if !a.Equal(read, wr.Uint64()) {
					t.Logf("read: %X", read)
					t.Logf("uint: %X,", wr.Uint64())
					t.FailNow()
				}
			}
		} else {
			read, err := rd.ReadNbitsBytes(int(readBits))
			if !a.Nil(err) {
				t.Errorf("reading bits: %d", readBits)
				t.Error(err.Error())
				t.FailNow()
			}

			err = wr.WriteNbitsFromBytes(int(readBits), read)
			if !a.Nil(err) {
				t.FailNow()
			}

			if err := wr.WriteBytes(); err != nil {
				t.FailNow()
			}

			if !a.Equal(read, wr.Bytes()) {
				t.Log("in   :", in2)
				t.Log("read :", read)
				t.Log("bytes:", wr.Bytes())
				t.Logf("words : %x", wr.Words())
				t.FailNow()
			}
		}

	}
}

func TestReadWrite2(t *testing.T) {
	t.Log(t.Name())
	_, a := tests.InitTest(t)

	const loops = 1000

	for i := 0; i < loops; i++ {
		sizeInBits := uint(rand.Intn(5000) + 1)
		readBits := uint(rand.Intn(int(sizeInBits)))

		if readBits > 64 {
			readBits = 64

		}
		in2 := tests.GenRandBytes(sizeInBits)

		rd, err := gobitstream.NewReaderLE(int(sizeInBits), in2)
		a.Nil(err)
		wr := gobitstream.NewWriterLE(int(sizeInBits))

		read, err := rd.ReadNbitsBytes(int(readBits))
		if readBits == 0 {
			a.NotNil(err)
		} else {
			if !a.Nil(err) {
				t.Errorf("reading bits: %d", readBits)
				t.Error(err.Error())
				t.FailNow()
			}
			err = wr.WriteNbitsFromBytes(int(readBits), read)
			if !a.Nil(err) {
				t.FailNow()
			}

			if err = wr.WriteBytes(); err != nil {
				t.FailNow()
			}

			a.Equal(read, wr.Bytes())

		}

	}
}
