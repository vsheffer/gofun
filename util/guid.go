package util

import (
	"crypto/rand"
)

// A Guid is a unique 16 byte array.
type Guid struct {
	value string
}

type buffer struct {
	chars []byte
	charIndex int
}

var lookupTable = []byte("0123456789ABCDFGHJKLMNPQRSTVWXYZ")

// Converts the 16 byte array to a 26 character printable string.
func (g *Guid) String() string {
	return g.value
}

func newBuffer(size int) *buffer {
	return &buffer{
		chars: make([]byte, size),
		charIndex: 0,
	}
}

func (buf *buffer) String() string {
	return string(buf.chars)
}

func (buf *buffer) append(b byte) {
	buf.chars[buf.charIndex] = b
	buf.charIndex += 1
}

func numCharactersRequired(numBytes int) int {
	bits := numBytes * 8
	if bits % 5 == 0 {
		return bits / 5
	} else {
		return bits / 5 + 1
	}
}

func bytesToString(bytes []byte) string {
	length := len(bytes)
	numCharactersRequired := numCharactersRequired(length)
	chars := newBuffer(numCharactersRequired)
	encodeBytes(bytes[0], bytes[0], 0, chars)
	for i := 1; i < length; i++ {
		encodeBytes(bytes[i-1], bytes[i], i, chars)
	}
	encodeLastByte(bytes[length-1], length - 1, chars)
	return chars.String()

}

func encodeBytes(previous byte, current byte, byteIndex int, chars *buffer) {
	var chunk byte
	offset := byteIndex % 5
	switch offset {
	case 0:
		chunk = ((current & 0xF8) >> 3)
		chars.append(lookupTable[chunk])
		return
	case 1:
		chunk = ((previous & 0x07) << 2) | ((current & 0xC0) >> 6)
		chars.append(lookupTable[chunk])
		chunk = ((current & 0x3E) >> 1)
		chars.append(lookupTable[chunk])
		return
	case 2:
        chunk = (((previous & 0x01) << 4) | ((current & 0xF0) >> 4))
        chars.append(lookupTable[chunk]);
        return
    case 3:
        chunk = (((previous & 0x0F) << 1) | ((current & 0x80) >> 7))
        chars.append(lookupTable[chunk])
        chunk = ((current & 0x7C) >> 2)
        chars.append(lookupTable[chunk])
        break;
    case 4:
        chunk = (((previous & 0x03) << 3) | ((current & 0xE0) >> 5))
        chars.append(lookupTable[chunk])
        chunk = (current & 0x1F);
        chars.append(lookupTable[chunk]);
        break;
	}
}

func encodeLastByte(last byte, byteIndex int, output *buffer) {
    var chunk byte
    offset := byteIndex % 5
    switch offset {
        case 0:
            chunk = ((last & 0x07) << 2)
            output.append(lookupTable[chunk])
            return
        case 1:
            chunk = ((last & 0x01) << 4)
            output.append(lookupTable[chunk])
            return
        case 2:
            chunk = ((last & 0x0F) << 1)
            output.append(lookupTable[chunk])
            return
        case 3:
            chunk = ((last & 0x03) << 3)
            output.append(lookupTable[chunk]);
            return
        case 4:
            return
    }
}

func NewGuid() (*Guid, error) {
	c := 16
	b := make([]byte, c)

	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	guid := &Guid{
		value: bytesToString(b),
	}

	return guid, nil
}
