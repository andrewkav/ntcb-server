package ntcb

import (
	"errors"
	"strings"
)

type BitArray []byte

func (ba BitArray) IsSet(i int) bool {
	if i/8 >= len(ba) {
		return false
	}
	return ba[i/8]&(1<<(7-i%8)) > 0
}

func NewBitArrayFromString(s string) (BitArray, error) {
	if len(s)%8 != 0 {
		s += strings.Repeat("0", 8-len(s)%8)
	}
	ba := make(BitArray, (len(s)+7)/8)

	for i, b := range []byte(s) {
		if b != '0' && b != '1' {
			return nil, errors.New("invalid character")
		}
		ba[i/8] <<= 1
		ba[i/8] += b - '0'
	}

	return ba, nil
}
