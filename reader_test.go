package gobitstream_test

import (
	"encoding/binary"
	"github.com/juju/errors"
	"github.com/lagarciag/gobitstream"
	"math/rand"
	"pvsimflowtracking/tests"
	"testing"
)

func TestReaderBasic(t *testing.T) {
	t.Log(t.Name())
	_, a := tests.InitTest(t)
	//in := bytes.NewBuffer([]byte{})

	const sizeInBits = 23
	sizeInBytes, mask := tests.SizeAndLastByteMaskLE(sizeInBits)
	in := make([]byte, sizeInBytes)

	for i, _ := range in {
		in[i] = 0xFF
	}

	tests.MaskLastByteLE(mask, in)

	t.Logf("in: %x", in)

	wr, err := gobitstream.NewReaderLE(sizeInBits, in)

	a.Nil(err)
	a.NotNil(wr)

	t.Logf("words: %x", wr.Words())

}

func TestBytesToWords32(t *testing.T) {
	t.Log(t.Name())
	_, a := tests.InitTest(t)

	const sizeInBits = 32
	sizeInBytes, mask := tests.SizeAndLastByteMaskLE(sizeInBits)

	t.Logf("mask: %x", mask)
	t.Logf("sizeInBytes in bytes: %d", sizeInBytes)

	in := make([]byte, sizeInBytes)

	const inInt = 0x0A0B0C0D

	binary.LittleEndian.PutUint32(in, uint32(inInt))

	//tests.MaskLastByteLE(mask, in)

	t.Logf("in: %x", in)

	words, err := gobitstream.ConvertBytesToWords(sizeInBits, in)

	a.Nil(err)

	t.Logf("words: %x", words)

	a.Equal(uint64(inInt), words[0])

}

func TestBytesToWordsRand(t *testing.T) {
	t.Log(t.Name())
	_, a := tests.InitTest(t)

	const loops = 1000

	for i := 0; i < loops; i++ {

		sizeInBits := uint64(rand.Intn(64))
		//sizeInBits := uint64(61)
		if sizeInBits == 0 {
			continue
		}

		sizeInBytes, mask := tests.SizeAndLastByteMaskLE(uint(sizeInBits))

		//t.Logf("mask: %v, sizeInBytes: %d", mask, sizeInBytes)

		in := make([]byte, sizeInBytes)

		for i, _ := range in {
			in[i] = 0xFF
		}
		tests.MaskLastByteLE(mask, in)

		//t.Logf("in %X bytes", in)
		in2 := in
		for i := 0; i < int(8-sizeInBytes); i++ {
			in2 = append(in2, 0x00)
		}

		verify := binary.LittleEndian.Uint64(in2)

		words, err := gobitstream.ConvertBytesToWords(int(sizeInBits), in)

		a.Nil(err)

		if !(a.Equal(uint64(verify), words[0])) {
			t.Log("sizeInBits: ", sizeInBits)
		}
	}
}

func TestBytesToWordsRand2(t *testing.T) {
	t.Log(t.Name())
	_, a := tests.InitTest(t)

	const loops = 1000

	for i := 0; i < loops; i++ {
		sizeInBits := uint(rand.Intn(63) + 1)
		in2, verify := tests.GenRandBytes64Bits(sizeInBits)
		words, err := gobitstream.ConvertBytesToWords(int(sizeInBits), in2)
		a.Nil(err)
		a.Equal(verify, words[0])
	}
}

func TestBytesToWordsRand3(t *testing.T) {
	t.Log(t.Name())
	_, a := tests.InitTest(t)

	const loops = 100

	for i := 0; i < loops; i++ {
		sizeInBits := uint(rand.Intn(500) + 1)
		sizeInWords := tests.SizeInWords(sizeInBits)
		in2 := tests.GenRandBytes(sizeInBits)

		wr, err := gobitstream.NewReaderLE(int(sizeInBits), in2)

		a.Nil(err)

		words := wr.Words()

		a.Equal(int(sizeInWords), len(words))

		//a.Equal(verify, wr.Words()[0])
	}
}

func TestBytesToWordsRand0(t *testing.T) {
	t.Log(t.Name())
	_, a := tests.InitTest(t)

	const loops = 100

	for i := 0; i < loops; i++ {
		sizeInBits := uint(rand.Intn(500) + 1)

		read1 := rand.Intn(int(sizeInBits))
		if sizeInBits > 64 {
			read1 = rand.Intn(64)
		}
		if read1 == 0 {
			continue
		}
		sizeInWords := tests.SizeInWords(sizeInBits)
		in2 := tests.GenRandBytes(sizeInBits)

		wr, err := gobitstream.NewReaderLE(int(sizeInBits), in2)

		a.Nil(err)

		words := wr.Words()

		a.Equal(int(sizeInWords), len(words))

		_, err = wr.ReadNbitsUint64(read1)
		if !a.Nil(err) {
			t.Errorf("sizeInBits: %d", sizeInBits)
			t.Error(err.Error())
			t.Error(errors.ErrorStack(err))
			t.FailNow()
		}
	}
}

