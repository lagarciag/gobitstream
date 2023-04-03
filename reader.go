package gobitstream

import (
	"encoding/binary"
	"github.com/juju/errors"
)

// Reader is a bit stream Reader.
// It does not have io.Reader interface
type Reader struct {
	currBitIndex   int // MSB: 7, LSB: 0
	currWordIndex  int // MSB: 7, LSB: 0
	offset         int
	size           int
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
	} else if nBits+wr.offset > wr.size {
		err := InvalidBitsSizeError
		err = errors.Annotatef(err, "nBits+wr.accOffset > wr.size , nBits: %d, accOffset: %d, wr.size: %d", nBits, wr.offset, wr.size)
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
	resWords, err := getFieldFromSlice(wr.resWordsBuffer, wr.inWord, uint64(nBits), uint64(wr.offset))
	wr.offset += nBits
	return resWords, nil
}

func (wr *Reader) ReadNbitsUint64(nBits int) (res uint64, err error) {
	if err = wr.checkNbitsSize(nBits); err != nil {
		return res, errors.Trace(err)
	}

	resWords, err := getFieldFromSlice(wr.resWordsBuffer, wr.inWord, uint64(nBits), uint64(wr.offset))
	if err != nil {
		err = errors.Annotatef(err, "width: %d, offset %d", nBits, wr.offset)
		return 0, errors.Trace(err)
	}
	if len(resWords) == 0 {
		err = InvalidResultAssertionError
		err = errors.Annotatef(err, "nBits: %d", nBits)
		return 0, errors.Trace(err)
	}
	wr.offset += nBits
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

// get64BitsFieldFromSlice extracts a range of bits from a slice of uint64s.
//
// The function takes a slice of uint64s, a width representing the number of bits to extract,
// and an offset representing the position of the first bit to extract. The function returns
// the extracted bits as a uint64 and an error.
//
// If the given width is 0 or greater than 64, or if the given offset is out of range for
// the slice, the function returns an InvalidBitsSizeError or an OffsetOutOfRangeError,
// respectively. Otherwise, the function extracts the relevant bits from the slice and returns
// the result and no error.
//
// The function calculates the number of uint64s required to extract the full range of bits,
// and uses bitwise operators to extract the relevant bits from those uint64s. The result is
// returned as a uint64.
func get64BitsFieldFromSlice(slice []uint64, width, offset uint64) (uint64, error) {
	if offset > 64 {
		err := InvalidOffsetError
		errors.Annotatef(err, "offset must be less than 64, got %d", offset)
		return 0, errors.Trace(err)
	}

	if width == 0 || width > 64 {
		err := InvalidBitsSizeError
		errors.Annotatef(err, "width must be between 1 and 64, got %d", width)
		return 0, errors.Trace(err)
	}
	if offset >= uint64(len(slice))*64 {
		err := OffsetOutOfRangeError
		errors.Annotatef(err, "offset: %d", offset)
		return 0, errors.Trace(err)
	}

	// Initialize the result variable to 0
	result := uint64(0)
	words := (width + offset) / 64
	if width%64 != 0 {
		words++
	}
	result = (slice[0] >> offset) & ((1 << width) - 1)
	if words > 1 {
		remainingBits := width - (64 - offset)
		result |= slice[1] & ((1 << remainingBits) - 1) << (64 - offset)
	}

	// Return the final result and no error
	return result, nil
}

func getFieldFromSlice(resultBuff []uint64, slice []uint64, width, offset uint64) (out []uint64, err error) {
	// Compute the number of uint64 values required to store the field
	//fieldSize := width
	wordOffset := offset / 64
	localOffset := offset % 64
	localSlice := slice[wordOffset:]
	localWidth := width
	remainingWidth := width
	widthWords := int(width / 64)
	if width%64 > 0 {
		widthWords++
	}

	// Allocate a slice to store the field
	resultBuff = resultBuff[:0]

	// Extract the bits of the field from the slice
	for i := 0; i < widthWords; i++ {
		if i != 0 {
			localOffset = 0
		}
		if remainingWidth > 64 && i == 0 {
			localWidth = 64 - localOffset
		} else if remainingWidth >= 64 {
			localWidth = 64
		} else {
			localWidth = width % 64
		}
		//fmt.Printf("localSlice: %X\n", localSlice)
		//fmt.Printf("localWidth: %d, localOffset: %d, wordOffset: %d\n", localWidth, localOffset, wordOffset)
		field, err := get64BitsFieldFromSlice(localSlice, localWidth, localOffset)
		if err != nil {
			err = errors.Annotatef(err, "wordOffset: %d, localWidth: %d, localOffset: %d", wordOffset, localWidth, localOffset)
			return nil, errors.Trace(err)
		}

		//fmt.Printf("local field: %X\n", field)

		remainingWidth -= localWidth
		resultBuff = append(resultBuff, field)
	}
	resultBuff, err = shiftSliceofUint64(resultBuff, int(offset%64))
	return resultBuff, nil
}

func shiftSliceofUint64(slice []uint64, shiftCount int) (result []uint64, err error) {
	lenSlice := len(slice)
	mask := uint64((1 << shiftCount) - 1)
	for i := 1; i < lenSlice; i++ {
		val := slice[i] & mask
		val = val<<64 - uint64(shiftCount)
		slice[i-1] = slice[i-1] | val
	}
	return slice, err
}
