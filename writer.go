package gobitstream

import (
	"encoding/binary"
	"github.com/juju/errors"
)

// Writer is a bit stream writer.
// It does not have io.Writer interface
type Writer struct {
	dst           []byte
	currWordIndex uint64
	dstWord       []uint64
	currBitIndex  uint // MSB: 7, LSB: 0
	offset        int
	accOffset     int //
}

// NewWriter creates a new Writer instance.
func NewWriter(totalBits int) *Writer {
	wr := &Writer{}
	totalWords := bitsToWordSize(totalBits)
	wr.dstWord = make([]uint64, totalWords)
	return wr
}

func calcMask64(nBits int) (mask uint64, err error) {
	if nBits < 64 {
		mask = (uint64(uint64(1) << nBits)) - 1
		return mask, nil
	}

	if nBits == 64 {
		return 0xFFFFFFFFFFFFFFFF, nil
	}
	return 0, InvalidBitsSize
}

func (wr *Writer) WriteBytes() error {
	sizeInBytes := len(wr.dstWord) * 8
	wr.dst = make([]byte, sizeInBytes)
	for i, word := range wr.dstWord {
		binary.LittleEndian.PutUint64(wr.dst[i*8:i*8+8], word)
	}
	sizeInBytes = bitsToBytesSize(wr.accOffset)
	wr.dst = wr.dst[0:sizeInBytes]

	return nil
}

func calcNextSizes(nBits, byteSize int) (nextByteSize, nextNBits, newByteSize, newNBits int) {
	if byteSize > 8 {
		nextByteSize = byteSize - 8
		nextNBits = nBits - 64
		byteSize = 8
		nBits = 64
	}
	return nextByteSize, nextNBits, byteSize, nBits
}

func (wr *Writer) ConvertBytesToWords(nBits int, val []byte) (words []uint64, err error) {
	wordSize := bitsToWordSize(nBits)
	byteSize := bitsToBytesSize(nBits)
	if byteSize > len(val) {
		err = InvalidInputSliceSize
		err = errors.Annotatef(err, "wanted bytes: %d, input slice size: %d", byteSize, len(val))
		return words, errors.Trace(err)
	}
	words = make([]uint64, wordSize)
	//nextByteSize := 0
	//nextNBits := 0

	return words, nil
}

func (wr *Writer) WriteNbitsFromBytes(nBits int, val []byte) (err error) {
	wordSize := bitsToWordSize(nBits)
	byteSize := bitsToBytesSize(nBits)
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
			if errx := writeBytesCase(i, 1, byteSize, words, val); err != nil {
				return errors.Trace(errx)
			}
		} else if byteSize <= 2 {
			if errx := writeBytesCase(i, 2, byteSize, words, val); err != nil {
				return errors.Trace(errx)
			}
		} else if byteSize <= 4 {
			if errx := writeBytesCase(i, 4, byteSize, words, val); err != nil {
				return errors.Trace(errx)
			}
		} else if byteSize <= 8 {
			if errx := writeBytesCase(i, 8, byteSize, words, val); err != nil {
				return errors.Trace(errx)
			}
		} else {
			return errors.Trace(UnexpectedCondition)
		}
		val = val[nextByteSize:]
		if err = wr.WriteNbitsOfWord(nBits, words[i]); err != nil {
			return errors.Trace(err)
		}

		nBits = nextNBits
		byteSize = nextByteSize
	}

	return nil
}

func writeBytesCase(i, theCase, byteSize int, words []uint64, val []byte) (err error) {
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

	return nil
}

func (wr *Writer) WriteNbitsOfWord(nBits int, val uint64) (err error) {
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

func (wr *Writer) CurrentWord() []uint64 {
	return wr.dstWord
}

func (wr *Writer) Bytes() []byte {
	return wr.dst
}

// **************************************************

func bitsToBytesSize(in int) int {
	size := in / 8
	if in%8 != 0 {
		size++
	}
	return size
}

func bitsToWordSize(in int) int {
	size := in / 64
	if in%64 != 0 {
		size++
	}
	return size
}
