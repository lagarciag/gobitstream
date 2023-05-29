package gobitstream

import "github.com/pkg/errors"

func SetFieldToSliceX(dstSlice []uint64, field []uint64, width, offset uint64) error {
	// Check if offset is larger than the size of dstSlice in bits.
	if offset > uint64(len(dstSlice)*64) {
		return errors.New("offset is out of range")
	}

	// Check if width is larger than the size of the field in bits.
	if width > uint64(len(field)*64) {
		return errors.New("width is larger than the size of the field")
	}

	// Handle zero-width case: do nothing and return nil.
	if width == 0 {
		return nil
	}

	// Prepare variables for bit manipulation
	var bitsHandled uint64 = 0
	var fieldIndex uint64 = 0
	var currentWord = field[fieldIndex]

	for bitsHandled < width {
		dstIndex := (offset + bitsHandled) / 64
		if dstIndex >= uint64(len(dstSlice)) {
			return errors.New("dstSlice is not large enough to hold the field")
		}

		bitPos := (offset + bitsHandled) % 64
		bitsAvailableInDst := 64 - bitPos
		bitsRemaining := width - bitsHandled
		bitsToCopy := bitsAvailableInDst
		if bitsRemaining < bitsAvailableInDst {
			bitsToCopy = bitsRemaining
		}

		// Prepare word to be inserted
		insertWord := currentWord & ((1 << bitsToCopy) - 1)
		currentWord = currentWord >> bitsToCopy

		// Insert bits to the destination slice
		dstSlice[dstIndex] &= ^(((1 << bitsToCopy) - 1) << bitPos)
		dstSlice[dstIndex] |= insertWord << bitPos

		bitsHandled += bitsToCopy

		// If current word is exhausted, move to the next
		if currentWord == 0 && fieldIndex < uint64(len(field))-1 {
			fieldIndex++
			currentWord = field[fieldIndex]
		}
	}

	return nil
}

func SetFieldToSlice(dstSlice []uint64, field []uint64, width, offset uint64) error {
	// Check if offset is larger than the size of dstSlice.
	if offset/64 >= uint64(len(dstSlice)) {
		return errors.New("offset is out of range")
	}

	// Check if width is larger than the size of the field.
	if width > uint64(len(field)*64) {
		return errors.New("width is larger than the size of the field")
	}

	// Handle zero-width case: do nothing and return nil.
	if width == 0 {
		return nil
	}

	// Compute the number of uint64 values required to store the field.
	// We need to consider both the remaining width and the offset.
	wordCount := (width + offset + 63) / 64 // round up division
	if wordCount > uint64(len(dstSlice)) {
		return errors.New("dstSlice is not large enough to hold the field")
	}

	//TODO: Performance check here

	tmpSlice := make([]uint64, len(dstSlice))

	_ = copy(tmpSlice, field)

	var err error
	tmpSlice, err = ShiftSliceOfUint64Left(tmpSlice, int(offset))
	if err != nil {
		return errors.WithStack(err)
	}

	for i, d := range tmpSlice {
		dstSlice[i] = dstSlice[i] | d
	}

	return nil
}

func SetFieldToSliceOrg(dstSlice []uint64, field []uint64, width, offset uint64) error {

	//TODO: Performance check here

	tmpSlice := make([]uint64, len(dstSlice))

	for i, f := range field {
		tmpSlice[i] = f
	}

	// Compute the number of uint64 values required to store the field
	remainingWidth := width

	// Check if offset is larger than the size of dstSlice.
	if offset/64 >= uint64(len(dstSlice)) {
		return errors.New("offset is out of range")
	}

	// Check if width is larger than the size of the field.
	if width > uint64(len(field)*64) {
		return errors.New("width is larger than the size of the field")
	}

	// Handle zero-width case: do nothing and return nil.
	if width == 0 {
		return nil
	}

	// Compute the number of uint64 values required to store the field.
	// We need to consider both the remaining width and the offset.
	wordCount := (width + offset + 63) / 64 // round up division
	if wordCount > uint64(len(dstSlice)) {
		return errors.New("dstSlice is not large enough to hold the field")
	}

	// Iterate over each word in the field
	for i, fieldWord := range field {
		currentOffset, index := computeLocalOffsets(offset, i)

		if index >= uint64(len(dstSlice)) {
			return errors.Errorf("index: %d is out of range", index)
		}

		localDstSlice := dstSlice[index:]
		localDstWidth := computeDstWidth(remainingWidth, currentOffset, i)

		var err error
		if localDstSlice, err = Set64BitsFieldToWordSlice(localDstSlice, fieldWord, localDstWidth, currentOffset); err != nil {
			return errors.Wrapf(err, "index: %d, localDstWidth: %d, currentOffset: %d", index, localDstWidth, currentOffset)
		}

		// Decrease remainingWidth only after successful operation
		remainingWidth -= localDstWidth

	}

	if remainingWidth > 0 {
		lastWord := field[len(field)-1]
		calculateShift := 64 - remainingWidth
		lastWord = lastWord >> calculateShift
		dstSlice[len(dstSlice)-1] = dstSlice[len(dstSlice)-1] | lastWord
	}

	return nil
}

// **************************************************

// computeDstWidth calculates the width of the destination slice for the current iteration
func computeDstWidth(remainingWidth, localFieldOffset uint64, i int) (localDstWidth uint64) {
	if remainingWidth > 64 && i == 0 {
		localDstWidth = 64 - localFieldOffset
	} else if remainingWidth >= 64 {
		localDstWidth = 64
	} else {
		localDstWidth = remainingWidth % 64
	}
	return
}

// computeLocalOffsets calculates the local field offset and the field offset for the current iteration
func computeLocalOffsets(offset uint64, i int) (currentOffset, index uint64) {
	currentOffset = (offset + uint64(64*i)) % 64
	index = (offset + uint64(64*i)) / 64
	if i != 0 {
		currentOffset = 0
	}
	return
}
