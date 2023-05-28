package gobitstream

import "github.com/pkg/errors"

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

	// Calculate which elements in the slice we need to consider
	startElement := offsetInBits / 64
	endElement := (offsetInBits + widthInBits - 1) / 64

	if endElement >= uint64(len(destinationField)) {
		return errors.New("offset and width exceed the size of the destinationField")
	}

	// Calculate the local offset within the startElement
	localOffset := offsetInBits % 64

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
