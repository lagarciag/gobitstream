// Package gobitstream provides a bit stream reader and a bit stream writer in Go, allowing reading bits from a byte slice.
// It supports both little-endian and big-endian byte orders.
package gobitstream

import (
	"encoding/binary"
	"fmt"

	"github.com/pkg/errors"
)

// Reader is a bit stream reader that allows reading bits from a byte slice.
type Reader struct {
	currBitIndex   int      // Current bit index in the current byte
	currWordIndex  int      // Current word index in the inWord slice
	offset         int      // Current offset in the bit stream
	size           int      // Size of the bit stream in bits
	inWord         []uint64 // Slice of uint64 words that represents the input byte slice
	resWordsBuffer []uint64 // Buffer to store the resulting words read from the bit stream
	in             []byte   // Input byte slice from which bits are read
	resBytesBuffer []byte   // Buffer to store the resulting bytes read from the bit stream
	isLittleEndian bool     // Boolean flag indicating whether the byte order is little-endian
}

// NewReader creates a new Reader instance with the specified size in bits and input byte slice.
// It returns a pointer to the created Reader and an error if the input byte slice is smaller than the specified size in bits.
func NewReader(sizeInBits int, in []byte) (wr *Reader, err error) {
	if len(in) < BitsToBytesSize(sizeInBits) {
		err = errors.Wrap(
			InvalidBitsSizeError,
			fmt.Sprintf("input byte slice is smaller than the specified size in bits: %d versus sizeInBits: %d",
				len(in), sizeInBits))
		return nil, errors.WithStack(err)
	}
	wr = &Reader{}
	wr.resWordsBuffer = make([]uint64, 0, len(in)*8)
	wr.resBytesBuffer = make([]byte, 0, len(in))

	wr.size = sizeInBits
	wr.in = in

	wr.inWord, err = ConvertBytesToWords(sizeInBits, in)

	//fmt.Printf("inWord: %X\n", wr.inWord)

	return wr, errors.WithStack(err)
}

// NewReaderLE creates a new Reader instance with the specified size in bits and input byte slice in little-endian byte order.
// It returns a pointer to the created Reader and an error if the input byte slice is smaller than the specified size in bits.
func NewReaderLE(sizeInBits int, in []byte) (wr *Reader, err error) {
	wr, err = NewReader(sizeInBits, in)
	if err != nil {
		return wr, errors.WithStack(err)
	}
	wr.isLittleEndian = true
	wr.currWordIndex = len(wr.inWord) - 1
	return wr, errors.WithStack(err)
}

// NewReaderBE creates a new Reader instance with the specified size in bits and input byte slice in big-endian byte order.
// It returns a pointer to the created Reader and an error if the input byte slice is smaller than the specified size in bits.
func NewReaderBE(sizeInBits int, in []byte) (wr *Reader, err error) {
	inx := make([]byte, len(in))
	_ = copy(inx, in)
	reverseSlice(inx)
	wr, err = NewReader(sizeInBits, inx)
	if err != nil {
		return wr, errors.WithStack(err)
	}

	wr.currWordIndex = 0

	return wr, errors.WithStack(err)
}

// Reset resets the Reader to its initial state, including resetting the current bit and word indices, offset, and byte order.
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

// checkNbitsSize checks the size of nBits and validates it against the Reader's offset and size.
// It returns an error if the size is invalid.
func (wr *Reader) checkNbitsSize(nBits int) error {
	if nBits <= 0 {
		err := InvalidBitsSizeError
		err = errors.Wrap(err, "nBits cannot be 0")
		return errors.WithStack(err)
	} else if nBits+wr.offset > wr.size {
		err := InvalidBitsSizeError
		errWrap := fmt.Sprintf("nBits+wr.accOffset > wr.size, nBits: %d, accOffset: %d, wr.size: %d", nBits, wr.offset, wr.size)
		err = errors.Wrap(err, errWrap)
		return errors.WithStack(err)
	}

	return nil
}

// calcParams calculates various parameters based on the provided nBits and offset values.
// It returns sizeInBytes, wordOffset, localOffset, nextOffset, nextWordIndex, and nextLocalOffset.
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

// ReadNbitsWords64 reads nBits number of bits from the bit stream and returns the resulting words as a slice of uint64 values.
// It also updates the offset in the bit stream. An error is returned if the number of bits to be read is invalid.
func (wr *Reader) ReadNbitsWords64(nBits int) (res []uint64, err error) {
	if err = wr.checkNbitsSize(nBits); err != nil {
		return res, err
	}
	resWords, err := getFieldFromSlice(wr.resWordsBuffer, wr.inWord, uint64(nBits), uint64(wr.offset))
	wr.offset += nBits
	return resWords, errors.WithStack(err)
}

// ReadNbitsUint64 reads nBits number of bits from the bit stream and returns the resulting uint64 value.
// It also updates the offset in the bit stream. An error is returned if the number of bits to be read is invalid.
func (wr *Reader) ReadNbitsUint64(nBits int) (res uint64, err error) {
	if err = wr.checkNbitsSize(nBits); err != nil {
		return res, errors.WithStack(err)
	}

	resWords, err := getFieldFromSlice(wr.resWordsBuffer, wr.inWord, uint64(nBits), uint64(wr.offset))

	if err != nil {
		err = errors.Wrapf(err, "width: %d, offset %d", nBits, wr.offset)
		return 0, errors.WithStack(err)
	}
	if len(resWords) == 0 {
		err = InvalidResultAssertionError
		err = errors.Wrapf(err, "nBits: %d", nBits)
		return 0, errors.WithStack(err)
	}
	wr.offset += nBits
	return resWords[0], nil
}

