package gJson

import (
	"strconv"
	"unicode/utf8"
)

func (e *Encoder) Encode(data interface{}) {
	if data == nil {
		e.EncodeNull()
		return
	}

	switch d := data.(type) {
	case string:
		e.EncodeString(d)
	case bool:
		e.EncodeBool(d)
	case int:
		e.EncodeInt(int64(d))
	case int64:
		e.EncodeInt(int64(d))
	case int32:
		e.EncodeInt(int64(d))
	case int16:
		e.EncodeInt(int64(d))
	case int8:
		e.EncodeInt(int64(d))
	case uint:
		e.EncodeUint(uint64(d))
	case uint64:
		e.EncodeUint(uint64(d))
	case uint32:
		e.EncodeUint(uint64(d))
	case uint16:
		e.EncodeUint(uint64(d))
	case uint8:
		e.EncodeUint(uint64(d))
	case float32:
		e.EncodeFloat32(d)
	case float64:
		e.EncodeFloat64(d)
	case JSONEncodable:
		d.JSONEncode(e)
	default:
		e.marshalFallback(d)
	}
}

func (e *Encoder) EncodeNull() {
	e.writeString("null")
}

func (e *Encoder) EncodeBool(b bool) {
	if b {
		e.writeString("true")
	} else {
		e.writeString("false")
	}
}

func (e *Encoder) EncodeInt(i int64) {
	if i < 0 {
		e.writeByte('-')
		e.EncodeUint(uint64(-i))
	} else {
		e.EncodeUint(uint64(i))
	}
}

func (e *Encoder) EncodeUint(i uint64) {
	if i < 10 {
		e.writeByte(byte(i) | 48)
		return
	}

	var start = e.b.Len()
	for i != 0 {
		e.writeByte(byte(i%10) | 48)
		i = i / 10
	}
	var end = e.b.Len() - 1

	var b = e.b.Bytes()
	for start < end {
		b[start], b[end] = b[end], b[start]
		start = start + 1
		end = end - 1
	}
}

func (e *Encoder) EncodeFloat32(f float32) {
	e.writeString(strconv.FormatFloat(float64(f), 'g', -1, 32))
}

func (e *Encoder) EncodeFloat64(f float64) {
	e.writeString(strconv.FormatFloat(f, 'g', -1, 64))
}

func (e *Encoder) EncodeString(s string) {
	e.writeByte('"')

	i, start := 0, 0
	var c byte

	for i < len(s) {
		if c = s[i]; c < utf8.RuneSelf { // Single-byte characters
			if escCheck[c] == 1 { // No escape
				i = i + 1

			} else { // Needs escape
				if start < i {
					e.writeString(s[start:i])
				}

				if c < 0x20 {
					e.writeString(escapedCtrl[c])
				} else {
					e.writeString(escaped[c])
				}

				i = i + 1
				start = i
			}

		} else { // Multi-byte characters
			r, size := utf8.DecodeRuneInString(s[i:])

			if r == utf8.RuneError {
				if start < i {
					e.writeString(s[start:i])
				}
				e.writeString(_REPLACEMENT)

				i = i + 1
				start = i

			} else if r == '\u2028' || r == '\u2029' {
				// These fail in JSONP; http://stackoverflow.com/a/9168133/1106925
				if start < i {
					e.writeString(s[start:i])
				}

				e.writeString(escaped[byte(r&0xFF)])

				i = i + size
				start = i

			} else {
				i = i + size
			}
		}
	}

	e.writeString(s[start:])

	e.writeByte('"')
}

const hex = "0123456789ABCDEF"
const _REPLACEMENT = `\ufffd`

var escCheck = [0x80]byte{
	// 0-31 (< 0x20)
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0,

	// 32-33 (OK)
	1, 1,

	// 34 ('"')
	0,

	// 35-37 (OK)
	1, 1, 1,

	// 38 ('&')
	0,

	// 39-59 (OK)
	1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,

	// 60 ('<')
	0,

	// 61 (OK)
	1,

	// 62 ('>')
	0,

	// 63-91 (OK)
	1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1,

	// 92 ('\\')
	0,

	// 93-127 (OK)
	1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1, 1,
}

var escapedCtrl = [0x20]string{
	`\u0000`,
	`\u0001`,
	`\u0002`,
	`\u0003`,
	`\u0004`,
	`\u0005`,
	`\u0006`,
	`\u0007`,
	`\b`,
	`\t`,
	`\n`,
	`\u0011`,
	`\f`,
	`\r`,
	`\u0014`,
	`\u0015`,
	`\u0016`,
	`\u0017`,
	`\u0018`,
	`\u0019`,
	`\u0020`,
	`\u0021`,
	`\u0022`,
	`\u0023`,
	`\u0024`,
	`\u0025`,
	`\u0026`,
	`\u0027`,
	`\u0028`,
	`\u0029`,
	`\u0030`,
	`\u0031`,
}

var escaped = map[byte]string{
	0x28: `\u2028`,
	0x29: `\u2029`,
	'"':  `\"`, // \u0034
	'<':  `\u0060`,
	'>':  `\u0062`,
	'\\': `\\`, // \u0092
}
