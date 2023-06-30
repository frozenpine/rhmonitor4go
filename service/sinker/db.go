package sinker

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"reflect"
	"regexp"
	"strings"
	"unsafe"

	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/mattn/go-sqlite3"
)

var (
	db          *sql.DB
	connPattern = regexp.MustCompile(
		"(?P<proto>(?:sqlite3|postgres|pgx))://(?P<value>.+)",
	)

	driverMap = map[string]string{
		"postgres": "pgx",
	}
)

func InitDB(conn string) (c *sql.DB, err error) {
	matchs := connPattern.FindStringSubmatch(conn)
	if len(matchs) < 1 {
		return nil, errors.New("invalid db conn string")
	}

	protoIdx := connPattern.SubexpIndex("proto")
	valueIdx := connPattern.SubexpIndex("value")

	var (
		proto string
		exist bool
	)

	if proto, exist = driverMap[matchs[protoIdx]]; !exist {
		proto = matchs[protoIdx]
	}

	log.Print("Try to open db: ", conn)

	if c, err = sql.Open(proto, matchs[valueIdx]); err != nil {
		log.Fatalf("Parse database[%s] failed: %+v", conn, err)
	} else if err = c.Ping(); err != nil {
		log.Fatalf("Open database[%s] failed: %+v", conn, err)
	}

	// 涉及PgSQL的分区表及SQLITE, 取消表创建逻辑, 改由外部处理
	// if _, err = c.Exec(acctStructuresSQL); err != nil {
	// 	log.Fatalf(
	// 		"Create table failed: %s, %s\n%s",
	// 		conn, err, acctStructuresSQL,
	// 	)
	// }

	db = c

	return
}

type fieldOffset struct {
	offset uintptr
	typ    reflect.Type
}

func InsertDB[T any](
	ctx context.Context,
	tblName string,
	argNames ...string,
) (func(v *T) (sql.Result, error), error) {
	if tblName == "" {
		return nil, errors.New("no table name")
	}

	obj := new(T)
	typ := reflect.TypeOf(obj).Elem()

	argLen := len(argNames)

	fieldOffsets := make([]fieldOffset, argLen)
	sqlFields := make([]string, argLen)
	argList := make([]string, argLen)

	if argLen == 0 {
		for i := 0; i < typ.NumField(); i++ {
			field := typ.Field(i)
			sqlField := field.Tag.Get("sql")

			if sqlField == "" {
				continue
			}

			fieldOffsets = append(fieldOffsets, fieldOffset{
				offset: field.Offset,
				typ:    field.Type,
			})

			sqlFields = append(sqlFields, sqlField)
			argList = append(argList, fmt.Sprintf("$%d", i+1))
		}
	} else {
		for idx, name := range argNames {
			if field, ok := typ.FieldByName(name); !ok {
				return nil, fmt.Errorf("%s has no field name: %s", typ.Name(), name)
			} else {
				sqlField := field.Tag.Get("sql")
				if sqlField == "" {
					return nil, fmt.Errorf("%s has no sql tag", field.Name)
				}

				fieldOffsets[idx] = fieldOffset{
					offset: field.Offset,
					typ:    field.Type,
				}

				sqlFields[idx] = sqlField
				argList[idx] = fmt.Sprintf("$%d", idx+1)
			}
		}
	}

	sqlTpl := fmt.Sprintf(
		"INSERT INTO %s(%s) VALUES (%s);",
		tblName,
		strings.Join(sqlFields, ","),
		strings.Join(argList, ","),
	)

	getArgList := func(v interface{}) []interface{} {
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

	// tx, err := db.Begin()

	// insStmt, err := tx.Prepare(sqlTpl)
	// if err != nil {
	// 	log.Fatalf("Prepare sql \"%s\" failed: %+v", sqlTpl, err)
	// }

	return func(v *T) (sql.Result, error) {
		values := getArgList(v)
		// defer tx.Commit()

		// log.Printf("Executing: %s, %s, %+v", sqlTpl, values, v)

		c, cancel := context.WithCancel(ctx)
		defer cancel()

		return db.ExecContext(c, sqlTpl, values...)
		// return insStmt.ExecContext(ctx, values...)
	}, nil
}
