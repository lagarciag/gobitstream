package gobitstream

import (
	"encoding/binary"
	"github.com/juju/errors"
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
	//fmt.Printf("---> %X\n", wr.dstWord)
	wr.dst, err = convertWordsToBytes(wr.dstWord, wr.dst, wr.offset, wr.isLittleEndian)
	if err != nil {
		err = errors.Trace(err)
	}
	return err
}

func convertWordsToBytes(words []uint64, outBuffer []byte, sizeInBits int, isLittleEndian bool) ([]byte, error) {
	sizeInBytes := BitsToBytesSize(sizeInBits)
	if isLittleEndian {
		for i, word := range words {
			binary.LittleEndian.PutUint64(outBuffer[i*8:i*8+8], word)
		}

		outBuffer = outBuffer[0:sizeInBytes]
		return outBuffer, nil
	}
	for i, _ := range words {
		binary.BigEndian.PutUint64(outBuffer[i*8:i*8+8], words[len(words)-1-i])
	}
	outBuffer = outBuffer[len(outBuffer)-sizeInBytes:]

	return outBuffer, nil
}

func (wr *Writer) WriteNbitsFromBytes(nBits int, val []byte) (err error) {
	byteSize := BitsToBytesSize(nBits)

	if errx := checkByteSize(byteSize, len(val)); errx != nil {
		return errors.Trace(errx)
	}

	words, err := ConvertBytesToWords(nBits, val)

	//fmt.Printf("bytes to words: %X -- %d\n", words, nBits)

	err = setFieldToSlice(wr.dstWord, words, uint64(nBits), uint64(wr.offset))
	wr.offset += nBits
	return nil
}

func (wr *Writer) WriteNbitsFromWord(nBits int, val uint64) (err error) {
	if nBits > 64 {
		return InvalidBitsSizeError
	}

	if wr.offset > 64 {
		err = setFieldToSlice(wr.dstWord, []uint64{val}, uint64(nBits), uint64(wr.offset))
		if err != nil {
			return errors.Trace(err)
		}
	} else {
		err = set64BitsFieldToWordSlice(wr.dstWord, val, uint64(nBits), uint64(wr.offset))
		if err != nil {
			return errors.Trace(err)
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
func set64BitsFieldToWordSlice(dstSlice []uint64, field, width, offset uint64) error {
	if offset > 64 {
		err := InvalidOffsetError
		errors.Annotatef(err, "offset must be less than 64, got %d", offset)
		return errors.Trace(err)
	}

	if width == 0 || width > 64 {
		err := InvalidBitsSizeError
		errors.Annotatef(err, "width must be between 1 and 64, got %d", width)
		return errors.Trace(err)
	} else {
		field &= (1 << width) - 1
	}
	if offset >= uint64(len(dstSlice))*64 {
		err := OffsetOutOfRangeError
		errors.Annotatef(err, "offset: %d", offset)
		return errors.Trace(err)
	}

	wordSpan := (width + offset) / 64
	if (width+offset)%64 != 0 {
		wordSpan++
	}
	dstSlice[0] = dstSlice[0] | (field << offset)

	if wordSpan > 1 {
		dstSlice[1] = dstSlice[1] | (field >> (64 - offset))
	}

	// Return the final result and no error
	return nil
}

func setFieldToSlice(dstSlice []uint64, field []uint64, width, offset uint64) (err error) {
	// Compute the number of uint64 values required to store the field
	localDstWidth := width
	remainingWidth := width
	widthWords := int(width / 64)
	if width%64 > 0 {
		widthWords++
	}
	//localFieldOffset := offset % 64
	//fieldOffset := offset / 64
	//localDstSlice := dstSlice[fieldOffset:]

	for i, fieldWord := range field {
		localFieldOffset := (offset + uint64(64*i)) % 64
		fieldOffset := (offset + uint64(64*i)) / 64
		localDstSlice := dstSlice[fieldOffset:]
		if i != 0 {
			localFieldOffset = 0
		}
		if remainingWidth > 64 && i == 0 {
			localDstWidth = 64 - localFieldOffset
		} else if remainingWidth >= 64 {
			//if remainingWidth >= 64 {
			localDstWidth = 64
		} else {
			localDstWidth = width % 64
		}

		//for _, fieldWord := range field {
		//fmt.Printf("fieldWord: %X --\n", fieldWord)
		err := set64BitsFieldToWordSlice(localDstSlice, fieldWord, localDstWidth, localFieldOffset)
		if err != nil {
			err = errors.Annotatef(err, "fieldOffset: %d, localDstWidth: %d, localFieldOffset: %d", fieldOffset, localDstWidth, localFieldOffset)
			return errors.Trace(err)
			//}
			remainingWidth -= localDstWidth
		}

	}
	err = shiftSliceofUint64(field, int(offset%64))
	return nil
}
