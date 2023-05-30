package gobitstream

import (
	"github.com/pkg/errors"
)

// Get64BitsFieldFromSlice extracts a field of bits from a slice of uint64.
// This function is used when you want to extract a field of specified width and at specified offset from a slice of uint64.
//
// Parameters:
// inputBuffer: A pre-allocated slice that can be used to avoid internal memory allocation. If nil, a new slice will be allocated.
// fieldSlice: Slice of uint64 from which the field will be extracted.
// widthInBits: Width of the field to be extracted.
// offsetInBits: Starting position of the field in the slice.
//
// Returns:
// uint64 value of the extracted field.
// error if widthInBits is 0 or more than 64, if offsetInBits and widthInBits combination is out of range of input slice,
// or if the provided inputBuffer isn't large enough to hold the fieldSlice.
func Get64BitsFieldFromSlice(inputBuffer, fieldSlice []uint64, widthInBits, offsetInBits uint64) (outputField uint64, err error) {

	// If an input buffer is provided and it's large enough, use it to avoid internal allocation.
	// Otherwise, allocate a new slice.
	// Create a copy of the input slice to avoid any side effects on the input data
	inputFieldSlice, err := prepareInputSlice(inputBuffer, fieldSlice)

	// Checking if width is more than 64 bits. If so, return an error as it cannot handle more than 64 bits.
	if widthInBits > 64 {
		return 0, errors.New("widthInBits cannot exceed 64")
	}

	// Checking if width is 0. If so, return an error as we cannot have a field of 0 width.
	if widthInBits == 0 {
		return 0, errors.New("widthInBits cannot be 0")
	}

	// Calculating the start and end elements in the slice that the field spans.
	startElement := offsetInBits / 64
	endElement := (offsetInBits + widthInBits - 1) / 64

	// If the field spans beyond the slice, return an error.
	if endElement >= uint64(len(inputFieldSlice)) {
		return 0, errors.New("offset and width exceed the size of the inputFieldSlice")
	}

	// Calculating the local offset within the start element.
	localOffset := offsetInBits % 64

	// If the field is contained within a single slice element, extract the field and return.
	if startElement == endElement {
		return (inputFieldSlice[startElement] >> localOffset) & ((1 << widthInBits) - 1), nil
	}

	// If the field spans two elements in the slice, we extract the lower and upper bits separately.
	lowerBits := inputFieldSlice[startElement] >> localOffset
	upperBits := inputFieldSlice[endElement] << (64 - localOffset)

	// Combine the lower and upper bits.
	outputField = lowerBits | upperBits

	// Mask the field to the desired width.
	outputField = outputField & ((1 << widthInBits) - 1)

	// Return the output field and nil error.
	return outputField, nil
}

// Set64BitsFieldToSlice sets a field of bits in a slice of uint64.
// This function is used when you want to set a field of specified width and at specified offset in a slice of uint64.
//
// Parameters:
// inputFieldField: The slice of uint64 where the field will be set.
// inputField: The field value to be set.
// widthInBits: Width of the field to be set.
// offsetInBits: Starting position of the field in the slice.
//
// Returns:
// Slice of uint64 with the field set at the specified position and width.
// error if widthInBits is 0 or more than 64, or if offsetInBits and widthInBits combination is out of range of input slice.
func Set64BitsFieldToSlice(inputFieldField []uint64, inputField, widthInBits, offsetInBits uint64) (result []uint64, err error) {
	// Creating a copy of the input slice to avoid any side effects on the input data
	result = make([]uint64, len(inputFieldField))
	copy(result, inputFieldField)

	// Checking if width is more than 64 bits. If so, return an error as it cannot handle more than 64 bits.
	if widthInBits > 64 {
		return nil, errors.New("widthInBits cannot exceed 64")
	}

	// Checking if width is 0. If so, return an error as we cannot have a field of 0 width.
	if widthInBits == 0 {
		return nil, errors.New("widthInBits cannot be 0")
	}

	// Mask the input field to the desired width.
	inputField = inputField & ((1 << widthInBits) - 1)

	// Calculating the start and end elements in the slice that the field spans.
	startElement := offsetInBits / 64
	endElement := (offsetInBits + widthInBits - 1) / 64

	// If the field spans beyond the slice, return an error.
	if endElement >= uint64(len(result)) {
		return nil, errors.New("offset and width exceed the size of the inputFieldSlice")
	}

	// Calculating the local offset within the start element.
	localOffset := offsetInBits % 64

	// If the field is contained within a single slice element, set the field and return.
	if startElement == endElement {
		// Create a mask to preserve the bits outside of the field we're setting
		mask := uint64(^(((1 << widthInBits) - 1) << localOffset))
		result[startElement] = (result[startElement] & mask) | (inputField << localOffset)
	} else {
		// If the field spans two elements in the slice, we set the lower and upper bits separately.
		// Create masks to preserve the bits outside of the field we're setting in both elements
		lowerMask := uint64((1 << localOffset) - 1)
		upperMask := uint64(^((1 << (widthInBits - (64 - localOffset))) - 1))

		// Shifting the input field bits to match the offset in the slice
		lowerSiftedInputField := inputField << localOffset
		upperShiftedInputField := inputField >> (64 - localOffset)

		// Setting the field bits in the slice
		result[startElement] = (result[startElement] & lowerMask) | lowerSiftedInputField
		result[endElement] = (result[endElement] & upperMask) | upperShiftedInputField
	}

	// Return the updated slice and nil error.
	return result, nil
}

//----------------------------------------

// prepareInputSlice prepares the input slice for the Get64BitsFieldFromSlice function.
func prepareInputSlice(inputBuffer, fieldSlice []uint64) ([]uint64, error) {
	var inputFieldSlice []uint64
	if inputBuffer != nil && cap(inputBuffer) >= len(fieldSlice) {
		inputBuffer = inputBuffer[:len(fieldSlice)]
		copy(inputBuffer, fieldSlice)
		inputFieldSlice = inputBuffer
	} else if inputBuffer == nil {
		inputFieldSlice = make([]uint64, len(fieldSlice))
		copy(inputFieldSlice, fieldSlice)
	} else {
		return nil, errors.New("inputBuffer is not large enough to hold the fieldSlice")
	}
	return inputFieldSlice, nil
}

// calculateSliceElements calculates the start and end elements of the slice that the field spans.
func calculateSliceElements(inputFieldSlice []uint64, widthInBits, offsetInBits uint64) (uint64, uint64, error) {
	startElement := offsetInBits / 64
	endElement := (offsetInBits + widthInBits - 1) / 64

	if endElement >= uint64(len(inputFieldSlice)) {
		return 0, 0, errors.New("offset and width exceed the size of the inputFieldSlice")
	}
	return startElement, endElement, nil
}
