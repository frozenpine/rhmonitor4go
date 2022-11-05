package rhmonitor4go

/*
#cgo CFLAGS: -I${SRCDIR}/include

#include "cRHMonitorApi.h"
*/
import "C"
import (
	"bytes"
	"io"
	"log"
	"reflect"
	"sync"
	"unsafe"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

func GbkToUtf8(s []byte) ([]byte, error) {
	reader := transform.NewReader(bytes.NewReader(s), simplifiedchinese.GBK.NewDecoder())
	d, e := io.ReadAll(reader)
	if e != nil {
		return nil, e
	}
	return d, nil
}

func CStr2GoStr(in unsafe.Pointer) string {
	gbkData := ([]byte)(C.GoString((*C.char)(in)))

	utf8Data, err := GbkToUtf8(gbkData)

	if err != nil {
		log.Printf("Decode gbk msg faild: %v", err)
	}

	return string(utf8Data)
}

func CopyN(dst []byte, src unsafe.Pointer, len int) {
	tmpSlice := *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{
		Data: uintptr(src),
		Len:  len, Cap: len,
	}))

	copy(dst, tmpSlice)
}

var stringBuffer = sync.Pool{New: func() any { return make([]byte, 0, 100) }}

type StringBuffer struct {
	buffer *bytes.Buffer
	under  []byte
}

func (buf *StringBuffer) Release() {
	stringBuffer.Put(buf.under[:0])
}

func (buf *StringBuffer) WriteString(v string) (int, error) {
	return buf.buffer.WriteString(v)
}

func (buf *StringBuffer) String() string {
	return buf.buffer.String()
}

func NewStringBuffer() *StringBuffer {
	under := stringBuffer.Get().([]byte)

	buf := StringBuffer{
		buffer: bytes.NewBuffer(under),
		under:  under,
	}

	return &buf
}
