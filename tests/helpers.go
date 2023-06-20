package tests

import (
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"math/rand"
	"testing"
)

func NewLogger() (l *zap.SugaredLogger) {
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	ul, err := config.Build()
	l = ul.Sugar()
	if err != nil {
		panic("could not instantiate logger")
	}
	return l
}

func InitTest(t *testing.T) (l *zap.SugaredLogger, a *assert.Assertions, r *rand.Rand) {
	t.Logf("namettt: %s", t.Name())
	a = assert.New(t)
	return NewLogger(), a, rand.New(rand.NewSource(0))
}

func MaskLastByteLE(mask uint8, inBytes []byte) {
	if mask != 0 {
		inBytes[len(inBytes)-1] = inBytes[len(inBytes)-1] & mask
	}

}
