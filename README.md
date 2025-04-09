[![Readme Card](https://github-readme-stats.vercel.app/api/pin/?username=cyclone-github&repo=base58&theme=gruvbox)](https://github.com/cyclone-github/base58/)

[![Go Report Card](https://goreportcard.com/badge/github.com/cyclone-github/base58)](https://goreportcard.com/report/github.com/cyclone-github/base58)
[![GitHub issues](https://img.shields.io/github/issues/cyclone-github/base58.svg)](https://github.com/cyclone-github/base58/issues)
[![License](https://img.shields.io/github/license/cyclone-github/base58.svg)](LICENSE)
[![GitHub release](https://img.shields.io/github/release/cyclone-github/base58.svg)](https://github.com/cyclone-github/base58/releases) [![Go Reference](https://pkg.go.dev/badge/github.com/cyclone-github/base58.svg)](https://pkg.go.dev/github.com/cyclone-github/base58)

---

# Base58

Base58 is a pure Go implementation of the Bitcoin Base58 encoding and decoding scheme, based on and following a similar API as the official Go stdlib [encoding/base64](https://pkg.go.dev/encoding/base64) package.

## Overview

Unlike other implementations (such as the btcutil version that uses big.Int for arithmetic), this Base58 package uses custom repeated division routines that work on raw byte slices. This approach offers:

- **Speed & Memory Efficiency:** Fewer allocations and less overhead for typical input sizes (e.g. crypto addresses, public key hashes, short strings).
- **Familiar API:** Modeled on Go’s standard `encoding/base64` package for ease of use and consistency.
- **Stream Support:** Provides stream encoder and decoder wrappers via `NewEncoder` and `NewDecoder`.

## Features

### Available Functions

#### Initialization
- **NewEncoding(alphabet string) Encoding**  
  Returns a new Base58 encoding scheme using the specified 58-character alphabet.

- **StdEncoding**  
  A pre-initialized `Encoding` using the standard Bitcoin alphabet:  
  `"123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"`

#### Encoding
- **(enc Encoding) Encode(dst, src []byte) int**  
  Encodes `src` into Base58, writes the result to `dst`, and returns the number of bytes written.

- **(enc Encoding) EncodeToBytes(src []byte) []byte**  
  Returns the Base58 encoding of `src` as a byte slice.

- **(enc Encoding) EncodeToString(src []byte) string**  
  Returns the Base58 encoding of `src` as a string.

#### Decoding
- **(enc Encoding) Decode(dst, src []byte) (int, error)**  
  Decodes Base58-encoded `src` into `dst` and returns the number of decoded bytes along with an error if any.

- **(enc Encoding) DecodeToBytes(src []byte) ([]byte, error)**  
  Returns a byte slice containing the decoded data from Base58-encoded `src`.

- **(enc Encoding) DecodeString(s string) ([]byte, error)**  
  Decodes the Base58 string `s` and returns the corresponding byte slice.

#### Stream Functions
- **NewEncoder(enc Encoding, w io.Writer) io.WriteCloser**  
  Returns a new stream encoder that writes Base58-encoded data to `w`. The data is buffered and encoded when `Close()` is called.

- **NewDecoder(enc Encoding, r io.Reader) io.Reader**  
  Returns a new stream decoder that reads Base58-encoded data from `r` and provides the decoded output.

## Usage

### One-Shot Encoding & Decoding

```go
package main

import (
	"fmt"
	"log"

	"github.com/cyclone-github/base58"
)

func main() {
	data := []byte("hello world")
	encoded := base58.StdEncoding.EncodeToString(data)
	fmt.Println("Encoded:", encoded)

	decoded, err := base58.StdEncoding.DecodeString(encoded)
	if err != nil {
		log.Fatalf("Decode error: %v", err)
	}
	fmt.Println("Decoded:", string(decoded))
}
```

### Stream-Based Encoding & Decoding

```go
package main

import (
	"bytes"
	"fmt"
	"io"
	"log"

	"github.com/cyclone-github/base58"
)

func main() {
	// Example of streaming encoding:
	var encodedBuffer bytes.Buffer
	encoder := base58.NewEncoder(base58.StdEncoding, &encodedBuffer)
	encoder.Write([]byte("streaming data"))
	encoder.Close()
	fmt.Println("Stream Encoded:", encodedBuffer.String())

	// Example of streaming decoding:
	decoder := base58.NewDecoder(base58.StdEncoding, &encodedBuffer)
	decodedBytes, err := io.ReadAll(decoder)
	if err != nil {
		log.Fatalf("Stream decode error: %v", err)
	}
	fmt.Println("Stream Decoded:", string(decodedBytes))
}
```

## Comparison to Go's stdlib "encoding/base64"

This Base58 package closely follows the design of Go’s `encoding/base64` package:

- **API Parity:**  
  Both packages use an `Encoding` type with methods like `EncodeToString` and `DecodeString`, making the Base58 API familiar to Go developers.

- **Stream Processing:**  
  Just like the base64 package, this package provides stream encoder and decoder wrappers through `NewEncoder` and `NewDecoder`.

- **Efficiency:**  
  While the standard library’s base64 handles 6-bit groups for Base64 conversion, this Base58 package uses custom repeated division routines on raw byte slices. Compared to other Base58 implementations that rely on `math/big`, this approach avoids the overhead of arbitrary-precision arithmetic, thus offering improved performance for typical inputs.

- **Extensibility:**  
  The package can easily be extended to support alternate Base58 alphabets or custom variants, similar to how custom encodings can be created with `encoding/base64`.

## Installation & Import

To install, run:

```bash
go get github.com/cyclone-github/base58
```

Then import the package in your project:

```go
import "github.com/cyclone-github/base58"
```

## License

This project is licensed under the BSD 3-Clause License. See the [LICENSE](LICENSE) file for details.
