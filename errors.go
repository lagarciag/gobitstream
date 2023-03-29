package gobitstream

import (
	"github.com/juju/errors"
)

var InvalidBitsSizeError = errors.New("invalid bits sizeInBytes")

var InvalidValueSizeError = errors.New("invalid value sizeInBytes")

var InvalidInputSliceSizeError = errors.New("invalid input slice sizeInBytes")

var InvalidResultAssertionError = errors.New("invalid result assertion")

var CaseWIPError = errors.New("case not supported yet")

var UnexpectedCondition = errors.New("unexpected condition")
