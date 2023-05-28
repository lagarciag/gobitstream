package gobitstream

import (
	"encoding/binary"
	"github.com/pkg/errors"
)

// Writer is a bit stream writer.
// It does not have io.Writer interface
type Writer struct {
	dst            []byte
	dstWord        []uint64
	currWordIndex  uint64
	currBitIndex   uint // MSB: 7, LSB: 0
	offset         int
	size           int
	accBits        int
	sizeInBytes    int
	sizeInWords    int
	isLittleEndian bool
}

func newWriter(totalBits int) *Writer {
	wr := &Writer{}
	wr.size = totalBits
	wr.sizeInWords = bitsToWordSize(totalBits)
	wr.dstWord = make([]uint64, wr.sizeInWords)
	wr.sizeInBytes = BitsToBytesSize(totalBits)
	return wr
}

// NewWriterLE creates a new Writer instance.
func NewWriterLE(totalBits int) *Writer {
	wr := newWriter(totalBits)
	wr.isLittleEndian = true
	return wr
}

func NewWriterBE(totalBits int) *Writer {
	wr := newWriter(totalBits)
	wr.isLittleEndian = false
	return wr
}

func (wr *Writer) Flush() (err error) {
	sizeInBytes := len(wr.dstWord) * 8
	wr.dst = make([]byte, sizeInBytes)
	sizeInBytes = BitsToBytesSize(wr.offset)
	wr.dst, err = convertWordsToBytes(wr.dstWord, wr.dst, wr.offset, wr.isLittleEndian)
	if err != nil {
		err = errors.WithStack(err)
	}
	return err
}

func convertWordsToBytes(words []uint64, outBuffer []byte, sizeInBits int, isLittleEndian bool) ([]byte, error) {
	sizeInBytes := BitsToBytesSize(sizeInBits)
	//if isLittleEndian {
	for i, word := range words {
		binary.LittleEndian.PutUint64(outBuffer[i*8:i*8+8], word)
	}

	outBuffer = outBuffer[0:sizeInBytes]
	if !isLittleEndian {
		outBuffer2 := make([]byte, len(outBuffer))
		_ = copy(outBuffer2, outBuffer)
		reverseSlice(outBuffer2)
		return outBuffer2, nil
	}
	return outBuffer, nil
}

// WriteNbitsFromBytes writes a specified number of bits from a byte slice to the writer's destination.
// The nBits parameter determines the number of bits to write.
// The xval byte slice contains the input bytes to be written.
// If the writer's endianness is not little endian, the byte order is reversed before writing.
// The function returns an error if the byte size is invalid or if there was an error during the field assignment.
func (wr *Writer) WriteNbitsFromBytes(nBits int, xval []byte) error {
	var val []byte

	// Reverse byte order if the writer's endianness is not little endian
	if !wr.isLittleEndian {
		val = make([]byte, len(xval))
		_ = copy(val, xval)
		reverseSlice(val)
	} else {
		val = xval
	}

	byteSize := BitsToBytesSize(nBits)
	if err := checkByteSize(byteSize, len(val)); err != nil {
		return errors.WithStack(err)
	}

	words, errConv := ConvertBytesToWords(nBits, val)
	if errConv != nil {
		return errors.WithStack(errConv)
	}

	if errSet := SetFieldToSlice(wr.dstWord, words, uint64(nBits), uint64(wr.offset)); errSet != nil {
		return errors.WithStack(errSet)
	}

	wr.offset += nBits
	return nil
}

// WriteNbitsFromWord writes a specified number of bits from a uint64 value to the writer's destination.
// The nBits parameter determines the number of bits to write.
// The val parameter is the uint64 value to be written.
// If the number of bits exceeds 64, the function returns an error.
// The function performs a field assignment based on the current writer's offset.
// It returns an error if there was an error during the field assignment.
func (wr *Writer) WriteNbitsFromWord(nBits int, val uint64) error {
	if nBits > 64 {
		return errors.New("invalid number of bits: exceeds 64")
	}

	var err error

	if wr.offset >= 64 {
		err = SetFieldToSlice(wr.dstWord, []uint64{val}, uint64(nBits), uint64(wr.offset))
	} else {
		err = Set64BitsFieldToWordSlice(wr.dstWord, val, uint64(nBits), uint64(wr.offset))
	}

	if err != nil {
		return errors.WithStack(err)
	}

	wr.offset += nBits
	return nil
}

func (wr *Writer) CurrentWord() []uint64 {
	return wr.dstWord
}

func (wr *Writer) Bytes() []byte {
	return wr.dst
}

func (wr *Writer) Words() []uint64 {
	return wr.dstWord
}

func (wr *Writer) Uint64() uint64 {
	return wr.CurrentWord()[0]
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

		if err := Set64BitsFieldToWordSlice(localDstSlice, fieldWord, localDstWidth, currentOffset); err != nil {
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
