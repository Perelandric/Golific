package gJson

import "bytes"

type Encoder struct {
	b bytes.Buffer
}

func (e *Encoder) writeString(s string) {
	_, err := e.b.WriteString(s)
	if err != nil {
		panic(err)
	}
}

func (e *Encoder) writeByte(b byte) {
	err := e.b.WriteByte(b)
	if err != nil {
		panic(err)
	}
}

func (e *Encoder) write(b []byte) {
	_, err := e.b.Write(b)
	if err != nil {
		panic(err)
	}
}

func (e *Encoder) WriteRawString(s string) {
	e.writeString(s)
}

func (e *Encoder) WriteRawByte(b byte) {
	e.writeByte(b)
}

func (e *Encoder) WriteRaw(b []byte) {
	e.write(b)
}

func (e *Encoder) Len() int {
	return e.b.Len()
}

func (e *Encoder) String() string {
	return e.b.String()
}

func (e *Encoder) Bytes() []byte {
	return e.b.Bytes()
}

func (e *Encoder) Truncate(n int) {
	e.b.Truncate(n)
}
