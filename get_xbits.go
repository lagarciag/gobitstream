package gobitstream

import (
	"github.com/pkg/errors"
)

func GetFieldFromSlice(fieldSlice []uint64, widthInBits, offsetInBits uint64) (resultSubBitstream []uint64, err error) {

	inputFieldSlice := make([]uint64, len(fieldSlice))
	copy(inputFieldSlice, fieldSlice)

	if widthInBits == 0 {
		return nil, errors.New("widthInBits cannot be 0")
	}

	// Calculate which elements in the slice we need to consider
	endElement := (offsetInBits + widthInBits - 1) / 64

	if endElement >= uint64(len(inputFieldSlice)) {
		return nil, errors.New("offset and width exceed the size of the inputFieldSlice")

	}

	sliceWords := widthInBits / 64
	if widthInBits%64 != 0 {
		sliceWords++
	}
	resultSubBitstream = make([]uint64, sliceWords)

	if widthInBits <= 64 {
		resultSubBitstream[0], err = Get64BitsFieldFromSlice(inputFieldSlice, widthInBits, offsetInBits)
		return resultSubBitstream, err
	}

	remainingBits := widthInBits

	for i := range resultSubBitstream {
		var localWidth int
		if remainingBits >= 64 {
			remainingBits -= 64
			localWidth = 64
		} else {
			localWidth = int(remainingBits % 64)
		}

		if localWidth == 0 {
			continue
		}

		resultSubBitstream[i], err = Get64BitsFieldFromSlice(fieldSlice, uint64(localWidth), offsetInBits+uint64(i*64))
		if err != nil {
			return nil, err
		}

	}

	return resultSubBitstream, nil

}
