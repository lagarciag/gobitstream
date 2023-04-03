package gobitstream

import (
	"encoding/binary"
	"github.com/juju/errors"
)

// Writer is a bit stream writer.
// It does not have io.Writer interface
type Writer struct {
	dst            []byte
	dstWord        []uint64
	currWordIndex  uint64
	currBitIndex   uint // MSB: 7, LSB: 0
	offset         int
	accOffset      int //
	size           int
	sizeInBytes    int
	sizeInWords    int
	isLittleEndian bool
}

func newWriter(totalBits int) *Writer {
	wr := &Writer{}
	wr.size = totalBits
	wr.sizeInWords = bitsToWordSize(totalBits)

	wr.dstWord = make([]uint64, wr.sizeInWords)
	wr.sizeInBytes = BitsToBytesSize(totalBits)
	return wr
}

// NewWriterLE creates a new Writer instance.
func NewWriterLE(totalBits int) *Writer {
	wr := newWriter(totalBits)
	wr.isLittleEndian = true
	return wr
}

func NewWriterBE(totalBits int) *Writer {
	wr := newWriter(totalBits)
	wr.isLittleEndian = false
	return wr
}

func (wr *Writer) Flush() error {
	sizeInBytes := len(wr.dstWord) * 8
	wr.dst = make([]byte, sizeInBytes)
	sizeInBytes = BitsToBytesSize(wr.accOffset)
	if wr.isLittleEndian {
		for i, word := range wr.dstWord {
			binary.LittleEndian.PutUint64(wr.dst[i*8:i*8+8], word)
		}

		wr.dst = wr.dst[0:sizeInBytes]
		return nil
	}
	for i, _ := range wr.dstWord {
		binary.BigEndian.PutUint64(wr.dst[i*8:i*8+8], wr.dstWord[len(wr.dstWord)-1-i])
	}
	wr.dst = wr.dst[len(wr.dst)-sizeInBytes:]
	return nil
}

func (wr *Writer) WriteNbitsFromBytes(nBits int, val []byte) (err error) {
	byteSize := BitsToBytesSize(nBits)

	if errx := checkByteSize(byteSize, len(val)); errx != nil {
		return errors.Trace(errx)
	}

	words, err := ConvertBytesToWords(nBits, val)
	nextByteSize := 0
	nextNBits := 0
	for i, _ := range words {
		if nextByteSize != 0 {
			byteSize = nextByteSize
			nBits = nextNBits
		}
		nextByteSize, nextNBits, byteSize, nBits = calcNextSizes(nBits, byteSize)
		if err = wr.WriteNbitsFromWord(nBits, words[i]); err != nil {
			return errors.Trace(err)
		}
		nBits = nextNBits
		byteSize = nextByteSize
	}

	return nil
}

func (wr *Writer) WriteNbitsFromWord(nBits int, val uint64) (err error) {
	if nBits > 64 {
		return InvalidBitsSizeError
	}
	mask, err := calcMask64(nBits)

	if err != nil {
		return errors.Trace(err)
	}

	if (nBits + wr.offset) <= 64 {
		wr.dstWord[wr.currWordIndex] |= (val & mask) << wr.offset
		wr.offset += nBits
		wr.accOffset += nBits
		if wr.offset >= 64 {
			wr.offset = wr.offset - 64
			wr.currWordIndex++
		}
	} else {
		remainingBits := (nBits + wr.offset) - 64
		newNbits := 64 - wr.offset
		remainingVal := val >> newNbits

		err = wr.WriteNbitsFromWord(newNbits, val)
		if err != nil {
			return errors.Trace(err)
		}

		err = wr.WriteNbitsFromWord(remainingBits, remainingVal)
		if err != nil {
			return errors.Trace(err)
		}
		//wr.currWordIndex++
	}

	return nil
}

func (wr *Writer) CurrentWord() []uint64 {
	return wr.dstWord
}

func (wr *Writer) Bytes() []byte {
	return wr.dst
}

func (wr *Writer) Words() []uint64 {
	out := wr.dstWord
	return out
}

func (wr *Writer) Uint64() uint64 {
	return wr.CurrentWord()[0]
}

// **************************************************
