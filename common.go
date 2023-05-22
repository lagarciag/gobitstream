package gobitstream

import (
	"encoding/binary"

	"github.com/juju/errors"
)

func totalOffsetToLocalOffset(offset int) (wordOffset, localOffset int) {
	return offset / 64, offset % 64
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

func reverseSlice[V uint64 | byte](s []V) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}

}

func sizeInWords(width int) (size int) {
	size = width / 64
	aMod := width % 64
	if aMod != 0 {
		size++
	}
	return size
}

func calcMask64(nBits int) (mask uint64, err error) {
	if nBits < 64 {
		mask = (uint64(uint64(1) << nBits)) - 1
		return mask, nil
	}

	if nBits == 64 {
		return 0xFFFFFFFFFFFFFFFF, nil
	}
	return 0, InvalidBitsSizeError
}

func ConvertBytesToWords(nBits int, val []byte) (words []uint64, err error) {
	wordSize := bitsToWordSize(nBits)
	byteSize := BitsToBytesSize(nBits)
	if errx := checkByteSize(byteSize, len(val)); errx != nil {
		return words, errors.Trace(errx)
	}

	words = make([]uint64, wordSize)
	nextByteSize := 0
	nextNBits := 0
	modShift := nBits % 64
	lastWordMask := uint64((1 << modShift) - 1)
	if nBits == 64 {
		lastWordMask = 0xFFFFFFFFFFFFFFFF
	}

	for i, _ := range words {
		if nextByteSize != 0 {
			byteSize = nextByteSize
			nBits = nextNBits
		}
		nextByteSize, nextNBits, byteSize, nBits = calcNextSizes(nBits, byteSize)
		if byteSize == 1 {
			if errx := writeBytesCase(i, 1, byteSize, words, val); err != nil {
				return words, errors.Trace(errx)
			}
		} else if byteSize <= 2 {
			if errx := writeBytesCase(i, 2, byteSize, words, val); err != nil {
				return words, errors.Trace(errx)
			}
		} else if byteSize <= 4 {
			if errx := writeBytesCase(i, 4, byteSize, words, val); err != nil {
				return words, errors.Trace(errx)
			}
		} else if byteSize <= 8 {
			if errx := writeBytesCase(i, 8, byteSize, words, val); err != nil {
				return words, errors.Trace(errx)
			}
		} else {
			return words, errors.Trace(UnexpectedCondition)
		}
		if nextByteSize < len(val) {
			val = val[byteSize:]
			nBits = nextNBits
			byteSize = nextByteSize
		}
	}

	//fmt.Printf("lastwordMask: %X -- %d\n", lastWordMask, modShift)
	if lastWordMask != 0 {
		words[len(words)-1] &= lastWordMask
	}
	return words, nil
}

func BitsToBytesSize(in int) int {
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

func checkByteSize(byteSize, valSize int) (err error) {
	if byteSize > valSize {
		err = InvalidInputSliceSizeError
		err = errors.Annotatef(err, "wanted bytes: %d, input slice sizeInBytes: %d", byteSize, valSize)
		return errors.Trace(err)
	}
	return nil
}

func writeBytesCase(i, theCase, byteSize int, words []uint64, val []byte) (err error) {
	newVal := val
	if byteSize <= theCase {
		missing := theCase - byteSize
		newVal = append(newVal, make([]byte, missing)...)
	}
	if len(newVal) == 0 {
		return nil
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
		return errors.Trace(InvalidBitsSizeError)
	}

	return nil
}
