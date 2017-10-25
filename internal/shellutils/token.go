// Copyright © 2017 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package shellutils

// Tokentype defines the type of a token
// tokenLiteral is liternal string text. The string is the text of the word
// being assembled.
// tokenEnvget is a reference to an environment variable.  The string is the
// environment variable name. Any punctuation (dollar, brace) has been stripped
// tokenEnvset is the operator to set an environment variable. The string is
// the assignment operator, i.e. =. This is represented as a token to prevent
// quoted = characters to be interpreted as setting environment variables
type Tokentype int

const (
	TokenLiteral = iota
	TokenEnvget
	TokenEnvset
)

// Token is a type and a string value. During parsing, we convert
// string input into a series of tokens.
type Token struct {
	V string
	T Tokentype
}
