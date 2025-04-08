package base58_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/cyclone-github/base58"
)

/*
BSD 3-Clause License, Copyright (c) 2025, cyclone
https://github.com/cyclone-github/base58/blob/main/LICENSE

Written in Pure Go by Cyclone
https://github.com/cyclone-github/base58

API based off the offical Go encoding/base64 package
https://pkg.go.dev/encoding/base64
*/

// testpair defines a decoded input and its expected Base58-encoded result
type testpair struct {
	decoded, encoded string
}

var pairs = []testpair{
	// non-printable byte examples
	{"\x14\xfb\x9c\x03\xd9\x7e", "BT2vGYLD"},
	{"\x14\xfb\x9c\x03\xd9", "3NJeu1J"},
	{"\x14\xfb\x9c\x03", "Y7GPC"},

	// simple examples
	{"", ""},
	{"sure.", "E2XFRyo"},
	{"sure", "3xB2TW"},
	{"sur", "fnKT"},
	{"su", "9nc"},
	{"leasure.", "K8aUZhGUNaR"},
	{"easure.", "4qq4WqChgZ"},
	{"asure.", "qXcNm9C1"},
	{"sure.", "E2XFRyo"},
}

func stdRef(ref string) string { return ref }

// base58 alphabet
const base58Alphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

// nonstandard encoding using the same alphabet
var funnyEncoding = base58.NewEncoding(base58Alphabet)

func funnyRef(ref string) string { return ref }

type encodingTest struct {
	enc  *base58.Encoding
	conv func(string) string
}

var encodingTests = []encodingTest{
	{base58.StdEncoding, stdRef},
	{funnyEncoding, funnyRef},
}

var bigtest = testpair{
	"Twas brillig, and the slithy toves",
	"2ukVBARx4fMCUZXaHR1XvNbb3HgzmGYFEEThDa86tN2q8oU",
}

func testEqual(t *testing.T, msg string, expected, actual any) bool {
	t.Helper()
	if expected != actual {
		t.Errorf(msg, expected, actual)
		return false
	}
	return true
}

func TestEncodeToString(t *testing.T) {
	for _, p := range pairs {
		for _, tt := range encodingTests {
			got := tt.enc.EncodeToString([]byte(p.decoded))
			msg := fmt.Sprintf("EncodeToString(%q): got %%q, want %%q", p.decoded)
			testEqual(t, msg, tt.conv(p.encoded), got)
		}
	}
}

func TestDecodeString(t *testing.T) {
	for _, p := range pairs {
		for _, tt := range encodingTests {
			res, err := tt.enc.DecodeString(tt.conv(p.encoded))
			if err != nil {
				t.Errorf("DecodeString(%q) failed: %v", p.encoded, err)
				continue
			}
			msg := fmt.Sprintf("DecodeString(%q): got %%q, want %%q", p.encoded)
			testEqual(t, msg, p.decoded, string(res))
		}
	}
}

func TestEncoder(t *testing.T) {
	for _, p := range pairs {
		bb := &strings.Builder{}
		encoder := base58.NewEncoder(base58.StdEncoding, bb)
		encoder.Write([]byte(p.decoded))
		encoder.Close()
		msg := fmt.Sprintf("Stream Encode(%q): got %%q, want %%q", p.decoded)
		testEqual(t, msg, p.encoded, bb.String())
	}
}

func TestEncoderBuffering(t *testing.T) {
	input := []byte(bigtest.decoded)
	for bs := 1; bs <= 12; bs++ {
		bb := &strings.Builder{}
		encoder := base58.NewEncoder(base58.StdEncoding, bb)
		for pos := 0; pos < len(input); pos += bs {
			end := pos + bs
			if end > len(input) {
				end = len(input)
			}
			n, err := encoder.Write(input[pos:end])
			msg := fmt.Sprintf("Write(%q): length got %%v, want %%v", input[pos:end])
			testEqual(t, msg, end-pos, n)
			if err != nil {
				t.Errorf("Write(%q) returned error: %v", input[pos:end], err)
			}
		}
		err := encoder.Close()
		testEqual(t, "Close: error got %%v, want nil", nil, err)
		msg := fmt.Sprintf("Buffered stream Encode/%d (%q): got %%q, want %%q", bs, bigtest.decoded)
		testEqual(t, msg, bigtest.encoded, bb.String())
	}
}

