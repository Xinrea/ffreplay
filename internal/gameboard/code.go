// Package gameboard reads and writes the FFXIV [stgy:a...] clipboard format.
// Format reference: mogworks/xivplan, cn, src/lib/strategy (see asset/strategy/NOTICE.md).
package gameboard

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"hash/crc32"
	"io"
	"strings"
)

const (
	Prefix   = "[stgy:a"
	maxCode  = 10000
	alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"
	cipher   = "fReAFBudk63KL+Y-zT5DnHhQU9GZIjNr1maOpoMXiJlg8Cxcv0sy2w7qStEV4PbW"
)

var (
	le      = binary.LittleEndian
	invalid = errors.New("游戏分享码损坏或格式不受支持")
)

func Decode(code string) (Board, error) {
	code = strings.TrimSpace(code)
	if len(code) > maxCode || len(code) < 12 || !strings.HasPrefix(code, Prefix) ||
		!strings.HasSuffix(code, "]") {
		return Board{}, invalid
	}

	body := code[len(Prefix) : len(code)-1]

	raw, err := decodeBody(body)
	if err != nil || len(raw) < 7 || crc32.ChecksumIEEE(raw[4:]) != le.Uint32(raw) {
		return Board{}, invalid
	}

	return inflate(raw)
}

func decodeBody(body string) ([]byte, error) {
	seed := strings.IndexByte(cipher, body[0])
	if seed < 0 {
		return nil, invalid
	}

	buf := make([]byte, len(body)-1)
	for i := range buf {
		v := strings.IndexByte(cipher, body[i+1])
		if v < 0 {
			return nil, invalid
		}

		buf[i] = alphabet[(v-seed-i)&63]
	}

	return base64.RawURLEncoding.DecodeString(string(buf))
}

func inflate(raw []byte) (Board, error) {
	size := int(le.Uint16(raw[4:]))

	z, err := zlib.NewReader(bytes.NewReader(raw[6:]))
	if err != nil {
		return Board{}, invalid
	}

	defer z.Close()

	payload, err := io.ReadAll(io.LimitReader(z, int64(size)+1))
	if err != nil || len(payload) != size {
		return Board{}, invalid
	}

	return unpack(payload)
}

func Encode(board Board) (string, error) {
	payload, err := pack(board)
	if err != nil {
		return "", err
	}

	var raw bytes.Buffer

	raw.Write(make([]byte, 6))

	w := zlib.NewWriter(&raw)

	_, err = w.Write(payload)
	if err != nil {
		return "", err
	}

	err = w.Close()
	if err != nil {
		return "", err
	}

	p := raw.Bytes()
	le.PutUint16(p[4:], uint16(len(payload)))
	le.PutUint32(p, crc32.ChecksumIEEE(p[4:]))
	b64 := base64.RawURLEncoding.EncodeToString(p)
	seed := int(p[0] & 63)

	var out strings.Builder

	out.WriteString(Prefix)
	out.WriteByte(cipher[seed])

	for i := range len(b64) {
		out.WriteByte(cipher[(strings.IndexByte(alphabet, b64[i])+i+seed)&63])
	}

	out.WriteByte(']')

	if out.Len() > maxCode {
		return "", errors.New("分享码超过游戏长度限制")
	}

	return out.String(), nil
}
