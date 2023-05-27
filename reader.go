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
			errors.New("invalid bits sizeInBytes"),
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
		err := errors.New("invalid bits sizeInBytes")
		err = errors.Wrap(err, "nBits cannot be 0")
		return errors.WithStack(err)
	} else if nBits+wr.offset > wr.size {
		err := errors.New("invalid bits sizeInBytes")
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
	resWords, err := getBitstreamFieldFromUint64Slice(wr.inWord, uint64(nBits), uint64(wr.offset))
	fmt.Printf("resWords: %X\n", resWords)

	wr.offset += nBits
	return resWords, errors.WithStack(err)
}

// ReadNbitsUint64 reads nBits number of bits from the bit stream and returns the resulting uint64 value.
// It also updates the offset in the bit stream. An error is returned if the number of bits to be read is invalid.
func (wr *Reader) ReadNbitsUint64(nBits int) (res uint64, err error) {
	if err = wr.checkNbitsSize(nBits); err != nil {
		return res, errors.WithStack(err)
	}

	resWords, err := getBitstreamFieldFromUint64Slice(wr.inWord, uint64(nBits), uint64(wr.offset))

	if err != nil {
		err = errors.Wrapf(err, "width: %d, offset %d", nBits, wr.offset)
		return 0, errors.WithStack(err)
	}
	if len(resWords) == 0 {
		err := errors.New("invalid result assertion")
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
	const asdf = 117 % 8
	fmt.Printf("resultWords %X %d\n", resultWords, nBits)

	// TODO: remove this
	if len(resultWords) != sizeInWords(nBits) {
		err := errors.New("invalid result assertion")
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

func get64BitsFieldFromSlice(inputFieldSlice []uint64, widthInBits, offsetInBits uint64) (outputField uint64, err error) {
	if widthInBits > 64 {
		return 0, errors.New("widthInBits cannot exceed 64")
	}

	if widthInBits == 0 {
		return 0, errors.New("widthInBits cannot be 0")
	}

	// Calculate which elements in the slice we need to consider
	startElement := offsetInBits / 64
	endElement := (offsetInBits + widthInBits - 1) / 64

	if endElement >= uint64(len(inputFieldSlice)) {
		return 0, errors.New("offset and width exceed the size of the inputFieldSlice")
	}

	// Calculate the local offset within the startElement
	localOffset := offsetInBits % 64

	// If the field is contained within a single slice element
	if startElement == endElement {
		return (inputFieldSlice[startElement] >> localOffset) & ((1 << widthInBits) - 1), nil
	}

	// If the field spans two elements in the slice
	lowerBits := inputFieldSlice[startElement] >> localOffset
	upperBits := inputFieldSlice[endElement] << (64 - localOffset)
	return (lowerBits | upperBits) & ((1 << widthInBits) - 1), nil
}

func get64BitsFieldFromSlice77(inputFieldSlice []uint64, widthInBits, offsetInBits uint64) (outputField uint64, err error) {
	if widthInBits > 64 {
		return 0, errors.New("widthInBits cannot exceed 64")
	}

	// Calculate which elements in the slice we need to consider
	startElement := offsetInBits / 64
	endElement := (offsetInBits + widthInBits - 1) / 64

	if endElement >= uint64(len(inputFieldSlice)) {
		return 0, errors.New("offset and width exceed the size of the inputFieldSlice")
	}

	// Calculate the local offset within the startElement
	localOffset := offsetInBits % 64

	// If the field is contained within a single slice element
	if startElement == endElement {
		return (inputFieldSlice[startElement] >> localOffset) & ((1 << widthInBits) - 1), nil
	}
	//slice := []uint64{0x0123456789abcdef, 0xfedcba9876543210}

	// lower = 0x0123456789
	// upper = 3210
	// If the field spans two elements in the slice
	lowerBitsMask := uint64((1<<(64-localOffset))-1) << localOffset

	fmt.Printf("lowerBitsMask %X %X\n", lowerBitsMask, inputFieldSlice[startElement])

	lowerBits := inputFieldSlice[startElement] & lowerBitsMask
	upperBits := inputFieldSlice[endElement] & ((1 << localOffset) - 1)

	fmt.Printf("lowerBits %X, upperBits %X\n", lowerBits, upperBits)

	return (lowerBits | upperBits) & ((1 << widthInBits) - 1), nil
}
func get64BitsFieldFromSliceZ(inputFieldSlice []uint64, widthInBits, offsetInBits uint64) (outputField uint64, err error) {
	if widthInBits > 64 {
		return 0, errors.New("widthInBits cannot exceed 64")
	}

	// Calculate which elements in the slice we need to consider
	startElement := offsetInBits / 64
	endElement := (offsetInBits + widthInBits - 1) / 64

	if endElement >= uint64(len(inputFieldSlice)) {
		return 0, errors.New("offset and width exceed the size of the inputFieldSlice")
	}

	// Calculate the local offset within the startElement
	localOffset := offsetInBits % 64

	// If the field is contained within a single slice element
	if startElement == endElement {
		return (inputFieldSlice[startElement] >> localOffset) & ((1 << widthInBits) - 1), nil
	}

	// If the field spans two elements in the slice

	lowerBits := (inputFieldSlice[startElement] >> localOffset) << (64 - localOffset)
	upperBits := (inputFieldSlice[endElement] << (64 - localOffset)) >> localOffset

	return (lowerBits | upperBits) & ((1 << widthInBits) - 1), nil
}

// getBitstreamFieldFromUint64Slice extracts a bitstream subfield from an input bitstream.
// The input bitstream is represented as a slice of uint64, where each uint64 represents 64 bits of the bitstream,
// starting from the least significant bit.
// The function takes as arguments the input bitstream, the width of the subfield in bits, and the offset (in bits)
// at which the subfield starts. The function returns the extracted subfield and any error encountered.
//
// inputBitstream: The input bitstream represented as a slice of uint64.
// widthInBits: The width of the subfield to be extracted, in bits.
// offsetInBits: The offset at which the subfield to be extracted starts, in bits.
//
// Returns a slice of uint64 representing the extracted subfield and an error. The error is non-nil if the requested
// subfield cannot be extracted (i.e., if widthInBits + offsetInBits exceeds the length of the input bitstream).
func getBitstreamFieldFromUint64SliceX(inputBitstream []uint64, widthInBits, offsetInBits uint64) (bitstreamSubField []uint64, err error) {
	// Check if the subfield can be extracted from the input bitstream
	if offsetInBits+widthInBits > uint64(len(inputBitstream))*64 {
		return nil, errors.New("width and offset exceeds bitstream length")
	}

	// Initialize the slice for the subfield
	bitstreamSubField = make([]uint64, ((offsetInBits+widthInBits+63)/64)-(offsetInBits/64))

	// Calculate the start and end indices in the input bitstream slice
	startIndex := offsetInBits / 64
	endIndex := (offsetInBits + widthInBits) / 64

	// Extract the lower bits of the subfield
	if startIndex == endIndex {
		// If the subfield is contained within one uint64, only extract the relevant bits
		bitstreamSubField[0] = (inputBitstream[startIndex] >> (offsetInBits % 64)) & ((1 << widthInBits) - 1)
	} else {
		// If the subfield spans multiple uint64s, include all higher bits in the first uint64
		bitstreamSubField[0] = inputBitstream[startIndex] >> (offsetInBits % 64)
	}

	// Extract the middle bits of the subfield
	for i := startIndex + 1; i < endIndex; i++ {
		bitstreamSubField[i-startIndex] = inputBitstream[i]
	}

	// Extract the upper bits of the subfield
	if startIndex != endIndex && widthInBits%64 != 0 {
		// If the subfield spans multiple uint64s and its width is not a multiple of 64,
		// only include the relevant lower bits in the last uint64
		bitstreamSubField[endIndex-startIndex] = inputBitstream[endIndex] & ((1 << (widthInBits % 64)) - 1)
	}

	return bitstreamSubField, nil
}

func getBitstreamFieldFromUint64Slice(inputBitStream []uint64, widthInBits, offsetInBits uint64) (resultSubBitstream []uint64, err error) {
	wordOffset := offsetInBits / 64
	localOffset := offsetInBits % 64
	localSlice := inputBitStream[wordOffset:]

	fmt.Printf("inputBitStream %X %d \n", inputBitStream, wordOffset)

	fmt.Println("wordOffset", wordOffset, "localOffset", localOffset, "localSlice",
		fmt.Sprintf("%X)", localSlice))

	localWidth, remainingWidth, widthWords, lastWordMask := calculateFieldParameters(widthInBits)

	fmt.Println("localWidth", localWidth)

	wordsSize := sizeInWords(int(widthInBits))

	resultSubBitstream = make([]uint64, 0, wordsSize)

	for i := 0; i < widthWords; i++ {
		localOffset = calculateLocalOffset(i, int(localOffset))
		localWidth = calculateLocalWidth(remainingWidth, localWidth, i, int(widthInBits))

		fmt.Println("xlocalOffset", localOffset, "xlocalWidth", localWidth)

		field, err := get64BitsFieldFromSlice(localSlice, uint64(localWidth), localOffset)
		if err != nil {
			err = errors.Wrapf(err, "widthInBits: %d, offsetInBits: %d", widthInBits, offsetInBits)
			return nil, errors.WithStack(err)
		}

		fmt.Printf("local slice %x -- field %x \n", localSlice, field)

		remainingWidth -= localWidth
		resultSubBitstream = append(resultSubBitstream, field)
	}

	if lastWordMask != 0 {
		resultSubBitstream[len(resultSubBitstream)-1] &= uint64(lastWordMask)
	}

	return resultSubBitstream, nil
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
//   - slice: A slice of uint64 values to be shifted.
//   - shiftCount: The number of bits by which the values in the slice should be shifted left.
//     The shift count is computed modulo 64 to ensure that it falls within the valid range of 0 to 63.
//
// Returns:
// - slice: The input slice after performing the left shift operation.
//
// Behavior:
//   - The function iterates over each uint64 value in the input slice.
//   - Each value is shifted left by the specified shift count using the bitwise left shift operator.
//   - The carry from the previous iteration, if any, is added to the shifted value using the bitwise OR operator.
//   - The carry for the next iteration is updated by shifting the current value right by (64 - shift count)
//     using the bitwise right shift operator.
//   - The input slice is updated in place with the shifted value.
//   - If there is a remaining carry after iterating through the entire slice, it is appended to the end of the slice.
//
// Example Usage:
//
//	slice := []uint64{1, 2, 3}
//	shiftedSlice := ShiftSliceOfUint64Left(slice, 3)
//	fmt.Println(shiftedSlice) // Output: [8 16 24]
func ShiftSliceOfUint64Left(slice []uint64, shiftCount int) ([]uint64, error) {

	numShifts := shiftCount % 64
	newIndex := shiftCount / 64

	// Check if newIndex is out of range
	if newIndex >= len(slice) {
		return nil, errors.New("shift count exceeds length of the slice")
	}

	carry := uint64(0)
	for i := 0; i < len(slice); i++ {
		// Shift left by numShifts bits
		temp := slice[i] << numShifts

		// Add carry from previous iteration
		temp |= carry

		// Update carry for next iteration
		carry = slice[i] >> (64 - numShifts)

		slice[i] = 0

		// Update slice with shifted value
		slice[i] = temp
	}

	// If there's a remaining carry, append it to the slice
	if carry > 0 {
		slice = append(slice, carry)
	}

	if newIndex > 0 {
		tempSlice := make([]uint64, len(slice))
		copy(tempSlice, slice)

		for i, v := range tempSlice {
			if newIndex+i < len(slice) {
				slice[newIndex+i] = v
			}
			if i < newIndex {
				slice[i] = 0
			}

		}
	}
	return slice, nil
}
