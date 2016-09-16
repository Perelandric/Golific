package gJson

import (
	"bytes"
	"encoding/json"
)

type JSONEncodable interface {
	JSONEncode(e *Encoder)
}

type Zeroable interface {
	IsZero() bool
}

/*
EmbedEncodedStruct panics if the `je` parameter isn't a struct or pointer to a
struct. It adds only the encoded fields to the encoder.
Returns `true` if at least one field was embedded.
*/
func (e *Encoder) EmbedEncodedStruct(je JSONEncodable) bool {
	if je == nil {
		return false
	}

	// TODO: Create a pool for this allocation
	var tempE Encoder
	je.JSONEncode(&tempE)

	return e.embedResult(tempE.Bytes())
}

/*
EmbedMarshaledStruct panics if the `m` parameter isn't a struct or pointer to a
struct. It adds only the marshaled fields to the encoder.
Returns `true` if at least one field was embedded.
*/
func (e *Encoder) EmbedMarshaledStruct(m interface{}) bool {
	if m == nil {
		return false
	}

	if r, err := json.Marshal(m); err != nil {
		panic(err)
	} else {
		return e.embedResult(r)
	}
}

func (e *Encoder) embedResult(b []byte) bool {
	res := bytes.TrimSpace(b)

	if len(res) >= 2 && res[0] == '{' && res[len(res)-1] == '}' {
		if toEmbed := bytes.TrimSpace(res[1 : len(res)-1]); len(toEmbed) > 0 {
			e.write(toEmbed)
			return true
		}
		return false

	} else {
		panic("Expected a struct")
	}
}
