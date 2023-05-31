package gobitstream

import (
	"github.com/pkg/errors"
)

// GetFieldFromSlice extracts a field of bits from a slice of uint64.
// This function is used when you want to extract a field of specified width and at specified offset from a slice of uint64.
//
// Parameters:
// fieldSlice: Slice of uint64 from which the field will be extracted.
// widthInBits: Width of the field to be extracted.
// offsetInBits: Starting position of the field in the slice.
//
// Returns:
// Slice of uint64 representing the extracted field.
// error if widthInBits is 0 or if offsetInBits and widthInBits combination is out of range of input slice.
func GetFieldFromSlice(widthInBits, offsetInBits uint64, fieldSlice, resultBuffer []uint64) (resultSubBitstream []uint64, err error) {

	// Check if the widthInBits is zero. If so, return an error as we cannot extract a field of 0 width.
	if widthInBits == 0 {
		return nil, errors.New("widthInBits cannot be 0")
	}

	// Calculate the number of words in the slice.
	sliceWords := calculateSliceWords(widthInBits)

	if resultBuffer == nil {
		resultSubBitstream = make([]uint64, sliceWords)
	} else {
		if len(resultBuffer) < int(sliceWords) {
			return nil, errors.New("resultBuffer is too small")
		}
		resultSubBitstream = resultBuffer[0:int(sliceWords)]
	}

	// Calculate the end element in the slice that the field ends.
	endElement := (offsetInBits + widthInBits - 1) / 64

	// If the field spans beyond the slice, return an error.
	if endElement >= uint64(len(fieldSlice)) {
		return nil, errors.New("offset and width exceed the size of the inputFieldSlice")
	}

	// If the widthInBits is less than or equal to 64 bits, use the Get64BitsFieldFromSlice function to extract the field.
	if widthInBits <= 64 {
		var err error
		resultSubBitstream[0], err = Get64BitsFieldFromSlice(fieldSlice, widthInBits, offsetInBits)
		return resultSubBitstream, err
	}

	// For widthInBits greater than 64, extract the field in segments.
	remainingBits := widthInBits
	for i := range resultSubBitstream {
		localWidth := calculateLocalWidth2(remainingBits)
		remainingBits -= uint64(localWidth)

		if localWidth == 0 {
			continue
		}

		var err error
		resultSubBitstream[i], err = Get64BitsFieldFromSlice(fieldSlice, uint64(localWidth), offsetInBits+uint64(i*64))
		if err != nil {
			return nil, err
		}

	}

	// Return the extracted field.
	return resultSubBitstream, nil
}

// calculateSliceWords calculates the number of words in the slice based on the widthInBits.
func calculateSliceWords(widthInBits uint64) uint64 {
	sliceWords := widthInBits / 64
	if widthInBits%64 != 0 {
		sliceWords++
	}
	return sliceWords
}

// calculateLocalWidth2 calculates the width of the segment to be extracted based on the remaining bits.
func calculateLocalWidth2(remainingBits uint64) int {
	if remainingBits >= 64 {
		return 64
	} else {
		return int(remainingBits % 64)
	}
}
