package gobitstream

import (
	"github.com/juju/errors"
)

var InvalidBitsSize = errors.New("invalid bits size")

var InvalidValueSize = errors.New("invalid value size")

var InvalidInputSliceSize = errors.New("invalid input slice size")

var CaseWIP = errors.New("case not supported yet")

var UnexpectedCondition = errors.New("unexpected condition")
