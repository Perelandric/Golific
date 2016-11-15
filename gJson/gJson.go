package gJson

import (
	"bytes"
	"encoding/json"
)

type JSONEncodable interface {
	JSONEncode(*Encoder) bool
}

type Elidable interface {
	CanElide() bool
}

/*
EmbedEncodedStruct panics if the `je` parameter isn't a struct or pointer to a
struct. It adds only the encoded fields to the encoder.
Returns `true` if anything was actually written.
*/
func (e *Encoder) EmbedEncodedStruct(je JSONEncodable, isFirst bool) bool {
	if je == nil {
		return false
	}

	// TODO: Create a pool for this allocation
	var tempE Encoder

	if je.JSONEncode(&tempE) {
		return e.embedResult(tempE.Bytes(), isFirst)
	}
	return false
}

/*
EmbedMarshaledStruct panics if the `m` parameter isn't a struct or pointer to a
struct. It adds only the marshaled fields to the encoder.
Returns `true` if `isFirst` was `true` and nothing was written.
*/
func (e *Encoder) EmbedMarshaledStruct(m interface{}, isFirst bool) bool {
	if m == nil {
		return false
	}

	if r, err := json.Marshal(m); err != nil || len(r) == 0 {
		return false
	} else {
		return e.embedResult(r, isFirst)
	}
}

func (e *Encoder) embedResult(b []byte, isFirst bool) bool {
	res := bytes.TrimSpace(b)

	if len(res) >= 2 && res[0] == '{' && res[len(res)-1] == '}' {
		if toEmbed := bytes.TrimSpace(res[1 : len(res)-1]); len(toEmbed) > 0 {
			if isFirst {
				e.writeByte(',')
			}
			e.write(toEmbed)
			return true
		}
		return false

	} else {
		panic("Expected a struct")
	}
}
