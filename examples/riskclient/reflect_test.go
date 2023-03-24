package main_test

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/frozenpine/rhmonitor4go/service"
)

var (
	acct = service.Account{
		PreBalance:     10000000,
		Deposit:        5000000,
		Withdraw:       1000000,
		Balance:        14000000,
		PositionProfit: 1,
		CloseProfit:    10,
	}

	fieldNames = []string{
		"Balance", "PreBalance", "Deposit",
		"Withdraw", "CloseProfit", "PositionProfit",
	}
)

func TestFieldReflect(t *testing.T) {

	typ := reflect.TypeOf(new(service.Account)).Elem()

	fieldOffsets := []struct {
		offset uintptr
		typ    reflect.Type
	}{}

	for _, name := range fieldNames {
		if field, ok := typ.FieldByName(name); !ok {
			t.Fatalf("%s has no field name: %s", typ.Name(), name)
		} else {
			fieldOffsets = append(fieldOffsets, struct {
				offset uintptr
				typ    reflect.Type
			}{
				offset: field.Offset,
				typ:    field.Type,
			})
		}
	}

	getArgList := func(v *service.Account) []interface{} {
		basePtr := uintptr(unsafe.Pointer(v))

		argList := []interface{}{}

		for _, field := range fieldOffsets {
			argList = append(
				argList,
				reflect.Indirect(reflect.NewAt(
					field.typ, unsafe.Pointer(basePtr+field.offset),
				)).Interface(),
			)
		}

		return argList
	}

	v := reflect.NewAt(
		reflect.TypeOf(acct.PreBalance),
		unsafe.Pointer(
			reflect.Indirect(reflect.ValueOf(&acct)).Addr().Pointer()+fieldOffsets[1].offset,
		),
	)
	t.Log(v, &acct.PreBalance)

	for idx, v := range getArgList(&acct) {
		f := reflect.Indirect(reflect.ValueOf(acct)).FieldByName(fieldNames[idx])
		t.Logf("%v, %+v, %v", reflect.TypeOf(v), v, f)
	}
}

func BenchmarkReflect(b *testing.B) {
	getter := func() func(*service.Account) []interface{} {
		typ := reflect.TypeOf(new(service.Account)).Elem()

		fieldOffsets := []struct {
			offset uintptr
			typ    reflect.Type
		}{}

		for _, name := range fieldNames {
			if field, ok := typ.FieldByName(name); !ok {
				b.Fatalf("%s has no field name: %s", typ.Name(), name)
			} else {
				fieldOffsets = append(fieldOffsets, struct {
					offset uintptr
					typ    reflect.Type
				}{
					offset: field.Offset,
					typ:    field.Type,
				})
			}
		}

		return func(v *service.Account) []interface{} {
			basePtr := reflect.Indirect(reflect.ValueOf(v)).Addr().Pointer()

			argList := []interface{}{}

			for _, field := range fieldOffsets {
				argList = append(
					argList,
					reflect.Indirect(reflect.NewAt(
						field.typ, unsafe.Pointer(basePtr+field.offset),
					)).Interface(),
				)
			}

			return argList
		}
	}()

	b.Run("reflect", func(b *testing.B) {
		v := reflect.Indirect(reflect.ValueOf(&acct))

		for i := 0; i < b.N; i++ {
			result := []interface{}{}

			for _, name := range fieldNames {
				result = append(result, v.FieldByName(name).Interface())
			}

			// b.Log(result...)
		}
	})

	b.Run("reflectX", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			// result := getter(&acct)
			// b.Log(result...)
			getter(&acct)
		}
	})
}