// ReadNbitsBytes reads nBits number of bits from the bit stream and returns the resulting bytes value.
// It also updates the offset in the bit stream. An error is returned if the number of bits to be read is invalid.
func (wr *Reader) ReadNbitsBytes(nBits int) (outBytes []byte, err error) {
	if err = wr.checkNbitsSize(nBits); err != nil {
		return outBytes, err
	}
	resultWords, err := wr.ReadNbitsWords64(nBits)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	fmt.Printf("%d - resultWords: %X", nBits, resultWords)

	// TODO: remove this
	if len(resultWords) != sizeInWords(nBits) {
		err = InvalidResultAssertionError
		err = errors.Wrapf(err, "expected resultWords size: %d, got %d", sizeInWords(nBits), len(resultWords))
		return nil, errors.WithStack(err)
	}

	wr.resBytesBuffer = wr.resBytesBuffer[:0]
	resultBytes := wr.resBytesBuffer

	sizeInBytes := BitsToBytesSize(nBits)

	for _, word := range resultWords {
		resultBytes = binary.LittleEndian.AppendUint64(resultBytes, word)
	}

	resultBytes = resultBytes[:sizeInBytes]

	if !wr.isLittleEndian {
		outBytes = make([]byte, len(resultBytes))
		_ = copy(outBytes, resultBytes)
		reverseSlice(outBytes)
		return outBytes, nil
	}

	return resultBytes, nil
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
		errors.Wrapf(err, "offset must be less than 64, got %d", offset)
		return 0, errors.WithStack(err)
	}

	if width == 0 || width > 64 {
		err := InvalidBitsSizeError
		errors.Wrapf(err, "width must be between 1 and 64, got %d", width)
		return 0, errors.WithStack(err)
	}
	if offset >= uint64(len(slice))*64 {
		err := OffsetOutOfRangeError
		errors.Wrapf(err, "offset: %d", offset)
		return 0, errors.WithStack(err)
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

// getFieldFromSlice returns a slice of uint64 that represents a field of specified width and offset
// from the provided slice of uint64 values.
//
// The function takes an output slice to store the resulting field values, a slice of uint64 values,
// a width representing the number of bits to extract, and an offset representing the position of
// the first bit to extract. It returns the resulting field values as a slice of uint64 and an error.
//
// The function splits the extraction process into smaller steps. It calculates the necessary parameters
// based on the width, such as the localWidth, remainingWidth, widthWords, and lastWordMask, which are used
// to determine the number of iterations and mask the last word if necessary. It iterates over the widthWords
// and extracts the relevant bits from the localSlice using the get64BitsFieldFromSlice function. It appends
// the extracted field values to the output slice. If the widthWords is greater than 1, it performs a bitwise
// left shift operation on the output slice. Finally, if the lastWordMask is nonzero, it applies the mask
// to the last element of the output slice.
//
// If any error occurs during the extraction process, such as an invalid width or offset, it returns
// an appropriate error with additional information. Otherwise, it returns the resulting field values
// and no error.
func getFieldFromSlice(out []uint64, slice []uint64, width, offset uint64) ([]uint64, error) {
	wordOffset := offset / 64
	localOffset := offset % 64
	localSlice := slice[wordOffset:]

	localWidth, remainingWidth, widthWords, lastWordMask := calculateFieldParameters(width)

	out = out[:0]

	for i := 0; i < widthWords; i++ {
		localOffset = calculateLocalOffset(i, int(localOffset))
		localWidth = calculateLocalWidth(remainingWidth, localWidth, i, int(width))

		field, err := get64BitsFieldFromSlice(localSlice, uint64(localWidth), localOffset)
		if err != nil {
			err = errors.Wrapf(err, "width: %d, offset: %d", width, offset)
			return nil, errors.WithStack(err)
		}

		remainingWidth -= localWidth
		out = append(out, field)
	}

	// if widthWords > 1 {
	// 	out = ShiftSliceOfUint64Left(out, int(offset%64))
	// }

	if lastWordMask != 0 {
		out[len(out)-1] &= uint64(lastWordMask)
	}

	return out, nil
}

func calculateFieldParameters(width uint64) (localWidth, remainingWidth, widthWords int, lastWordMask uint64) {
	localWidth = int(width % 64)
	remainingWidth = int(width)
	widthWords = remainingWidth / 64
	mod64 := remainingWidth % 64
	if mod64 > 0 {
		widthWords++
		lastWordMask = (1 << mod64) - 1
	}
	return localWidth, remainingWidth, widthWords, lastWordMask
}

func calculateLocalOffset(i, localOffset int) uint64 {
	if i != 0 {
		return 0
	}
	return uint64(localOffset)
}

func calculateLocalWidth(remainingWidth, localWidth, i, width int) int {
	if remainingWidth > 64 && i == 0 {
		return 64 - localWidth
	} else if remainingWidth >= 64 {
		return 64
	} else {
		return width % 64
	}
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
