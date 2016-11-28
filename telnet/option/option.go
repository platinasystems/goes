// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package option

const (
	BINARY byte = iota // RFC 856

	ECHO      // RFC 857
	RECONNECT //
	SGA       // RFC 858
	NAMS      //
	STATUS    // RFC 859
	TIMING    // RFC 860
	RCTE      // RFC 563, 726
	WIDTH     //
	SIZE      //
	NAOCRD    // RFC 652
	NAOHTS    // RFC 653
	NAOHTD    // RFC 654
	NAOFFD    // RFC 655
	NAOVTS    // RFC 656
	NAOVTD    // RFC 657
	NAOLFD    // RFC 658
	EXTASC    // RFC 698
	LOGOUT    // RFC 727
	BM        // RFC 735
	DET       // RFC 732,1043
	SUPDUP    // RFC 734, 736
	SUPDUPOUT // RFC 749
	SNDLOC    // RFC 779
	TYPE      // RFC 1091
	EOR       // RFC 885
	TUID      // RFC 927
	OUTMRK    // RFC 933
	TTYLOC    // RFC 946
	T3270     // RFC 1041
	Xdot3pad  // RFC 1053
	NAWS      // RFC 1073
	SPEED     // RFC 1079
	TFC       // RFC 1372
	LINEMODE  // RFC 1184
	XDISPlOC  // RFC 1096
	ENV       // RFC 1408
	AUTH      // RFC 1416, 2941, 2942, 2943,2951
	ENCRYPT   // RFC 2946
	NEWENV    // RFC 1572
	TN3270E   // RFC 2355
	XAUTH     //
	CHARSET   // RFC 2066
	RSP       //
	CPC       // RFC 2217
	NOECHO    //
	TLS       //
	KERMIT    // RFC 2840
	URL       //
	FORWARDX  //
)

const (
	PRAGMA_LOGON byte = 138 + iota
	SSPI_LOGON
	PRAGMA_HEARTBEAT
)

const EXTOPT byte = 255 // RFC861
