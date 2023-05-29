package gobitstream

import "github.com/pkg/errors"

func GetAnySizeFieldFromUint64Slice(inputBitStream []uint64, widthInBits, offsetInBits uint64) (resultSubBitstream []uint64, err error) {
	wordOffset := offsetInBits / 64
	localOffset := offsetInBits % 64
	localSlice := inputBitStream[wordOffset:]

	localWidth, remainingWidth, widthWords, lastWordMask := calculateFieldParameters(widthInBits)

	wordsSize := sizeInWords(int(widthInBits))

	resultSubBitstream = make([]uint64, 0, wordsSize)

	for i := 0; i < widthWords; i++ {
		localOffset = calculateLocalOffset(i, int(localOffset))
		localWidth = calculateLocalWidth(remainingWidth, localWidth, i, int(widthInBits))

		field, err := Get64BitsFieldFromSlice(localSlice, uint64(localWidth), localOffset)
		if err != nil {
			err = errors.Wrapf(err, "widthInBits: %d, offsetInBits: %d", widthInBits, offsetInBits)
			return nil, errors.WithStack(err)
		}

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
