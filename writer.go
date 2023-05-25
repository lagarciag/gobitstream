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
		err = set64BitsFieldToWordSlice(wr.dstWord, val, uint64(nBits), uint64(wr.offset))
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
// validateParameters performs the initial checks on the parameters of set64BitsFieldToWordSlice
func validateParameters(dstSlice []uint64, width, offset uint64) error {
	if offset >= 64 {
		return errors.Errorf("offset must be less than 64, got %d", offset)
	}
	if width == 0 || width > 64 {
		return errors.Errorf("width must be between 1 and 64, got %d", width)
	}
	if offset >= uint64(len(dstSlice))*64 {
		return errors.Errorf("offset: %d is out of range", offset)
	}
	return nil
}

// calculateWordSpan calculates the number of words the field spans
func calculateWordSpan(width, offset uint64) uint64 {
	wordSpan := (width + offset) / 64
	if (width+offset)%64 != 0 {
		wordSpan++
	}
	return wordSpan
}

// setFieldInSlice sets the field in the given slice
func setFieldInSlice(dstSlice []uint64, field, offset uint64, wordSpan uint64) error {
	dstSlice[0] = dstSlice[0] | (field << offset)
	if wordSpan > 1 {
		if len(dstSlice) < 2 {
			return errors.New("dstSlice is not large enough to hold the result")
		}
		dstSlice[1] = dstSlice[1] | (field >> (64 - offset))
	}
	return nil
}

// set64BitsFieldToWordSlice sets a 64-bit field in a slice of uint64 words.
// This function receives a destination slice (dstSlice), a field value (field),
// the width of the field in bits (width) and the offset in bits from the beginning of the slice (offset).
// If the offset or width are not valid, it returns an error.
// If the field spans more than one word in the slice and the slice is not large enough to hold the result,
// it also returns an error.
func set64BitsFieldToWordSlice(dstSlice []uint64, field, width, offset uint64) error {
	err := validateParameters(dstSlice, width, offset)
	if err != nil {
		return errors.WithStack(err)
	}

	field &= (1 << width) - 1
	wordSpan := calculateWordSpan(width, offset)

	err = setFieldInSlice(dstSlice, field, offset, wordSpan)
	if err != nil {
		return errors.WithStack(err)
	}

	// Return the final result and no error
	return nil
}

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
func computeLocalOffsets(offset uint64, i int) (localFieldOffset, fieldOffset uint64) {
	localFieldOffset = (offset + uint64(64*i)) % 64
	fieldOffset = (offset + uint64(64*i)) / 64
	if i != 0 {
		localFieldOffset = 0
	}
	return
}

func SetFieldToSlice(dstSlice []uint64, field []uint64, width, offset uint64) error {
	// Compute the number of uint64 values required to store the field
	remainingWidth := width

	// If width is zero, return nil immediately
	if width == 0 {
		return nil
	}

	// Check if width is larger than the size of the field
	if width > uint64(len(field)*64) {
		return errors.Errorf("width: %d is larger than the size of the field", width)
	}

	// Iterate over each word in the field
	for i, fieldWord := range field {
		localFieldOffset, fieldOffset := computeLocalOffsets(offset, i)

		if fieldOffset >= uint64(len(dstSlice)) {
			return errors.Errorf("fieldOffset: %d is out of range", fieldOffset)
		}

		localDstSlice := dstSlice[fieldOffset:]
		localDstWidth := computeDstWidth(remainingWidth, localFieldOffset, i)

		if err := set64BitsFieldToWordSlice(localDstSlice, fieldWord, localDstWidth, localFieldOffset); err != nil {
			return errors.Wrapf(err, "fieldOffset: %d, localDstWidth: %d, localFieldOffset: %d", fieldOffset, localDstWidth, localFieldOffset)
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
