package gobitstream

import (
	"fmt"
	"github.com/pkg/errors"
)

func GetBitsFieldFromSliceBroken(fieldSlice []uint64, widthInBits, offsetInBits uint64) (outputField []uint64, err error) {

	inputFieldSlice := make([]uint64, len(fieldSlice))
	copy(inputFieldSlice, fieldSlice)

	if widthInBits == 0 {
		return nil, errors.New("widthInBits cannot be 0")
	}

	// Calculate which elements in the slice we need to consider
	startElement := offsetInBits / 64
	endElement := (offsetInBits + widthInBits - 1) / 64

	if endElement >= uint64(len(inputFieldSlice)) {
		return nil, errors.New("offset and width exceed the size of the inputFieldSlice")
	}

	outputField = make([]uint64, endElement-startElement+1)

	// Calculate the local offset within the startElement
	localOffset := offsetInBits % 64

	// Loop over each uint64 in the outputField slice
	for i := range outputField {
		if i == 0 {
			// If the field is contained within a single slice element
			if startElement == endElement {
				outputField[i] = (inputFieldSlice[startElement] >> localOffset) & ((1 << widthInBits) - 1)
			} else { // If the field spans multiple elements in the slice
				lowerBits := inputFieldSlice[int(startElement)+i] >> localOffset
				upperBits := inputFieldSlice[int(startElement)+i+1] << (64 - localOffset)
				outputField[i] = lowerBits | upperBits
			}
		} else if uint64(i) < endElement-startElement {
			// For middle elements, take the whole uint64 value
			outputField[i] = inputFieldSlice[int(startElement)+i]
		} else {
			// For the last element, mask out bits beyond the field width
			outputField[i] = inputFieldSlice[int(startElement)+i] & ((1 << (widthInBits % 64)) - 1)
		}
	}
	return outputField, nil
}

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
		if remainingBits > 64 {
			remainingBits -= 64
			localWidth = 64
		} else {
			localWidth = int(remainingBits % 64)
		}

		fmt.Printf("localWidth: %d offsetInBits: %d\n", localWidth, offsetInBits)
		resultSubBitstream[i], err = Get64BitsFieldFromSlice(fieldSlice, uint64(localWidth), offsetInBits+uint64(i*64))
		if err != nil {
			return nil, err
		}

	}

	return resultSubBitstream, nil

}

func calculateFieldParameters(width uint64) (remainingWidth, widthWords int, lastWordMask uint64) {
	remainingWidth = int(width)
	widthWords = remainingWidth / 64
	mod64 := remainingWidth % 64
	if mod64 > 0 {
		widthWords++
		lastWordMask = (1 << mod64) - 1
	}
	return remainingWidth, widthWords, lastWordMask
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
