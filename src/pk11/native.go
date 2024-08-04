// Copyright lowRISC contributors (OpenTitan project).
// Licensed under the Apache License, Version 2.0, see LICENSE for details.
// SPDX-License-Identifier: Apache-2.0

// package native is a cgo wrapper for reaching into the C API of a PKCS#11
// module when package pkcs11 does not provide a useable interface.
package native

import (
	"unsafe"

	"github.com/miekg/pkcs11"
)

/*
// We don't depend on a full pkcs11.h; instead, we redefine things as we need
// them, because the definitions are set in stone by the standard.

#include <stdlib.h> // For calloc() and free().

typedef unsigned char CK_BBOOL;
typedef unsigned long CK_ULONG;
typedef CK_ULONG CK_RV;
typedef CK_ULONG CK_SESSION_HANDLE;
typedef CK_ULONG CK_OBJECT_HANDLE;
typedef CK_ULONG CK_ATTRIBUTE_TYPE;
typedef CK_ULONG CK_MECHANISM_TYPE;

struct CK_MECHANISM {
  CK_MECHANISM_TYPE mech;
  void* param;
  CK_ULONG len;
};

struct CK_ATTRIBUTE {
  CK_ATTRIBUTE_TYPE attr;
  void* param;
  CK_ULONG len;
};

typedef struct CK_MECHANISM* CK_MECHANISM_PTR;
typedef struct CK_ATTRIBUTE* CK_ATTRIBUTE_PTR;
typedef CK_OBJECT_HANDLE* CK_OBJECT_HANDLE_PTR;

#define kDeriveKeyOffset 63
typedef CK_RV (*CK_C_DeriveKey)(
  CK_SESSION_HANDLE, CK_MECHANISM_PTR, CK_OBJECT_HANDLE,
  CK_ATTRIBUTE_PTR, CK_ULONG, CK_OBJECT_HANDLE_PTR);

struct CK_HKDF_PARAMS {
  CK_BBOOL extract, expand;
  CK_MECHANISM_TYPE hash;
  CK_ULONG salt_type;
  void* salt;
  CK_ULONG salt_len;
  CK_OBJECT_HANDLE salt_key;
  void* info;
  CK_ULONG info_len;
};

struct ctx {
  void* handle;
  void** vtable;
};

CK_RV RawDeriveKey(struct ctx** c, CK_SESSION_HANDLE session,
		CK_MECHANISM_PTR mech, CK_OBJECT_HANDLE basekey,
		CK_ATTRIBUTE_PTR attrs, CK_ULONG attr_count, CK_OBJECT_HANDLE_PTR newkey)
{
  void* fn = (**c).vtable[kDeriveKeyOffset];
  return ((CK_C_DeriveKey)fn)(session, mech, basekey, attrs, attr_count, newkey);
}
*/
import "C"

// RawMech is a non-owning reference to a mechanism whose parameter may be
// allocated on the C heap.
type RawMech struct {
	typ   C.CK_MECHANISM_TYPE
	param unsafe.Pointer
	len   C.CK_ULONG
}

// HKDFParams is a Go representation of CK_HKDF_PARAMS.
type HKDFParams struct {
	Extract, Expand bool
	Hash            uint
	SaltType        uint
	Salt            []byte
	SaltKey         pkcs11.ObjectHandle
	Info            []byte

	raw *C.struct_CK_HKDF_PARAMS
}

// MakeRawMech allocates (and caches) a CK_HKDF_PARAM containing copies of
// data in the Go fields, and returns it wrapped in a RawMech.
//
// It must be freed with Free().
func (p *HKDFParams) MakeRawMech(mech uint) RawMech {
	if p.raw != nil {
		goto done
	}

	// This is calloc() to avoid Go getting confused from witnessing uninitialized
	// memory.
	p.raw = (*C.struct_CK_HKDF_PARAMS)(C.calloc(C.sizeof_struct_CK_HKDF_PARAMS, 1))
	if p.raw == nil {
		panic("calloc() returned nil")
	}

	*p.raw = C.struct_CK_HKDF_PARAMS{
		hash:      C.CK_MECHANISM_TYPE(p.Hash),
		salt_type: C.CK_ULONG(p.SaltType),
		salt_key:  C.CK_OBJECT_HANDLE(p.SaltKey),
	}

	if p.Extract {
		p.raw.extract = 1
	}
	if p.Expand {
		p.raw.expand = 1
	}

	if len(p.Salt) > 0 {
		p.raw.salt = C.CBytes(p.Salt)
		p.raw.salt_len = C.CK_ULONG(len(p.Salt))
	}
	if len(p.Info) > 0 {
		p.raw.info = C.CBytes(p.Info)
		p.raw.info_len = C.CK_ULONG(len(p.Info))
	}

done:
	return RawMech{
		typ:   C.CK_MECHANISM_TYPE(mech),
		param: unsafe.Pointer(p.raw),
		len:   C.CK_ULONG(C.sizeof_struct_CK_HKDF_PARAMS),
	}
}

