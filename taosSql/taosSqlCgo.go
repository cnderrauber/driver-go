/*
 * Copyright (c) 2019 TAOS Data, Inc. <jhtao@taosdata.com>
 *
 * This program is free software: you can use, redistribute, and/or modify
 * it under the terms of the GNU Affero General Public License, version 3
 * or later ("AGPL"), as published by the Free Software Foundation.
 *
 * This program is distributed in the hope that it will be useful, but WITHOUT
 * ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
 * FITNESS FOR A PARTICULAR PURPOSE.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program. If not, see <http://www.gnu.org/licenses/>.
 */
package taosSql

/*
#cgo CFLAGS: -IC:/TDengine/include -I/usr/include
#cgo linux LDFLAGS: -L/usr/lib -ltaos
#cgo windows LDFLAGS: -LC:/TDengine/driver -ltaos
#cgo darwin LDFLAGS: -L/usr/local/taos/driver -ltaos
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <taos.h>
*/
import "C"

import (
	"unsafe"

	"github.com/taosdata/driver-go/errors"
)

type UserChar struct {
	Str *C.char
	Len int
}

func (mc *taosConn) taosConnect(ip, user, pass, db string, port int) (taos unsafe.Pointer, err error) {
	cuser := C.CString(user)
	cpass := C.CString(pass)
	cdb := C.CString(db)
	defer C.free(unsafe.Pointer(cuser))
	defer C.free(unsafe.Pointer(cpass))
	defer C.free(unsafe.Pointer(cdb))
	var taosObj unsafe.Pointer
	if len(ip) == 0 {
		taosObj = C.taos_connect(nil, cuser, cpass, cdb, (C.ushort)(0))
	} else {
		cip := C.CString(ip)
		defer C.free(unsafe.Pointer(cip))
		taosObj = C.taos_connect(cip, cuser, cpass, cdb, (C.ushort)(port))
	}

	if taosObj == nil {
		return nil, &errors.TaosError{
			Code:   errors.TSC_INVALID_CONNECTION,
			ErrStr: "invalid connection",
		}
	}

	return taosObj, nil
}

func (mc *taosConn) taosQuery(sqlstr string) (int, error) {
	//csqlstr := (*UserChar)(unsafe.Pointer(&sqlstr))

	csqlstr := C.CString(sqlstr)
	defer C.free(unsafe.Pointer(csqlstr))
	if mc.result != nil {
		C.taos_free_result(mc.result)
		mc.result = nil
	}
	mc.result = unsafe.Pointer(C.taos_query(mc.taos, csqlstr))
	//mc.result = unsafe.Pointer(C.taos_query_c(mc.taos, csqlstr.Str, C.uint32_t(csqlstr.Len)))
	code := C.taos_errno(mc.result)
	if code != 0 {
		errStr := C.GoString(C.taos_errstr(mc.result))
		mc.taos_error()

		return 0, &errors.TaosError{
			Code:   int32(code) & 0xffff,
			ErrStr: errStr,
		}
	}

	// read result and save into mc struct
	num_fields := int(C.taos_field_count(mc.result))
	if num_fields == 0 {
		// there are no select and show kinds of commands
		mc.affectedRows = int(C.taos_affected_rows(mc.result))
		mc.insertId = 0
	}

	return num_fields, nil
}

func (mc *taosConn) taosFetchBlock() (uint, error) {
	var block C.TAOS_ROW
	mc.block = unsafe.Pointer(&block)
	mc.blockOffset = 0
	mc.blockSize = uint(C.taos_fetch_block(mc.result, (*C.TAOS_ROW)(mc.block)))
	return mc.blockSize, nil
}

func (mc *taosConn) taos_close() {
	C.taos_close(mc.taos)
	mc.taos = nil
}

func (mc *taosConn) taos_error() {
	// free local resouce: allocated memory/metric-meta refcnt
	//var pRes unsafe.Pointer
	C.taos_free_result(mc.result)
	mc.result = nil
}

func (mc *taosConn) free_result() {
	// free result in db.Exec immediately
	C.taos_free_result(mc.result)
	mc.result = nil
}
