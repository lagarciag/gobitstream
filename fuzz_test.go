package gobitstream_test

import (
	"github.com/juju/errors"
	"github.com/lagarciag/gobitstream"
	"pvsimflowtracking/tests"
	"testing"
)

func FuzzReadWrite(f *testing.F) {
	sizeInBits := uint(9)
	sizeInBytes := tests.SizeInBytes(sizeInBits)
	in2 := make([]byte, sizeInBytes)
	for i, _ := range in2 {
		in2[i] = 0xFF
	}
	readBits := 2

	f.Add(in2, sizeInBits, readBits)

	f.Fuzz(func(t *testing.T, in2 []byte, sizeInBits uint, readBits int) {
		_, a := tests.InitTest(t)
		rd, err := gobitstream.NewReaderLE(int(sizeInBits), in2)
		sizeInBytes = tests.SizeInBytes(sizeInBits)

		if len(in2) >= int(sizeInBytes) {
			if !a.Nil(err) {
				t.Errorf(err.Error())
				t.Log("size in bits:", sizeInBits)
				t.Log("sizeInBytes: ", sizeInBytes)
				t.Log("len in2", len(in2))
				t.Errorf(errors.ErrorStack(err))
			}
			wr := gobitstream.NewWriterLE(int(sizeInBits))

			if readBits <= 64 {

				read, err := rd.ReadNbitsUint64(int(readBits))
				if readBits == 0 || sizeInBits == 0 || int(sizeInBits) < readBits || readBits < 0 {
					a.NotNil(err)
				} else {
					if !a.Nil(err) {
						if readBits != 0 {
							t.Errorf("reading bits: %d", readBits)
							t.Errorf("size in bits: %d", sizeInBits)
							t.Error(err.Error())
							t.Errorf(errors.ErrorStack(err))
							t.FailNow()
						}
					}

					err = wr.WriteNbitsFromWord(int(readBits), read)
					if !a.Nil(err) {
						t.FailNow()
					}

					if err := wr.Flush(); err != nil {
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
				if readBits == 0 || sizeInBits == 0 || int(sizeInBits) < readBits {
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

					if err := wr.Flush(); err != nil {
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
		} else {
			if !a.NotNil(err) {
				t.Log("size in bits:", sizeInBits)
				t.Log("sizeInBytes: ", sizeInBytes)
				t.Log("len in2", len(in2))
				t.FailNow()
			}
		}

	})
}
