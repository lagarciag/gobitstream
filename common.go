package gobitstream

import (
	"encoding/binary"
	"fmt"

	"github.com/pkg/errors"
)

func totalOffsetToLocalOffset(offset int) (wordOffset, localOffset int) {
	return offset / 64, offset % 64
}

// calcNextSizes calculates the sizes for the next iteration when dealing with data that has a size greater than 64 bits (or 8 bytes).
// It subtracts 64 bits (8 bytes) from the current size, and also provides the current size limited to a maximum of 64 bits.
// If the initial size is 64 bits or less, it returns the same values as it received.
func calcNextSizes(nBits, byteSize int) (nextByteSize, nextNBits, newByteSize, newNBits int) {
	if byteSize > 8 {
		nextByteSize = byteSize - 8
		nextNBits = nBits - 64
		byteSize = 8
		nBits = 64
	}
	return nextByteSize, nextNBits, byteSize, nBits
}

func reverseSlice[V uint64 | byte](s []V) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}

}

func sizeInWords(width int) (size int) {
	size = width / 64
	aMod := width % 64
	if aMod != 0 {
		size++
	}
	return size
}

// ConvertBytesToWords converts a byte slice into a slice of uint64 words,
// packing the bytes into words based on the specified number of bits.
//
// It first determines the required word and byte sizes based on the input number of bits.
// An error is returned if the input byte slice is smaller than the expected byte size.
// A new uint64 slice is created to hold the words.
//
// Then for each word:
//   - The function calculates the sizes for the next iteration,
//   - Writes the corresponding bytes into the word, and
//   - Slices the input byte slice to remove the already processed bytes.
//
// After all bytes have been processed, any leftover bits in the last word are cleared
// using a mask to ensure that the result matches the specified number of bits.
//
// The function returns the resulting slice of words, or an error if one occurred.
func ConvertBytesToWords(nBits int, val []byte) (words []uint64, err error) {
	wordSize := bitsToWordSize(nBits)
	byteSize := BitsToBytesSize(nBits)
	if err = checkByteSize(byteSize, len(val)); err != nil {
		return words, errors.WithStack(err)
	}

	words = make([]uint64, wordSize)
	nextByteSize := 0
	nextNBits := 0
	modShift := nBits % 64
	lastWordMask := uint64((1 << modShift) - 1)

	for i := range words {
		if nextByteSize != 0 {
			byteSize = nextByteSize
			nBits = nextNBits
		}
		nextByteSize, nextNBits, byteSize, nBits = calcNextSizes(nBits, byteSize)
		if err = writeBytesToUint64Array(i, byteSize, words, val); err != nil {
			return nil, errors.WithStack(err)
		}
		if byteSize < len(val) {
			val = val[byteSize:]
			nBits = nextNBits
			byteSize = nextByteSize
		} else {
			break
		}
	}
	if lastWordMask != 0 {
		words[len(words)-1] &= lastWordMask
	}
	return words, nil
}

// BitsToBytesSize calculates the number of bytes required to accommodate a certain number of bits.
// If the number of bits is not an exact multiple of 8 (since there are 8 bits in a byte),
// it adds one more to the byte count to accommodate the extra bits.
func BitsToBytesSize(in int) int {
	size := in / 8
	if in%8 != 0 {
		size++
	}
	return size
}

// bitsToWordSize calculates the number of 64-bit words needed to accommodate a certain number of bits.
// If the number of bits is not an exact multiple of 64, one more word is added to accommodate the extra bits.
func bitsToWordSize(in int) int {
	size := in / 64
	if in%64 != 0 {
		size++
	}
	return size
}

// checkByteSize checks whether the size of the input byte slice (valSize) is at least as large as a specified size (byteSize).
// If the input slice is smaller, it returns an InvalidInputSliceSizeError annotated with the desired and actual sizes.
func checkByteSize(byteSize, valSize int) (err error) {
	if byteSize > valSize {
		err = InvalidInputSliceSizeError
		err = errors.Wrapf(err, "wanted bytes: %d, input slice sizeInBytes: %d", byteSize, valSize)
		return errors.WithStack(err)
	}
	return nil
}

// writeBytesToUint64Array writes a sequence of bytes into an array of uint64 values at the specified index.
// The byteSize parameter determines the number of bytes to be written.
// The val slice should contain at least byteSize elements.
// The converted uint64 value is stored in the words array at the specified index i.
// The function returns an error if the byteSize is invalid or if the val slice is too small.
func writeBytesToUint64Array(i, byteSize int, words []uint64, val []byte) error {
	// Error checking: ensure `val` has at least `byteSize` elements
	if len(val) < byteSize {
		return errors.WithStack(fmt.Errorf("input slice too small: expected at least %d elements, got %d", byteSize, len(val)))
	}

	// Switch statement for different cases of byte sizes
	switch byteSize {
	case 1:
		words[i] = uint64(val[0]) // Store the first byte as a uint64 value
	case 2:
		words[i] = uint64(binary.LittleEndian.Uint16(val[:2])) // Convert the first 2 bytes to a uint16 value and store as uint64
	case 3, 4:
		var buf [4]byte
		copy(buf[:], val[:byteSize])                          // Copy the bytes to a 4-byte buffer, zero-padding if necessary
		words[i] = uint64(binary.LittleEndian.Uint32(buf[:])) // Convert the buffer to a uint32 value and store as uint64
	case 5, 6, 7, 8:
		var buf [8]byte
		copy(buf[:], val[:byteSize])                  // Copy the bytes to an 8-byte buffer, zero-padding if necessary
		words[i] = binary.LittleEndian.Uint64(buf[:]) // Convert the buffer to a uint64 value and store
	default:
		return errors.WithStack(InvalidBitsSizeError) // Return an error for unsupported byte sizes
	}

	return nil
}
