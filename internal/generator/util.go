package generator

// CamelCase camel-cases a protobuf name for use as a Go identifier.
//
// If there is an interior underscore followed by a lower case letter,
// drop the underscore and convert the letter to upper case.
// from https://github.com/protocolbuffers/protobuf-go/blob/master/internal/strs/strings.go
func CamelCase(s string) string {
	// Invariant: if the next letter is lower case, it must be converted
	// to upper case.
	// That is, we process a word at a time, where words are marked by _ or
	// upper case letter. Digits are treated as words.
	var b []byte
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c == '.' && i+1 < len(s) && isASCIILower(s[i+1]):
			// Skip over '.' in ".{{lowercase}}".
		case c == '.':
			b = append(b, '_') // convert '.' to '_'
		case c == '_' && (i == 0 || s[i-1] == '.'):
			// Convert initial '_' to ensure we start with a capital letter.
			// Do the same for '_' after '.' to match historic behavior.
			b = append(b, 'X') // convert '_' to 'X'
		case c == '_' && i+1 < len(s) && isASCIILower(s[i+1]):
			// Skip over '_' in "_{{lowercase}}".
		case isASCIIDigit(c):
			b = append(b, c)
		default:
			// Assume we have a letter now - if not, it's a bogus identifier.
			// The next word is a sequence of characters that must start upper case.
			if isASCIILower(c) {
				c -= 'a' - 'A' // convert lowercase to uppercase
			}
			b = append(b, c)

			// Accept lower case sequence that follows.
			for ; i+1 < len(s) && isASCIILower(s[i+1]); i++ {
				b = append(b, s[i+1])
			}
		}
	}
	return string(b)
}

func isASCIILower(c byte) bool {
	return 'a' <= c && c <= 'z'
}
func isASCIIUpper(c byte) bool {
	return 'A' <= c && c <= 'Z'
}
func isASCIIDigit(c byte) bool {
	return '0' <= c && c <= '9'
}

func keySize(fieldNumber int, wire Wire) int {
	x := keyValue(fieldNumber, wire)
	size := 0
	for size = 0; x > 127; size++ {
		x >>= 7
	}
	size++
	return size
}

func keyValue(fieldNumber int, wire Wire) uint32 {
	return uint32(fieldNumber)<<3 | uint32(wire)
}

func wireString(t Type) string {
	switch t := t.(type) {
	case ScalarValueType:
		return t.wire
	case *MessageType:
		return "bytes"
	case *EnumType:
		return "varint"
	}
	panic("unreachable")
}

func (g *Generator) proto2() bool { return g.version == 2 }
func (g *Generator) proto3() bool { return g.version == 3 }

func (g *Generator) importGoPackage(path, alias string) {
	g.goImport[GoPackage{path, alias}] = true
}

func conv(x string, src, dst Type) string {
	if dst != src {
		x = dst.GoType() + "(" + x + ")"
	}
	return x
}