// Frees memory allocated with MakeRawMech.
func (p *HKDFParams) Free() {
	if p.raw == nil {
		return
	}

	C.free(p.raw.salt)
	C.free(p.raw.info)
	C.free(unsafe.Pointer(p.raw))
	p.raw = nil
}

// copyAttrs creates a copy of attrs in the ABI expected by Cryptoki, living on
// the C heap.
//
// The returned memory will not be gc'd; use freeAttrs() to free it.
func copyAttrs(attrs []*pkcs11.Attribute) C.CK_ATTRIBUTE_PTR {
	ptr := C.calloc(C.size_t(len(attrs)), C.sizeof_struct_CK_ATTRIBUTE)
	if ptr == nil {
		panic("calloc() returned nil")
	}

	for i, attr := range attrs {
		offset := C.sizeof_struct_CK_ATTRIBUTE * C.size_t(i)
		cAttr := (*C.struct_CK_ATTRIBUTE)(unsafe.Pointer(uintptr(ptr) + uintptr(offset)))

		var value unsafe.Pointer
		if len(attr.Value) > 0 {
			value = C.CBytes(attr.Value)
		}
		*cAttr = C.struct_CK_ATTRIBUTE{
			attr:  C.CK_ATTRIBUTE_TYPE(attr.Type),
			param: value,
			len:   C.CK_ULONG(len(attr.Value)),
		}
	}
	return C.CK_ATTRIBUTE_PTR(ptr)
}

// freeAttrs frees memory allocated by copyAttrs.
func freeAttrs(attrs C.CK_ATTRIBUTE_PTR, count C.size_t) {
	ptr := unsafe.Pointer(attrs)
	for i := C.size_t(0); i < count; i++ {
		offset := C.sizeof_struct_CK_ATTRIBUTE * C.size_t(i)
		cAttr := (*C.struct_CK_ATTRIBUTE)(unsafe.Pointer(uintptr(ptr) + uintptr(offset)))
		C.free(cAttr.param)
	}
	C.free(unsafe.Pointer(attrs))
}

// RawDeriveKey performs a raw call to C_DeriveKey using a raw mechanism.
//
// This function is intended to support key derivation with mechanisms that
// package pkcs11 does not support.
func RawDeriveKey(ctx *pkcs11.Ctx, sess pkcs11.SessionHandle, basekey pkcs11.ObjectHandle, mech RawMech, attrs []*pkcs11.Attribute) (pkcs11.ObjectHandle, error) {
	cAttrs := copyAttrs(attrs)
	defer freeAttrs(cAttrs, C.size_t(len(attrs)))

	cMech := C.struct_CK_MECHANISM{
		mech:  mech.typ,
		param: mech.param,
		len:   mech.len,
	}

	var obj C.CK_OBJECT_HANDLE
	rv := C.RawDeriveKey(
		(**C.struct_ctx)(unsafe.Pointer(ctx)),
		C.CK_SESSION_HANDLE(sess),
		(*C.struct_CK_MECHANISM)(unsafe.Pointer(&cMech)),
		C.CK_OBJECT_HANDLE(basekey),
		cAttrs,
		C.CK_ULONG(len(attrs)),
		(*C.CK_OBJECT_HANDLE)(unsafe.Pointer(&obj)))

	if rv != 0 {
		return 0, pkcs11.Error(rv)
	}
	return pkcs11.ObjectHandle(obj), nil
}
