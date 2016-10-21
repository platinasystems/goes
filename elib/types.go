// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package elib

//go:generate gentemplate -d Package=elib -id Byte  -d VecType=ByteVec -d Type=byte vec.tmpl

//go:generate gentemplate -d Package=elib -id String -d VecType=StringVec -d Type=string vec.tmpl
//go:generate gentemplate -d Package=elib -id String -d PoolType=StringPool -d Type=string -d Data=Strings pool.tmpl

//go:generate gentemplate -d Package=elib -id Int64 -d VecType=Int64Vec -d Type=int64 vec.tmpl
//go:generate gentemplate -d Package=elib -id Int32 -d VecType=Int32Vec -d Type=int32 vec.tmpl
//go:generate gentemplate -d Package=elib -id Int16 -d VecType=Int16Vec -d Type=int16 vec.tmpl
//go:generate gentemplate -d Package=elib -id Int8  -d VecType=Int8Vec -d Type=int8  vec.tmpl

//go:generate gentemplate -d Package=elib -id Uint64 -d VecType=Uint64Vec -d Type=uint64 vec.tmpl
//go:generate gentemplate -d Package=elib -id Uint32 -d VecType=Uint32Vec -d Type=uint32 vec.tmpl
//go:generate gentemplate -d Package=elib -id Uint16 -d VecType=Uint16Vec -d Type=uint16 vec.tmpl
//go:generate gentemplate -d Package=elib -id Uint8  -d VecType=Uint8Vec -d Type=uint8  vec.tmpl
