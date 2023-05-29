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
		wr.dstWord, err = Set64BitsFieldToWordSlice(wr.dstWord, val, uint64(nBits), uint64(wr.offset))
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
