package gJson

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"unicode/utf8"
)

// EncodeKeyVal writes the provide key/value to the encoder, with a leading
// comma if `isFirst` is `false`.
// The pair is not written if `canElide` is `true` and the value provided is a
// zero value, is a JSONEncoder that did returned `false`, or a `Marshaler` that
// panic'd or did not write anything.
func (e *Encoder) EncodeKeyVal(k string, v interface{}, isFirst, canElide bool) bool {
	var pos = e.b.Len()

	fmt.Printf("Encoding: %s, %#v\n\n", k, v)

	if !isFirst {
		e.writeByte(',')
	}

	e.EncodeString(k, false)
	e.writeByte(':')

	if e.Encode(v, canElide) == false {
		e.b.Truncate(pos)
		return false
	}
	return true
}

func (e *Encoder) Encode(data interface{}, canElide bool) bool {
	if data == nil {
		return e.EncodeNull(canElide)
	}

	switch d := data.(type) {
	case string:
		return e.EncodeString(d, canElide)
	case bool:
		return e.EncodeBool(d, canElide)

	case int:
		return e.EncodeInt(int64(d), canElide)
	case int64:
		return e.EncodeInt(int64(d), canElide)
	case int32:
		return e.EncodeInt(int64(d), canElide)
	case int16:
		return e.EncodeInt(int64(d), canElide)
	case int8:
		return e.EncodeInt(int64(d), canElide)

	case uint:
		return e.EncodeUint(uint64(d), canElide)
	case uint64:
		return e.EncodeUint(uint64(d), canElide)
	case uint32:
		return e.EncodeUint(uint64(d), canElide)
	case uint16:
		return e.EncodeUint(uint64(d), canElide)
	case uint8:
		return e.EncodeUint(uint64(d), canElide)

	case float32:
		return e.EncodeFloat32(d, canElide)
	case float64:
		return e.EncodeFloat64(d, canElide)

	default:
		if canElide {
			if de, ok := data.(Elidable); ok && de.CanElide() {
				return false
			}

			if de, ok := data.(Zeroable); ok && de.IsZero() {
				return false
			}

			v := reflect.ValueOf(data)

			if v.CanInterface() && v.IsNil() {
				return false
			}

			switch v.Kind() {
			case reflect.Array, reflect.Slice, reflect.Map:
				if v.Len() == 0 {
					return false
				}
			case reflect.Interface, reflect.Ptr:
				if v.IsNil() {
					return false
				}
			}

			if v.CanAddr() {
				itf := v.Addr().Interface()

				if de, ok := itf.(Elidable); ok && de.CanElide() {
					return false
				}
				if de, ok := itf.(Zeroable); ok && de.IsZero() {
					return false
				}
			}
		}

		if je, ok := data.(JSONEncodable); ok {
			if !je.JSONEncode(e) {
				return e.EncodeNull(canElide)
			}
			return true
		}
	}

	return e.marshalFallback(data, canElide)
}

func (e *Encoder) marshalFallback(d interface{}, canElide bool) bool {
	fmt.Printf("Marshal fallback on: %#v\n\n", d)

	if b, err := json.Marshal(d); err == nil && len(b) > 0 {
		e.write(b)
		return true
	}
	return e.EncodeNull(canElide)
}

func (e *Encoder) EncodeNull(canElide bool) bool {
	if canElide {
		return false
	}
	e.writeString("null")
	return true
}

/*
func (e *Encoder) EncodeStruct(s interface{}, canElide bool) bool {

}
*/

func (e *Encoder) EncodeBool(b bool, canElide bool) bool {
	if b {
		e.writeString("true")
	} else {
		if canElide {
			return false
		}
		e.writeString("false")
	}
	return true
}

func (e *Encoder) EncodeInt(i int64, canElide bool) bool {
	if i < 0 {
		e.writeByte('-')
		return e.EncodeUint(uint64(-i), canElide)
	}
	return e.EncodeUint(uint64(i), canElide)
}

func (e *Encoder) EncodeUint(i uint64, canElide bool) bool {
	if canElide && i == 0 {
		return false
	}

	if i < 10 {
		e.writeByte(byte(i) | 48)
		return true
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

	return true
}

func (e *Encoder) EncodeFloat32(f float32, canElide bool) bool {
	if canElide && f == 0 {
		return false
	}
	e.writeString(strconv.FormatFloat(float64(f), 'g', -1, 32))
	return true
}

func (e *Encoder) EncodeFloat64(f float64, canElide bool) bool {
	if canElide && f == 0 {
		return false
	}
	e.writeString(strconv.FormatFloat(f, 'g', -1, 64))
	return true
}

func (e *Encoder) EncodeString(s string, canElide bool) bool {
	if canElide && s == "" {
		return false
	}
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

	return true
}

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