func TestBytesToWordsRand4(t *testing.T) {
	t.Log(t.Name())
	_, a := tests.InitTest(t)

	const loops = 1000

	for i := 0; i < loops; i++ {
		sizeInBits := uint(rand.Intn(500) + 1)

		read1 := rand.Intn(int(sizeInBits))
		if sizeInBits > 64 {
			read1 = rand.Intn(64)
		}
		if read1 == 0 {
			continue
		}
		read2 := rand.Intn(read1)
		sizeInWords := tests.SizeInWords(sizeInBits)
		in2 := tests.GenRandBytes(sizeInBits)

		wr, err := gobitstream.NewReaderLE(int(sizeInBits), in2)

		a.Nil(err)

		words := wr.Words()

		a.Equal(int(sizeInWords), len(words))

		_, err = wr.ReadNbitsUint64(read1)
		if !a.Nil(err) {
			t.Error(err.Error(), read1)
		}

		_, err = wr.ReadNbitsUint64(read2)
		if sizeInBits < uint(read1+read2) {
			if !a.NotNil(err) {
				t.Logf("read1 + read2 = %d -- %d", read1+read2, sizeInBits)
				t.FailNow()
			}

		} else {
			if read2 == 0 {
				a.NotNil(err)
			} else if !a.Nil(err) {
				t.Logf("read1 + read2 = %d -- %d", read1+read2, sizeInBits)
				t.Error(err.Error(), read2)
				t.FailNow()
			}

		}

	}
}

func TestSimpleCase1(t *testing.T) {
	t.Log(t.Name())
	_, a := tests.InitTest(t)

	const loops = 1000

	for i := 0; i < loops; i++ {
		sizeInBits := uint(rand.Intn(500) + 1)
		sizeInWords := tests.SizeInWords(sizeInBits)
		in2 := tests.GenRandBytes(sizeInBits)

		wr, err := gobitstream.NewReaderLE(int(sizeInBits), in2)

		a.Nil(err)

		words := wr.Words()

		a.Equal(int(sizeInWords), len(words))

		read, err := wr.ReadNbitsBytes(12)
		if sizeInBits >= 12 {
			if !a.Nil(err) {
				t.Errorf(err.Error())
				t.Errorf(errors.ErrorStack(err))
			}
		} else {
			if !a.NotNil(err) {
				t.Errorf("sizeInBits: %d", sizeInBits)
				t.Errorf("read : %X ", read)
				t.FailNow()
			}

		}

	}
}

func TestSimpleBE(t *testing.T) {
	t.Log(t.Name())
	_, a := tests.InitTest(t)

	const loops = 1000

	for i := 0; i < loops; i++ {
		sizeInBits := uint(rand.Intn(500) + 1)
		//sizeInBits := uint(161)
		readSize := uint(rand.Intn(int(sizeInBits)))
		//readSize := 47
		in0 := tests.GenRandBytes(sizeInBits)
		mask := tests.LastByteMaskLE(sizeInBits)
		tests.MaskLastByteLE(mask, in0)
		//t.Logf("in0: %x", in0)

		wr, err := gobitstream.NewReaderLE(int(sizeInBits), in0)

		a.Nil(err)

		read, err := wr.ReadNbitsBytes(int(readSize))

		if readSize != 0 {
			if !a.Nil(err) {
				t.Error(errors.ErrorStack(err))
				t.Error(err.Error())
				t.FailNow()
			}
			in1 := make([]byte, len(in0))
			_ = copy(in1, in0)
			reverseSlice(in1)
			wr2, err2 := gobitstream.NewReaderBE(int(sizeInBits), in1)
			if !a.Nil(err2) {
				t.Error(errors.ErrorStack(err))
				t.Error(err2.Error())
				t.FailNow()
			}

			read2, err2 := wr2.ReadNbitsBytes(int(readSize))
			if !a.Nil(err) {
				t.Error(errors.ErrorStack(err))
				t.Error(err2.Error())
				t.FailNow()
			}

			reverseSlice(read)

			if !(a.Equal(read, read2)) {
				t.Logf("readSize: %d", readSize)
				t.Logf("sizeInBits: %d ", sizeInBits)
				t.Log("read2: ", read2)
				t.Errorf("not equal")
				t.Errorf("readSize=%d", readSize)
				t.Errorf("read : %X", read)
				t.Errorf("read2: %X", read2)
				t.FailNow()
			}
		} else {
			if !a.NotNil(err) {
				t.Error(err.Error())
				t.FailNow()
			}
		}
	}
}

func BenchmarkIntMin(b *testing.B) {
	sizeInBits := uint(3000)
	in2 := tests.GenRandBytes(sizeInBits)
	wr, _ := gobitstream.NewReaderLE(int(sizeInBits), in2)
	for i := 0; i < b.N; i++ {
		_, err := wr.ReadNbitsBytes(500)
		if err != nil {
			b.Error(err.Error())
			b.FailNow()
		}
		wr.Reset()

	}
}

func reverseSlice(s []byte) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}

}
