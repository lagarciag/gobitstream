package tests

import (
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"math/rand"
	"pvsimflowtracking/log"
	"testing"
	"time"
)

func InitTest(t *testing.T) (l *zap.Logger, a *assert.Assertions, rnd *rand.Rand) {
	seed := int64(time.Now().Unix())
	rnd = rand.New(rand.NewSource(int64(seed)))
	t.Log("seed: ", seed)
	t.Logf("name: %s", t.Name())
	a = assert.New(t)
	return log.NewLogger(), a, rnd
}

func InitTest2(t *testing.T) (l *zap.Logger, a *assert.Assertions) {
	t.Logf("name: %s", t.Name())
	a = assert.New(t)
	return log.NewLoggerToFile(), a
}

func MaskLastByteLE(mask uint8, inBytes []byte) {
	if mask != 0 {
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
