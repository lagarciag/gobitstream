package gobitstream

import (
	"github.com/pkg/errors"
)

func Get64BitsFieldFromSlice(fieldSlice []uint64, widthInBits, offsetInBits uint64) (outputField uint64, err error) {

	inputFieldSlice := make([]uint64, len(fieldSlice))
	copy(inputFieldSlice, fieldSlice)

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

	outputField = lowerBits | upperBits

	outputField = outputField & ((1 << widthInBits) - 1)
	return outputField, nil
}
func Set64BitsFieldToSlice(inputFieldField []uint64, inputField, widthInBits, offsetInBits uint64) (result []uint64, err error) {

	result = make([]uint64, len(inputFieldField))
	copy(result, inputFieldField)

	if widthInBits > 64 {
		return nil, err
	}

	if widthInBits == 0 {
		return nil, err
	}

	inputField = inputField & ((1 << widthInBits) - 1)

	// Calculate which elements in the slice we need to consider
	startElement := offsetInBits / 64
	endElement := (offsetInBits + widthInBits - 1) / 64

	if endElement >= uint64(len(result)) {
		return nil, errors.New("offset and width exceed the size of the inputFieldSlice")
	}

	// Calculate the local offset within the startElement
	localOffset := offsetInBits % 64

	if startElement == endElement {
		// If the field is contained within a single slice element
		// Create a mask to preserve the bits outside of the field we're setting
		mask := uint64(^(((1 << widthInBits) - 1) << localOffset))
		result[startElement] = (result[startElement] & mask) | (inputField << localOffset)
	} else {
		// If the field spans two elements in the slice
		// Create masks to preserve the bits outside of the field we're setting in both elements
		lowerMask := uint64((1 << localOffset) - 1)
		upperMask := uint64(^((1 << (widthInBits - (64 - localOffset))) - 1))

		lowerSiftedInputField := inputField << localOffset

		result[startElement] = (result[startElement] & lowerMask) | lowerSiftedInputField

		upperShiftedInputField := inputField >> (64 - localOffset)

		result[endElement] = (result[endElement] & upperMask) | upperShiftedInputField
	}

	return result, err
}
