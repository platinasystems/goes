// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netlink

import (
	_ "embed"
	"strings"
	"sync"

	"github.com/platinasystems/goes/v2/pkg/integer"
)

//go:embed nlattr_type.txt
var NlAttrTypeText string

var NlAttrTypeLines = sync.OnceValue(func() []string {
	return strings.Split(NlAttrTypeText, "\n")
})

func NlAttrTypeName[T integer.Int | integer.Uint](i T) string {
	return integer.Name(i, NlAttrTypeLines())
}

//go:embed nlmsgerr_attr.txt
var NlMsgErrAttrText string

var NlMsgErrAttrLines = sync.OnceValue(func() []string {
	return strings.Split(NlMsgErrAttrText, "\n")
})

func NlMsgErrAttrName[T integer.Int | integer.Uint](i T) string {
	return integer.Name(i, NlMsgErrAttrLines())
}

//go:embed nlpolicy_type_attr.txt
var NlPolicyTypeAttrText string

var NlPolicyTypeAttrLines = sync.OnceValue(func() []string {
	return strings.Split(NlPolicyTypeAttrText, "\n")
})

func NlPolicyTypeAttrName[T integer.Int | integer.Uint](i T) string {
	return integer.Name(i, NlPolicyTypeAttrLines())
}
