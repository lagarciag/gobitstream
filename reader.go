package gobitstream

import (
	"encoding/binary"
	"github.com/juju/errors"
)

// Reader is a bit stream Reader.
// It does not have io.Reader interface
type Reader struct {
	inWord         []uint64
	currWordIndex  int
	currBitIndex   int // MSB: 7, LSB: 0
	offset         int
	in             []byte
	isLittleEndian bool
	size           int
	accOffset      int
}

// NewReaderLE creates a new Reader instance.
func NewReaderLE(sizeInBits int, in []byte) (wr *Reader, err error) {
	wr = &Reader{}
	wr.size = sizeInBits
	wr.isLittleEndian = true
	wr.in = in
	wr.inWord, err = ConvertBytesToWords(sizeInBits, in)
	wr.currWordIndex = len(wr.inWord) - 1
	return wr, err
}

func NewReaderBE(sizeInBits int, in []byte) (wr *Reader, err error) {
	wr = &Reader{}
	wr.isLittleEndian = false
	wr.in = in
	wr.inWord, err = ConvertBytesToWords(sizeInBits, in)
	wr.currWordIndex = 0
	return wr, err
}

func (wr *Reader) checkNbitsSize(nBits int) error {
	if nBits == 0 {
		err := InvalidBitsSizeError
		err = errors.Annotate(err, "nBits cannot be 0")
		return errors.Trace(err)
	} else if nBits+wr.accOffset > wr.size {
		err := InvalidBitsSizeError
		err = errors.Annotatef(err, "nBits: %d, accOffset: %d, sizeInBytes: %d", nBits, wr.accOffset, wr.size)
		return errors.Trace(err)
	}

	return nil
}

func (wr *Reader) calcParams(nBits, offset int) (sizeInBytes, wordOffset, localOffset, nextOffset, nextWordIndex, nextLocalOffset int) {
	sizeInBytes = BitsToBytesSize(nBits)
	nextOffset = wr.offset + nBits
	nextWordIndex = wr.currWordIndex

	wordOffset, localOffset = totalOffsetToLocalOffset(offset)
	_, nextLocalOffset = totalOffsetToLocalOffset(nextOffset)

	if wr.isLittleEndian {
		nextWordIndex = wordOffset - 1
	} else {
		nextWordIndex = wordOffset + 1
	}

	return sizeInBytes, wordOffset, localOffset, nextOffset, nextWordIndex, nextLocalOffset
}

func (wr *Reader) ReadNbitsWords64(nBits int) (res []uint64, err error) {
	if err = wr.checkNbitsSize(nBits); err != nil {
		return res, errors.Trace(err)
	}
	numOfWords := sizeInWords(nBits)
	offset := wr.offset

	tmpNBits := nBits
	_, wordOffset, localBitOffset, nextOffset, nextWordIndex, _ := wr.calcParams(tmpNBits, offset)

	resWords := make([]uint64, 0, numOfWords)
	//fmt.Println("len resWords: ", numOfWords, localBitOffset, nextOffset, nextWordIndex, nextLocalOffset, numOfWords)
	for i := 0; i < (numOfWords); i++ {
		var lOffset int
		var mask uint64
		var lerr error

		if i == 0 {
			lOffset = localBitOffset
			if nBits > 64 {
				tmpNBits = 64
			} else {
				tmpNBits = nBits
			}
		} else if i == numOfWords-1 {
			lOffset = 0
			tmpNBits = nBits - (numOfWords-1)*64
		} else {
			lOffset = 0
			tmpNBits = 64
		}

		mask, lerr = calcMask64(tmpNBits)
		if lerr != nil {
			return nil, errors.Trace(err)
		}
		if len(resWords) == i+1 {
			resWords[i] = resWords[i] | ((wr.inWord[wordOffset+i] >> lOffset) & mask)
		} else {
			r := (wr.inWord[wordOffset+i] >> lOffset) & mask
			resWords = append(resWords, r)
		}

	}
	wr.offset = nextOffset
	wr.accOffset += wr.offset
	wr.currWordIndex = nextWordIndex

	return resWords, nil
}

func (wr *Reader) ReadNbitsUint64(nBits int) (res uint64, err error) {
	if err = wr.checkNbitsSize(nBits); err != nil {
		return res, errors.Trace(err)
	}
	resWords, err := wr.ReadNbitsWords64(nBits)
	if err != nil {
		return 0, errors.Trace(err)
	}
	return resWords[0], nil
}

func (wr *Reader) ReadNbitsBytes(nBits int) (res []byte, err error) {
	if err = wr.checkNbitsSize(nBits); err != nil {
		return res, errors.Trace(err)
	}
	resWords, err := wr.ReadNbitsWords64(nBits)
	if err != nil {
		return nil, errors.Trace(err)
	}

	//fmt.Printf("resWords: %X\n", resWords)

	// TODO: remove this
	if len(resWords) != sizeInWords(nBits) {
		err = InvalidResultAssertionError
		err = errors.Annotatef(err, "expected resWords size: %d, got %d", sizeInWords(nBits), len(resWords))
		return nil, errors.Trace(err)
	}

	res = make([]byte, 0, len(resWords)*8)
	sizeInBytes := BitsToBytesSize(nBits)

	if wr.isLittleEndian {
		for _, word := range resWords {
			res = binary.LittleEndian.AppendUint64(res, word)
		}
		return res[:sizeInBytes], nil
	}
	for i := range resWords {
		res = binary.BigEndian.AppendUint64(res, resWords[(len(resWords)-1)-i])
	}
	return res[len(res)-sizeInBytes:], nil
}

func (wr *Reader) Words() []uint64 { return wr.inWord }
