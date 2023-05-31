package gobitstream

import (
	"github.com/pkg/errors"
)

// SetFieldToSlice embeds a field of bits into a slice of uint64 at a specific offset.
// This function is used when you want to insert a field of specified width at a specified offset into a slice of uint64.
//
// Parameters:
// dstSlice: Destination slice of uint64 where the field will be inserted.
// field: Field of uint64 bits to be inserted into the destination slice.
// widthInBits: Width of the field to be inserted.
// offsetInBits: Offset in the destination slice where the field will be inserted.
//
// Returns:
// Updated slice with the inserted field.
// error if offsetInBits or widthInBits is out of range, or if widthInBits is zero.
func SetFieldToSlice(dstSlice []uint64, field []uint64, widthInBits, offsetInBits uint64) ([]uint64, error) {

	// Check if the offsetInBits is within the boundaries of the destination slice.
	if offsetInBits > uint64(len(dstSlice)*64) {
		return nil, errors.New("offsetInBits is out of range")
	}

	// Check if the widthInBits of the field is within its size.
	if widthInBits > uint64(len(field)*64) {
		return nil, errors.New("widthInBits is larger than the size of the field")
	}

	// If widthInBits is zero, return as there's nothing to insert.
	if widthInBits == 0 {
		return nil, errors.New("widthInBits is zero")
	}

	// If widthInBits is less than or equal to 64 bits, use the Set64BitsFieldToSlice function to insert the field.
	if widthInBits <= 64 {
		var err error
		dstSlice, err = Set64BitsFieldToSlice(dstSlice, field[0], widthInBits, offsetInBits)
		if err != nil {
			return nil, err
		}
		return dstSlice, nil
	}

	// For widthInBits greater than 64, insert the field in segments.
	remainingWidth := widthInBits
	for i := range field {
		localWidth, er := calculateLocalWidth(i, remainingWidth, field)
		if er != nil {
			return nil, er
		}
		remainingWidth -= uint64(localWidth)

		var err error
		dstSlice, err = Set64BitsFieldToSlice(dstSlice, field[i], uint64(localWidth), offsetInBits)
		if err != nil {
			return nil, err
		}
		offsetInBits += 64
	}

	// Return the updated slice.
	return dstSlice, nil
}

// calculateLocalWidth determines the width of the segment to be inserted based on the remaining width.
func calculateLocalWidth(index int, remainingWidth uint64, field []uint64) (int, error) {
	if index < len(field)-1 {
		if remainingWidth > 64 {
			return 64, nil
		} else {
			return int(remainingWidth), nil
		}
	}

	if remainingWidth > 64 {
		return 0, errors.New("widthInBits is larger than the size of the field")
	}

	return int(remainingWidth), nil
}
