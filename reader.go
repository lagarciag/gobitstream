package gobitstream

import (
	"encoding/binary"
	"github.com/juju/errors"
)

// Reader is a bit stream Reader.
// It does not have io.Reader interface
type Reader struct {
	currWordIndex  int
	currBitIndex   int // MSB: 7, LSB: 0
	offset         int
	size           int
	accOffset      int
	inWord         []uint64
	resWordsBuffer []uint64
	in             []byte
	resBytesBuffer []byte //
	isLittleEndian bool
}

// NewReader creates a new Reader instance.
func NewReader(sizeInBits int, in []byte) (wr *Reader, err error) {
	if len(in) < BitsToBytesSize(sizeInBits) {
		err = InvalidBitsSizeError
		return nil, errors.Trace(err)
	}
	wr = &Reader{}
	wr.resWordsBuffer = make([]uint64, 0, len(in)*8)
	wr.resBytesBuffer = make([]byte, 0, len(in))

	wr.size = sizeInBits
	wr.in = in
	wr.inWord, err = ConvertBytesToWords(sizeInBits, in)
	return wr, err
}

// NewReaderLE creates a new Reader instance.
func NewReaderLE(sizeInBits int, in []byte) (wr *Reader, err error) {
	wr, err = NewReader(sizeInBits, in)
	if err != nil {
		return wr, errors.Trace(err)
	}
	wr.isLittleEndian = true
	wr.currWordIndex = len(wr.inWord) - 1
	return wr, err
}

func NewReaderBE(sizeInBits int, in []byte) (wr *Reader, err error) {
	wr, err = NewReader(sizeInBits, in)
	if err != nil {
		return wr, errors.Trace(err)
	}
	wr.isLittleEndian = false
	wr.currWordIndex = 0

	return wr, err
}

func (wr *Reader) Reset() {
	wr.accOffset = 0
	wr.currWordIndex = 0
	wr.currBitIndex = 0
	wr.offset = 0
	if wr.isLittleEndian {
		wr.currWordIndex = len(wr.inWord) - 1
	} else {
		wr.currWordIndex = 0
	}
}

func (wr *Reader) checkNbitsSize(nBits int) error {
	if nBits <= 0 {
		err := InvalidBitsSizeError
		err = errors.Annotate(err, "nBits cannot be 0")
		return errors.Trace(err)
	} else if nBits+wr.accOffset > wr.size {
		err := InvalidBitsSizeError
		err = errors.Annotatef(err, "nBits+wr.accOffset > wr.size , nBits: %d, accOffset: %d, wr.size: %d", nBits, wr.accOffset, wr.size)
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

	resWords := wr.resWordsBuffer
	resWords = resWords[:0]

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
			if tmpNBits < 0 {
				err = InvalidResultAssertionError
				err = errors.Annotatef(err, "nBits: %d . numOfWords: %d", tmpNBits, numOfWords)
				errors.Trace(err)
				return nil, err
			}

		} else {
			lOffset = 0
			tmpNBits = 64
		}
		if tmpNBits < 0 {
			err = InvalidResultAssertionError
			err = errors.Annotatef(err, "nBits: %d . numOfWords: %d", tmpNBits, numOfWords)
			errors.Trace(err)
			return nil, err
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
	if len(resWords) == 0 {
		err = InvalidResultAssertionError
		err = errors.Annotatef(err, "nBits: %d", nBits)
		return 0, errors.Trace(err)
	}
	return resWords[0], nil
}

func (wr *Reader) ReadNbitsBytes(nBits int) (resultBytes []byte, err error) {
	if err = wr.checkNbitsSize(nBits); err != nil {
		return resultBytes, errors.Trace(err)
	}
	resultWords, err := wr.ReadNbitsWords64(nBits)
	if err != nil {
		return nil, errors.Trace(err)
	}

	// TODO: remove this
	if len(resultWords) != sizeInWords(nBits) {
		err = InvalidResultAssertionError
		err = errors.Annotatef(err, "expected resultWords size: %d, got %d", sizeInWords(nBits), len(resultWords))
		return nil, errors.Trace(err)
	}

	wr.resBytesBuffer = wr.resBytesBuffer[:0]
	resultBytes = wr.resBytesBuffer

	sizeInBytes := BitsToBytesSize(nBits)

	if wr.isLittleEndian {
		for _, word := range resultWords {
			resultBytes = binary.LittleEndian.AppendUint64(resultBytes, word)
		}

		return resultBytes[:sizeInBytes], nil
	}
	for i := range resultWords {
		resultBytes = binary.BigEndian.AppendUint64(resultBytes, resultWords[(len(resultWords)-1)-i])
	}

	return resultBytes[len(resultBytes)-sizeInBytes:], nil
}

func (wr *Reader) Words() []uint64 { return wr.inWord }
