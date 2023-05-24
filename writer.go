package gobitstream

import (
	"encoding/binary"
	"fmt"
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
	//fmt.Println("sizeInBytes: ", wr.sizeInBytes)
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
	//fmt.Printf("---> %X\n", wr.dstWord)
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

func (wr *Writer) WriteNbitsFromBytes(nBits int, xval []byte) (err error) {
	var val []byte

	if !wr.isLittleEndian {
		val = make([]byte, len(xval))
		_ = copy(val, xval)
		reverseSlice(val)
	} else {
		val = xval
	}
	byteSize := BitsToBytesSize(nBits)

	if errx := checkByteSize(byteSize, len(val)); errx != nil {
		return errors.WithStack(errx)
	}

	words, err := ConvertBytesToWords(nBits, val)

	//fmt.Printf("bytes to words: %X -- %d\n", words, nBits)

	err = setFieldToSlice(wr.dstWord, words, uint64(nBits), uint64(wr.offset))

	fmt.Printf("%d -- words: %X\n", nBits, words)

	wr.offset += nBits
	return nil
}

func (wr *Writer) WriteNbitsFromWord(nBits int, val uint64) (err error) {
	if nBits > 64 {
		return InvalidBitsSizeError
	}

	if wr.offset >= 64 {
		err = setFieldToSlice(wr.dstWord, []uint64{val}, uint64(nBits), uint64(wr.offset))
		if err != nil {
			return errors.WithStack(err)
		}
	} else {
		err = set64BitsFieldToWordSlice(wr.dstWord, val, uint64(nBits), uint64(wr.offset))
		if err != nil {
			return errors.WithStack(err)
		}
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

// setFieldToSlice sets a field (represented as a slice of uint64 values)
// to a destination slice (dstSlice) starting at a given bit offset.
// The total width of the field in bits is provided by the 'width' argument.
// Note that if the total width of the field exceeds the length of dstSlice,
// it may lead to a runtime panic.
func setFieldToSlice(dstSlice []uint64, field []uint64, width, offset uint64) (err error) {
	// Compute the number of uint64 values required to store the field
	remainingWidth := width

	// Iterate over each word in the field
	for i, fieldWord := range field {
		localFieldOffset, fieldOffset := computeLocalOffsets(offset, i)

		if fieldOffset >= uint64(len(dstSlice)) {
			return errors.Errorf("fieldOffset: %d is out of range", fieldOffset)
		}

		localDstSlice := dstSlice[fieldOffset:]
		localDstWidth := computeDstWidth(remainingWidth, localFieldOffset, i)

		err = set64BitsFieldToWordSlice(localDstSlice, fieldWord, localDstWidth, localFieldOffset)
		if err != nil {
			return errors.Wrapf(err, "fieldOffset: %d, localDstWidth: %d, localFieldOffset: %d", fieldOffset, localDstWidth, localFieldOffset)
		}

		// Decrease remainingWidth only after successful operation
		remainingWidth -= localDstWidth
	}

	return nil
}
