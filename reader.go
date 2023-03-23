package gobitstream

import (
	"encoding/binary"
	"github.com/juju/errors"
)

// Reader is a bit stream Reader.
// It does not have io.Reader interface
type Reader struct {
	dst           []byte
	currWordIndex uint64
	dstWord       []uint64
	currBitIndex  uint // MSB: 7, LSB: 0
	offset        int
	accOffset     int //
}

// NewReader creates a new Reader instance.
func NewReader(totalBits int) *Reader {
	wr := &Reader{}
	totalWords := wr.bitsToWordSize(totalBits)
	wr.dstWord = make([]uint64, totalWords)
	return wr
}

func (wr *Reader) WriteBytes() error {
	sizeInBytes := len(wr.dstWord) * 8
	wr.dst = make([]byte, sizeInBytes)
	for i, word := range wr.dstWord {
		binary.LittleEndian.PutUint64(wr.dst[i*8:i*8+8], word)
	}
	sizeInBytes = wr.bitsToBytesSize(wr.accOffset)
	wr.dst = wr.dst[0:sizeInBytes]

	return nil
}

func (wr *Reader) WriteNbitsFromBytes(nBits int, val []byte) (err error) {
	wordSize := wr.bitsToWordSize(nBits)
	byteSize := wr.bitsToBytesSize(nBits)
	if byteSize > len(val) {
		err = InvalidInputSliceSize
		err = errors.Annotatef(err, "wanted bytes: %d, input slice size: %d", byteSize, len(val))
		return errors.Trace(err)
	}

	words := make([]uint64, wordSize)
	nextByteSize := 0
	nextNBits := 0
	for i, _ := range words {
		if nextByteSize != 0 {
			byteSize = nextByteSize
			nBits = nextNBits
		}
		nextByteSize, nextNBits, byteSize, nBits = calcNextSizes(nBits, byteSize)
		if byteSize == 1 {
			if errx := writeBytesToWordsCases(i, 1, byteSize, nBits, nextByteSize, words, val); err != nil {
				return errors.Trace(errx)
			}
		} else if byteSize <= 2 {
			if errx := writeBytesToWordsCases(i, 2, byteSize, nBits, nextByteSize, words, val); err != nil {
				return errors.Trace(errx)
			}
		} else if byteSize <= 4 {
			if errx := writeBytesToWordsCases(i, 4, byteSize, nBits, nextByteSize, words, val); err != nil {
				return errors.Trace(errx)
			}
		} else if byteSize <= 8 {
			if errx := writeBytesToWordsCases(i, 8, byteSize, nBits, nextByteSize, words, val); err != nil {
				return errors.Trace(errx)
			}
		} else {
			return errors.Trace(UnexpectedCondition)
		}
		if err = wr.WriteNbitsOfWord(nBits, words[i]); err != nil {
			return errors.Trace(err)
		}
		nBits = nextNBits
		byteSize = nextByteSize
	}

	return nil
}

func writeBytesToWordsCases(i, theCase, byteSize, nBits, nextByteSize int, words []uint64, val []byte) (err error) {
	newVal := val
	if byteSize < theCase {
		missing := theCase - byteSize
		newVal = append(newVal, make([]byte, missing)...)
	}

	if byteSize == 1 {
		words[i] = uint64(newVal[0])
	} else if byteSize <= 2 {
		words[i] = uint64(binary.LittleEndian.Uint16(newVal[0:2]))
	} else if byteSize <= 4 {
		words[i] = uint64(binary.LittleEndian.Uint32(newVal[0:4]))
	} else if byteSize <= 8 {
		words[i] = binary.LittleEndian.Uint64(newVal[0:8])
	} else {
		return errors.Trace(InvalidBitsSize)
	}
	val = val[nextByteSize:]

	return nil
}

func (wr *Reader) WriteNbitsOfWord(nBits int, val uint64) (err error) {
	if nBits > 64 {
		return InvalidBitsSize
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

		err = wr.WriteNbitsOfWord(newNbits, val)
		if err != nil {
			return errors.Trace(err)
		}

		err = wr.WriteNbitsOfWord(remainingBits, remainingVal)
		if err != nil {
			return errors.Trace(err)
		}
		//wr.currWordIndex++
	}

	return nil
}

func (wr *Reader) CurrentWord() []uint64 {
	return wr.dstWord
}

func (wr *Reader) Bytes() []byte {
	return wr.dst
}

// **************************************************

func (wr *Reader) bitsToBytesSize(in int) int {
	size := in / 8
	if in%8 != 0 {
		size++
	}
	return size
}

func (wr *Reader) bitsToWordSize(in int) int {
	size := in / 64
	if in%64 != 0 {
		size++
	}
	return size
}