func TestDecoder(t *testing.T) {
	for _, p := range pairs {
		res, err := base58.StdEncoding.DecodeString(p.encoded)
		if err != nil {
			t.Fatalf("DecodeString(%q) failed: %v", p.encoded, err)
		}
		msg := fmt.Sprintf("DecodeString(%q): got %%q, want %%q", p.encoded)
		testEqual(t, msg, p.decoded, string(res))

		decoder := base58.NewDecoder(base58.StdEncoding, strings.NewReader(p.encoded))
		decodedBytes, err := io.ReadAll(decoder)
		if err != nil && err != io.EOF {
			t.Fatalf("Stream ReadAll failed for %q: %v", p.encoded, err)
		}
		msg = fmt.Sprintf("Stream decoding of %q: got %%q, want %%q", p.encoded)
		testEqual(t, msg, p.decoded, string(decodedBytes))
	}
}

func TestDecoderBuffering(t *testing.T) {
	for bs := 1; bs <= 12; bs++ {
		decoder := base58.NewDecoder(base58.StdEncoding, strings.NewReader(bigtest.encoded))
		buf := make([]byte, len(bigtest.decoded)+12)
		total := 0
		var n int
		var err error
		for total < len(bigtest.decoded) && err == nil {
			n, err = decoder.Read(buf[total : total+bs])
			total += n
		}
		if err != nil && err != io.EOF {
			t.Errorf("Buffered decoding (buffer size %d) unexpected error: %v", bs, err)
		}
		msg := fmt.Sprintf("Buffered decoding/%d of %q: got %%q, want %%q", bs, bigtest.encoded)
		testEqual(t, msg, bigtest.decoded, string(buf[:total]))
	}
}

func TestBig(t *testing.T) {
	n := 3*1000 + 1
	raw := make([]byte, n)
	const alpha = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	for i := 0; i < n; i++ {
		raw[i] = alpha[i%len(alpha)]
	}
	encodedBuf := new(bytes.Buffer)
	w := base58.NewEncoder(base58.StdEncoding, encodedBuf)
	nn, err := w.Write(raw)
	if nn != n || err != nil {
		t.Fatalf("Encoder.Write(raw) = %d, %v; want %d, nil", nn, err, n)
	}
	err = w.Close()
	if err != nil {
		t.Fatalf("Encoder.Close() = %v; want nil", err)
	}
	decoded, err := io.ReadAll(base58.NewDecoder(base58.StdEncoding, encodedBuf))
	if err != nil {
		t.Fatalf("Stream ReadAll failed: %v", err)
	}
	if !bytes.Equal(raw, decoded) {
		var i int
		for i = 0; i < len(decoded) && i < len(raw); i++ {
			if decoded[i] != raw[i] {
				break
			}
		}
		t.Errorf("Decode(Encode(%d-byte string)) failed at offset %d", n, i)
	}
}

type nextRead struct {
	n   int
	err error
}

type faultInjectReader struct {
	source string
	nextc  <-chan nextRead
}

func (r *faultInjectReader) Read(p []byte) (int, error) {
	nr := <-r.nextc
	if len(p) > nr.n {
		p = p[:nr.n]
	}
	n := copy(p, r.source)
	r.source = r.source[n:]
	return n, nr.err
}

func TestDecoderIssue3577(t *testing.T) {
	next := make(chan nextRead, 10)
	wantErr := errors.New("my error")
	next <- nextRead{5, nil}
	next <- nextRead{10, wantErr}
	next <- nextRead{0, wantErr}
	d := base58.NewDecoder(base58.StdEncoding, &faultInjectReader{
		source: "2ukVBARx4fMCUZXaHR1XvNbb3HgzmGYFEEThDa86tN2q8oU", // from bigtest
		nextc:  next,
	})
	errc := make(chan error, 1)
	go func() {
		_, err := io.ReadAll(d)
		errc <- err
	}()
	select {
	case err := <-errc:
		if err != wantErr {
			t.Errorf("Fault injection: got error %v; want %v", err, wantErr)
		}
	case <-time.After(5 * time.Second):
		t.Errorf("Timeout: Decoder blocked without returning an error")
	}
}

func BenchmarkEncodeToString(b *testing.B) {
	data := make([]byte, 8192)
	b.SetBytes(int64(len(data)))
	for i := 0; i < b.N; i++ {
		base58.StdEncoding.EncodeToString(data)
	}
}

func BenchmarkDecodeString(b *testing.B) {
	sizes := []int{2, 4, 8, 64, 8192}
	benchFunc := func(b *testing.B, benchSize int) {
		data := base58.StdEncoding.EncodeToString(make([]byte, benchSize))
		b.SetBytes(int64(len(data)))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			base58.StdEncoding.DecodeString(data)
		}
	}
	for _, size := range sizes {
		b.Run(fmt.Sprintf("%d", size), func(b *testing.B) {
			benchFunc(b, size)
		})
	}
}
