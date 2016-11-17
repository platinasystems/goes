// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

package cli_escapes

const Name = "cli-escapes"

type cmd struct{}

func New() cmd { return cmd{} }

func (cmd) String() string { return Name }
func (cmd) Tag() string    { return "builtin" }
func (cmd) Usage() string  { return "man " + Name }

func (cmd) Man() map[string]string {
	return map[string]string{
		"en_US.UTF-8": `ESCAPES
	A COMMAND may extend to multiple lines by escaping the end of
	line with the backslash character ('\').

		COMMAND ..............\
			..............\
			..............

	Similarly, the space between arguments may be escaped.

		COMMAND with\ one\ argument\ having\ five\ spaces

	Also, the arguments may be single or double quoted.

		COMMAND 'with two arguments' each "having two spaces"
		COMMAND "hello 'beautiful world'"
		COMMAND 'hello \"beautiful world\"'

	But *not*,

		COMMAND 'hello "beautiful world"'

	The command may encode these special characters.

		\a   U+0007 alert or bell
		\b   U+0008 backspace
		\f   U+000C form feed
		\n   U+000A line feed or newline
		\r   U+000D carriage return
		\t   U+0009 horizontal tab
		\v   U+000b vertical tab
		\\   U+005c backslash

	The command may also encode any byte or unicode rune with these.

		\OOO	where OOO are three octal digits
		\xXX	where XX are two hex digits
		\uXXXX
		\UXXXXXXXX

	Finally, the command line may include any unicode rune literal
	supported by Go.
	
		ä 本 日本語`,
	}
}
