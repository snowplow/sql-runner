package pool

import (
	"encoding/binary"
	"io"
)

type WriteBuffer struct {
	w     io.Writer
	Bytes []byte

	msgStart, paramStart int
}

func NewWriteBuffer(w io.Writer, b []byte) *WriteBuffer {
	return &WriteBuffer{
		w:     w,
		Bytes: b,
	}
}

func (buf *WriteBuffer) StartMessage(c byte) {
	if c == 0 {
		buf.msgStart = len(buf.Bytes)
		buf.Bytes = append(buf.Bytes, 0, 0, 0, 0)
	} else {
		buf.msgStart = len(buf.Bytes) + 1
		buf.Bytes = append(buf.Bytes, c, 0, 0, 0, 0)
	}
}

func (buf *WriteBuffer) FinishMessage() {
	binary.BigEndian.PutUint32(
		buf.Bytes[buf.msgStart:], uint32(len(buf.Bytes)-buf.msgStart))
}

func (buf *WriteBuffer) StartParam() {
	buf.paramStart = len(buf.Bytes)
	buf.Bytes = append(buf.Bytes, 0, 0, 0, 0)
}

func (buf *WriteBuffer) FinishParam() {
	binary.BigEndian.PutUint32(
		buf.Bytes[buf.paramStart:], uint32(len(buf.Bytes)-buf.paramStart-4))
}

var nullParamLength = int32(-1)

func (buf *WriteBuffer) FinishNullParam() {
	binary.BigEndian.PutUint32(
		buf.Bytes[buf.paramStart:], uint32(nullParamLength))
}

func (buf *WriteBuffer) Write(b []byte) (int, error) {
	buf.Bytes = append(buf.Bytes, b...)
	return len(b), nil
}

func (buf *WriteBuffer) WriteInt16(num int16) {
	buf.Bytes = append(buf.Bytes, 0, 0)
	binary.BigEndian.PutUint16(buf.Bytes[len(buf.Bytes)-2:], uint16(num))
}

func (buf *WriteBuffer) WriteInt32(num int32) {
	buf.Bytes = append(buf.Bytes, 0, 0, 0, 0)
	binary.BigEndian.PutUint32(buf.Bytes[len(buf.Bytes)-4:], uint32(num))
}

func (buf *WriteBuffer) WriteString(s string) {
	buf.Bytes = append(buf.Bytes, s...)
	buf.Bytes = append(buf.Bytes, 0)
}

func (buf *WriteBuffer) WriteBytes(b []byte) {
	buf.Bytes = append(buf.Bytes, b...)
	buf.Bytes = append(buf.Bytes, 0)
}

func (buf *WriteBuffer) WriteByte(c byte) {
	buf.Bytes = append(buf.Bytes, c)
}

func (buf *WriteBuffer) Flush() error {
	_, err := buf.w.Write(buf.Bytes)
	buf.Reset()
	return err
}

func (buf *WriteBuffer) Reset() {
	buf.Bytes = buf.Bytes[:0]
}

func (buf *WriteBuffer) ReadFrom(r io.Reader) (int64, error) {
	n, err := r.Read(buf.Bytes[len(buf.Bytes):cap(buf.Bytes)])
	buf.Bytes = buf.Bytes[:len(buf.Bytes)+int(n)]
	return int64(n), err
}
