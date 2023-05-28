package gobitstream

import (
	"fmt"
	"github.com/pkg/errors"
)

func Get64BitsFieldFromSlice(inputFieldSlice []uint64, widthInBits, offsetInBits uint64) (outputField uint64, err error) {
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
func Set64BitsFieldToWordSlice(destinationField []uint64, inputField, widthInBits, offsetInBits uint64) (err error) {
	if widthInBits > 64 {
		return errors.New("widthInBits cannot exceed 64")
	}

	if widthInBits == 0 {
		return errors.New("widthInBits cannot be 0")
	}

	inputField = inputField & ((1 << widthInBits) - 1)

	fmt.Printf("destinationField: %X\n", destinationField)
	fmt.Printf("inputField: %X\n", inputField)

	// Calculate which elements in the slice we need to consider
	startElement := offsetInBits / 64
	endElement := (offsetInBits + widthInBits - 1) / 64

	fmt.Printf("startElement: %d\n", startElement)
	fmt.Printf("endElement: %d\n", endElement)

	if endElement >= uint64(len(destinationField)) {
		return errors.New("offset and width exceed the size of the destinationField")
	}

	// Calculate the local offset within the startElement
	localOffset := offsetInBits % 64

	fmt.Printf("localOffset: %d\n", localOffset)

	if startElement == endElement {
		// If the field is contained within a single slice element
		// Create a mask to preserve the bits outside of the field we're setting
		mask := uint64(^(((1 << widthInBits) - 1) << localOffset))
		destinationField[startElement] = (destinationField[startElement] & mask) | (inputField << localOffset)
	} else {
		// If the field spans two elements in the slice
		// Create masks to preserve the bits outside of the field we're setting in both elements
		lowerMask := uint64(^((1 << localOffset) - 1))
		upperMask := uint64((1 << (64 - localOffset)) - 1)
		destinationField[startElement] = (destinationField[startElement] & lowerMask) | (inputField << localOffset)
		destinationField[endElement] = (destinationField[endElement] & upperMask) | (inputField >> (64 - localOffset))
	}

	return nil
}

///----------------------------------

// Set64BitsFieldToWordSlice sets a 64-bit field in a slice of uint64 words.
// This function receives a destination slice (dstSlice), a field value (field),
// the width of the field in bits (width) and the offset in bits from the beginning of the slice (offset).
// If the offset or width are not valid, it returns an error.
// If the field spans more than one word in the slice and the slice is not large enough to hold the result,
// it also returns an error.
func Set64BitsFieldToWordSlicex(dstSlice []uint64, field, width, offset uint64) error {
	err := validateParameters(dstSlice, width, offset)
	if err != nil {
		return errors.WithStack(err)
	}

	field &= (1 << width) - 1
	wordSpan := calculateWordSpan(width, offset)

	err = setFieldInSlice(dstSlice, field, offset, wordSpan)
	if err != nil {
		return errors.WithStack(err)
	}

	// Return the final result and no error
	return nil
}

// calculateWordSpan calculates the number of words the field spans
func calculateWordSpan(width, offset uint64) uint64 {
	wordSpan := (width + offset) / 64
	if (width+offset)%64 != 0 {
		wordSpan++
	}
	return wordSpan
}

// setFieldInSlice sets the field in the given slice
func setFieldInSlice(dstSlice []uint64, field, offset uint64, wordSpan uint64) error {
	dstSlice[0] = dstSlice[0] | (field << offset)
	if wordSpan > 1 {
		if len(dstSlice) < 2 {
			return errors.New("dstSlice is not large enough to hold the result")
		}
		dstSlice[1] = dstSlice[1] | (field >> (64 - offset))
	}
	return nil
}

// validateParameters performs the initial checks on the parameters of Set64BitsFieldToWordSlice
func validateParameters(dstSlice []uint64, width, offset uint64) error {
	if offset >= 64 {
		return errors.Errorf("offset must be less than 64, got %d", offset)
	}
	if width == 0 || width > 64 {
		return errors.Errorf("width must be between 1 and 64, got %d", width)
	}
	if offset >= uint64(len(dstSlice))*64 {
		return errors.Errorf("offset: %d is out of range", offset)
	}
	return nil
}
