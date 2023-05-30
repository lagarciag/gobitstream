package gobitstream

import (
	"github.com/pkg/errors"
)

func SetFieldToSlice(dstSlice []uint64, field []uint64, widthInBits, offsetInBits uint64) ([]uint64, error) {
	//dstSlice := make([]uint64, len(inSlice))
	//copy(dstSlice, inSlice)
	// Check if offsetInBits is larger than the size of dstSlice in bits.
	if offsetInBits > uint64(len(dstSlice)*64) {
		return nil, errors.New("offsetInBits is out of range")
	}

	// Check if widthInBits is larger than the size of the field in bits.
	if widthInBits > uint64(len(field)*64) {
		return nil, errors.New("widthInBits is larger than the size of the field")
	}

	// Handle zero-widthInBits case: do nothing and return nil.
	if widthInBits == 0 {
		return nil, errors.New("widthInBits is zero")
	}

	if widthInBits <= 64 {
		var err error
		dstSlice, err = Set64BitsFieldToSlice(dstSlice, field[0], widthInBits, offsetInBits)
		if err != nil {
			return nil, err
		}
		return dstSlice, nil
	}

	remainingWidth := widthInBits

	for i := range field {
		var localWidth int
		// three cases:
		if i < len(field)-1 {
			if remainingWidth > 64 {
				localWidth = 64
				remainingWidth -= 64
			} else {
				localWidth = int(remainingWidth)
				remainingWidth = 0
			}

		} else if i == len(field)-1 {
			if remainingWidth > 64 {
				return nil, errors.New("widthInBits is larger than the size of the field")
			} else {
				localWidth = int(remainingWidth)
				remainingWidth = 0
			}
		}

		var err error
		dstSlice, err = Set64BitsFieldToSlice(dstSlice, field[i], uint64(localWidth), offsetInBits)
		if err != nil {
			return nil, err
		}
		offsetInBits += 64
	}
	return dstSlice, nil
}
