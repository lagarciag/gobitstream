package gobitstream

import (
	"encoding/binary"
	"fmt"
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

	//fmt.Printf("inWord: %X\n", wr.inWord)

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
	inx := make([]byte, len(in))
	_ = copy(inx, in)
	reverseSlice(inx)
	wr, err = NewReader(sizeInBits, inx)
	if err != nil {
		return wr, errors.Trace(err)
	}

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

func (wr *Reader) ReadNbitsBytes(nBits int) (outBytes []byte, err error) {
	if err = wr.checkNbitsSize(nBits); err != nil {
		return outBytes, errors.Trace(err)
	}
	resultWords, err := wr.ReadNbitsWords64(nBits)
	if err != nil {
		return nil, errors.Trace(err)
	}

	fmt.Printf("%d - resultWords: %X", nBits, resultWords)

	// TODO: remove this
	if len(resultWords) != sizeInWords(nBits) {
		err = InvalidResultAssertionError
		err = errors.Annotatef(err, "expected resultWords size: %d, got %d", sizeInWords(nBits), len(resultWords))
		return nil, errors.Trace(err)
	}

	wr.resBytesBuffer = wr.resBytesBuffer[:0]
	resultBytes := wr.resBytesBuffer

	sizeInBytes := BitsToBytesSize(nBits)

	fmt.Printf("%d - 1 - resultWords %X  -- %d\n", nBits, resultWords, nBits)

	//if wr.isLittleEndian {
	for _, word := range resultWords {
		resultBytes = binary.LittleEndian.AppendUint64(resultBytes, word)
	}

	resultBytes = resultBytes[:sizeInBytes]

	if !wr.isLittleEndian {
		outBytes = make([]byte, len(resultBytes))
		_ = copy(outBytes, resultBytes)
		reverseSlice(outBytes)
		fmt.Printf("%d - 2 - resultWordsBE %X  -- %d\n", nBits, outBytes, nBits)
		return outBytes, nil
	}
	fmt.Printf("%d - 2 - resultWordsLE %X  -- %d\n", nBits, resultWords, nBits)

	return resultBytes, nil
	//}
	//for i := range resultWords {
	//	resultBytes = binary.BigEndian.AppendUint64(resultBytes, resultWords[(len(resultWords)-1)-i])
	//}

	//return resultBytes[len(resultBytes)-sizeInBytes:], nil
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
	if (width+offset)%64 != 0 {
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
	lastWordMask := 0
	// Compute the number of uint64 values required to store the field
	wordOffset := offset / 64
	localOffset := offset % 64
	localSlice := slice[wordOffset:]

	localWidth := width
	remainingWidth := width
	widthWords := int(width / 64)
	mod64 := width % 64
	if mod64 > 0 {
		widthWords++
		lastWordMask = (1 << mod64) - 1
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
		field, err := get64BitsFieldFromSlice(localSlice, localWidth, localOffset)
		if err != nil {
			err = errors.Annotatef(err, "wordOffset: %d, localWidth: %d, localOffset: %d localSlice: %X", wordOffset, localWidth, localOffset, localSlice)
			return nil, errors.Trace(err)
		}

		//fmt.Printf("local field: %X, wordOffset: %d localOffset: %d localSlice: %X \n", field, wordOffset, localOffset, localSlice)

		remainingWidth -= localWidth
		resultBuff = append(resultBuff, field)
	}
	resultBuff = ShiftSliceOfUint64Left(resultBuff, int(offset%64))

	if lastWordMask != 0 {
		resultBuff[len(resultBuff)-1] &= uint64(lastWordMask)
	}

	return resultBuff, nil
}

// ShiftSliceOfUint64Left performs a left shift on a slice of uint64 values by a given shift count.
// The shift is performed in place on the input slice.
//
// Parameters:
// - slice: A slice of uint64 values to be shifted.
// - shiftCount: The number of bits by which the values in the slice should be shifted left.
//   The shift count is computed modulo 64 to ensure that it falls within the valid range of 0 to 63.
//
// Returns:
// - slice: The input slice after performing the left shift operation.
//
// Behavior:
// - The function iterates over each uint64 value in the input slice.
// - Each value is shifted left by the specified shift count using the bitwise left shift operator.
// - The carry from the previous iteration, if any, is added to the shifted value using the bitwise OR operator.
// - The carry for the next iteration is updated by shifting the current value right by (64 - shift count)
//   using the bitwise right shift operator.
// - The input slice is updated in place with the shifted value.
// - If there is a remaining carry after iterating through the entire slice, it is appended to the end of the slice.
//
// Example Usage:
//   slice := []uint64{1, 2, 3}
//   shiftedSlice := ShiftSliceOfUint64Left(slice, 3)
//   fmt.Println(shiftedSlice) // Output: [8 16 24]

func ShiftSliceOfUint64Left(slice []uint64, shiftCount int) []uint64 {
	numShifts := shiftCount % 64
	carry := uint64(0)

	for i := 0; i < len(slice); i++ {
		// Shift left by numShifts bits
		temp := slice[i] << numShifts

		// Add carry from previous iteration
		temp |= carry

		// Update carry for next iteration
		carry = slice[i] >> (64 - numShifts)

		// Update slice with shifted value
		slice[i] = temp
	}

	// If there's a remaining carry, append it to the slice
	if carry > 0 {
		slice = append(slice, carry)
	}

	return slice
}
