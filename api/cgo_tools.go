package api

/*
#cgo CFLAGS: -I${SRCDIR}/include -I${SRCDIR}/cRHMonitorApi

#include "cRHMonitorApi.h"
*/
import "C"
import (
	"bytes"
	"io"
	"log"
	"reflect"
	"unsafe"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

var logger = log.New(
	log.Default().Writer(),
	"[rhmonitor4go.api] ", log.Default().Flags(),
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
		logger.Printf("Decode gbk msg faild: %v", err)
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
