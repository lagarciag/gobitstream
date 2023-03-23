package tests

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"pvsimflowtracking/log"
	"testing"
)

func InitTest(t *testing.T) (l *zap.Logger, a *assert.Assertions) {
	t.Logf("name: %s", t.Name())
	a = assert.New(t)
	return log.NewLogger(), a
}

func InitTest2(t *testing.T) (l *zap.Logger, a *assert.Assertions) {
	t.Logf("name: %s", t.Name())
	a = assert.New(t)
	return log.NewLoggerToFile(), a
}

func SizeAndMask(width uint) (size uint, mask uint8) {
	size = width / 8
	aMod := width % 8
	mask = 0

	if aMod != 0 {
		size++
		mask = ((1 << aMod) - 1) << (8 - aMod)
	}

	return size, mask

}

func MaskLastByteBE(mask uint8, inBytes []byte) {
	if len(inBytes)%8 != 0 {
		fmt.Println("mask: ", mask)
		inBytes[len(inBytes)-1] = inBytes[len(inBytes)-1] & mask
	}

}

func SizeInBytes(width uint) (size uint) {
	size = width / 8
	aMod := width % 8
	if aMod != 0 {
		size++
	}
	return size
}

func FixInBytesSize(widthInBytes int, inBytes []byte) []byte {
	if len(inBytes) < widthInBytes {
		revCount := widthInBytes - 1
		inTmp := make([]byte, widthInBytes)
		for _, data := range inBytes {
			inTmp[revCount] = data
		}
		return inTmp
	}
	return inBytes
}
