package base58

import (
	"bytes"
	"errors"
	"io"
)

/*
Base58 Encoding Implementation in Pure Go

BSD 3-Clause License, Copyright (c) 2025, cyclone
https://github.com/cyclone-github/base58/blob/main/LICENSE

This base58 encoder/decoder works directly on byte slices using custom repeated
division routines. Compared with the btcutil (big.Int) approach, this method
reduces allocations and avoids overhead associated with arbitrary‐precision
arithmetic. This makes it more efficient for typical input sizes like crypto addresses,
public key hashes, short strings, etc.

Written in Pure Go by Cyclone
https://github.com/cyclone-github/base58

API based off the offical Go encoding/base64 package
https://pkg.go.dev/encoding/base64

changelog:
0.1.0; 2025-04-08
	initial github release
*/

// standard base58 alphabet used in Bitcoin
const BitcoinAlphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

// radix-58 encoding/decoding scheme
type Encoding struct {
	encode  [58]byte
	reverse [256]int8
}

// encode with 58-char alphabet
func NewEncoding(alphabet string) *Encoding {
	if len(alphabet) != 58 {
		panic("base58 alphabet must be 58 characters")
	}
	enc := new(Encoding)
	for i := 0; i < 58; i++ {
		enc.encode[i] = alphabet[i]
	}
	for i := 0; i < 256; i++ {
		enc.reverse[i] = -1
	}
	for i := 0; i < 58; i++ {
		enc.reverse[alphabet[i]] = int8(i)
	}
	return enc
}

// std bitcoin base58 encoding
var StdEncoding = NewEncoding(BitcoinAlphabet)

// encode src to base58 and write to dst
func (enc *Encoding) Encode(dst, src []byte) int {
	s := enc.EncodeToBytes(src)
	copy(dst, s)
	return len(s)
}

// return base58 encoding as bytes
func (enc *Encoding) EncodeToBytes(src []byte) []byte {
	zeros := 0
	for zeros < len(src) && src[zeros] == 0 {
		zeros++
	}
	input := make([]byte, len(src))
	copy(input, src)
	var b58 []byte
	for len(input) > 0 && !allZero(input) {
		var remainder int
		input, remainder = divmod(input, 58)
		b58 = append(b58, byte(remainder))
	}
	for i := 0; i < zeros; i++ {
		b58 = append(b58, 0)
	}
	reverseBytes(b58)
	for i, v := range b58 {
		b58[i] = enc.encode[v]
	}
	return b58
}

// return base58 encoding as string
func (enc *Encoding) EncodeToString(src []byte) string {
	return string(enc.EncodeToBytes(src))
}

// decode src from base58 and write to dst
func (enc *Encoding) Decode(dst, src []byte) (int, error) {
	res, err := enc.DecodeToBytes(src)
	if err != nil {
		return 0, err
	}
	copy(dst, res)
	return len(res), nil
}

// decode src from base58 to bytes
func (enc *Encoding) DecodeToBytes(src []byte) ([]byte, error) {
	digits := make([]byte, len(src))
	for i, c := range src {
		val := enc.reverse[c]
		if val == -1 {
			return nil, errors.New("base58: invalid character")
		}
		digits[i] = byte(val)
	}
	zeros := 0
	for zeros < len(digits) && digits[zeros] == 0 {
		zeros++
	}
	var b256 []byte
	for len(digits) > 0 && !allZero(digits) {
		var remainder int
		digits, remainder = divmod58(digits, 256)
		b256 = append(b256, byte(remainder))
	}
	for i := 0; i < zeros; i++ {
		b256 = append(b256, 0)
	}
	reverseBytes(b256)
	return b256, nil
}

// decode s from base58
func (enc *Encoding) DecodeString(s string) ([]byte, error) {
	return enc.DecodeToBytes([]byte(s))
}

// check if all bytes are zero
func allZero(b []byte) bool {
	for _, v := range b {
		if v != 0 {
			return false
		}
	}
	return true
}

// reverse bytes in place
func reverseBytes(b []byte) {
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}
}

// divmod for base256 number
func divmod(number []byte, divisor int) ([]byte, int) {
	var remainder int
	quotient := make([]byte, 0, len(number))
	for _, digit := range number {
		accumulator := int(digit) + remainder*256
		quotientDigit := accumulator / divisor
		remainder = accumulator % divisor
		if len(quotient) > 0 || quotientDigit != 0 {
			quotient = append(quotient, byte(quotientDigit))
		}
	}
	return quotient, remainder
}

// divmod for base58 number
func divmod58(number []byte, divisor int) ([]byte, int) {
	var remainder int
	quotient := make([]byte, 0, len(number))
	for _, digit := range number {
		accumulator := int(digit) + remainder*58
		quotientDigit := accumulator / divisor
		remainder = accumulator % divisor
		if len(quotient) > 0 || quotientDigit != 0 {
			quotient = append(quotient, byte(quotientDigit))
		}
	}
	return quotient, remainder
}

type encoder struct {
	enc *Encoding
	w   io.Writer
	buf bytes.Buffer
}

// buffer data
func (e *encoder) Write(p []byte) (int, error) {
	return e.buf.Write(p)
}

// encode and write buffered data
func (e *encoder) Close() error {
	encoded := e.enc.EncodeToBytes(e.buf.Bytes())
	_, err := e.w.Write(encoded)
	return err
}

// base58 stream encoder
func NewEncoder(enc *Encoding, w io.Writer) io.WriteCloser {
	return &encoder{enc: enc, w: w}
}

type decoder struct {
	enc *Encoding
	r   io.Reader
	buf bytes.Buffer
}

// read decoded data
func (d *decoder) Read(p []byte) (int, error) {
	if d.buf.Len() == 0 {
		_, err := d.buf.ReadFrom(d.r)
		if err != nil && err != io.EOF {
			return 0, err
		}
		decoded, err := d.enc.DecodeToBytes(d.buf.Bytes())
		if err != nil {
			return 0, err
		}
		d.buf.Reset()
		d.buf.Write(decoded)
	}
	return d.buf.Read(p)
}

// base58 stream decoder
func NewDecoder(enc *Encoding, r io.Reader) io.Reader {
	return &decoder{enc: enc, r: r}
}
