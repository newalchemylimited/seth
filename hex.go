package seth

import (
	"encoding/hex"
	"fmt"
)

// hexlen gives the number of bytes of data
// that correspond to a hex string of the
// given data
func hexlen(b []byte) int {
	if len(b) < 2 {
		return 0
	}
	return (len(b[2:]) + 1) / 2
}

// hexprefix checks if b starts with "0x"
func hexprefix(s []byte) bool {
	return len(s) >= 2 && s[0] == '0' && s[1] == 'x'
}

// hexparse parses hex strings, like "0xaf1"
func hexparse(s []byte) ([]byte, error) {
	dst := make([]byte, hexlen(s))
	if err := hexdecode(dst, s); err != nil {
		return nil, err
	}
	return dst, nil
}

func fromhex(c byte) (byte, bool) {
	switch {
	case '0' <= c && c <= '9':
		return c - '0', true
	case 'a' <= c && c <= 'f':
		return c - 'a' + 10, true
	case 'A' <= c && c <= 'F':
		return c - 'A' + 10, true
	}
	return 0, false
}

// hexparse decodes hex strings, like "0xaf1"
func hexdecode(dst, src []byte) error {
	if !hexprefix(src) {
		return fmt.Errorf("bad hex string %q", src)
	}
	if len(dst) != hexlen(src) {
		return fmt.Errorf("size mismatch: %d != %d", len(dst), hexlen(src))
	}

	// If the input is an odd-sized hex string,
	// make sure that it implicitly has a 0 nibble
	// at the beginning rather than at the end.
	addend := uint(len(src) & 1)
	for i, b := range src[2:] {
		a, ok := fromhex(b)
		if !ok {
			return hex.InvalidByteError(b)
		}
		dst[(uint(i)+addend)>>1] |= a << (4 * (^(uint(i) ^ addend) & 1))
	}

	return nil
}

// hexstring returns a hex string of the given data
func hexstring(b []byte, trunc bool) []byte {
	buf := make([]byte, 2+2*len(b))
	hex.Encode(buf[2:], b)
	if trunc && len(buf) > 2 && buf[2] == '0' {
		buf = buf[1:]
	}
	copy(buf, "0x")
	return buf
}
