// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package vpn

import "errors"

var (
	ErrBusy               = errors.New("busy")
	ErrEmpty              = errors.New("empty")
	ErrIncomplete         = errors.New("incomplete")
	ErrInvalid            = errors.New("invalid")
	ErrInvalidBlockType   = errors.New("invalid block type")
	ErrInvalidKey         = errors.New("invalid key")
	ErrNeedServer         = errors.New("need " + vpnRegistrySyntax)
	ErrNoCertificates     = errors.New("no certificates")
	ErrNoExchange         = errors.New("no registered exchange")
	ErrNoNonce            = errors.New("no cipher nonce")
	ErrNoPrivateKeys      = errors.New("no provate keys")
	ErrNoRegistryURI      = errors.New("no registry URI")
	ErrNoServiceIP        = errors.New("unspecified and unfound IP address")
	ErrNotFound           = errors.New("not found")
	ErrNotPEM             = errors.New("isn't PEM encoded")
	ErrPending            = errors.New("already pending approval")
	ErrOverrun            = errors.New("overrun")
	ErrRange              = errors.New("out of range")
	ErrSubscribed         = errors.New("already subscribed")
	ErrTooLong            = errors.New("duration can't exceed 10 years")
	ErrTooManyExchanges   = errors.New("too many exchanges")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrUnderrun           = errors.New("underrun")
	ErrUnnamedCertificate = errors.New("unnamed certificate")
	ErrUnspecifiedAddress = errors.New("unspecified <address>")
	ErrUnspecifiedVPN     = errors.New("unspecified <vpn>")
	ErrUnspecifiedSub     = errors.New("unspecified <subscriber>")

	FIXME = errors.New("FIXME")
)
