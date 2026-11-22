package gameboard

import (
	"bytes"
	"errors"
	"strings"
	"unicode/utf8"
)

const MaxObjects = 50

type Board struct {
	Name       string
	Background uint16
	// The wire stores frontmost first, opposite to the planner's paint order.
	Objects []Object
}

type Object struct {
	Type, Flags uint16
	Text        string
	X, Y, Angle int16
	Scale       byte
	RGBA        [4]byte // last byte is transparency, 0 = opaque
	Params      [3]uint16
}

// A bounded cursor: no offset arithmetic or slice allocation uses an
// unchecked length from the clipboard.
type reader struct {
	p   []byte
	bad bool
}

func (r *reader) take(n int) []byte {
	if n > len(r.p) {
		r.bad = true
		return nil
	}
	p := r.p[:n]
	r.p = r.p[n:]
	return p
}
func (r *reader) word() uint16 {
	p := r.take(2)
	if p == nil {
		return 0
	}
	return le.Uint16(p)
}
func (r *reader) str() string {
	n := int(r.word())
	p := r.take(n)
	p = bytes.TrimRight(p, "\x00")
	if !utf8.Valid(p) || bytes.IndexByte(p, 0) >= 0 {
		r.bad = true
	}
	return string(p)
}

var sections = []struct{ tag, kind uint16 }{{4, 1}, {5, 3}, {6, 1}, {7, 0}, {8, 2}, {10, 1}, {11, 1}, {12, 1}}

func unpack(data []byte) (Board, error) {
	var b Board
	if len(data) < 36 || le.Uint32(data) != 2 || int(le.Uint32(data[4:])) != len(data)-16 || int(le.Uint32(data[18:])) != len(data)-28 {
		return b, invalid
	}
	r := reader{p: data[24:]}
	if r.word() != 1 {
		return b, invalid
	}
	b.Name = r.str()
	for len(r.p) >= 4 && le.Uint16(r.p) == 2 {
		r.word()
		o := Object{Type: r.word()}
		if o.Type == 100 {
			if r.word() != 3 {
				return b, invalid
			}
			o.Text = r.str()
		}
		if len(b.Objects) >= MaxObjects {
			return b, errors.New("游戏战术板最多 50 个对象")
		}
		b.Objects = append(b.Objects, o)
	}
	n := len(b.Objects)
	if n != 0 || len(r.p) > 8 {
		for _, s := range sections {
			if r.word() != s.tag || r.word() != s.kind || int(r.word()) != n {
				return b, invalid
			}
			for i := range b.Objects {
				o := &b.Objects[i]
				switch s.tag {
				case 4:
					o.Flags = r.word()
				case 5:
					o.X, o.Y = int16(r.word()), int16(r.word())
				case 6:
					o.Angle = int16(r.word())
				case 7:
					p := r.take(1)
					if p != nil {
						o.Scale = p[0]
					}
				case 8:
					copy(o.RGBA[:], r.take(4))
				default:
					o.Params[s.tag-10] = r.word()
				}
			}
			if s.tag == 7 && n%2 != 0 {
				r.take(1)
			}
		}
	}
	if r.word() != 3 || r.word() != 1 || r.word() != 1 {
		return b, invalid
	}
	b.Background = r.word()
	if r.bad || len(r.p) != 0 || b.Background < 1 || b.Background > 7 {
		return Board{}, invalid
	}
	return b, nil
}

func pack(b Board) ([]byte, error) {
	if len(b.Objects) > MaxObjects {
		return nil, errors.New("游戏战术板最多 50 个对象（含连线）")
	}
	if len(b.Name) > 20 || strings.ContainsRune(b.Name, 0) || !utf8.ValidString(b.Name) {
		return nil, errors.New("游戏板名最多 20 个 UTF-8 字节，请缩短步骤名称")
	}
	if b.Background < 1 || b.Background > 7 {
		return nil, invalid
	}
	p := make([]byte, 24)
	le.PutUint32(p, 2)
	word := func(v uint16) { p = le.AppendUint16(p, v) }
	str := func(v string) {
		size := (len(v) + 4) &^ 3
		word(uint16(size))
		p = append(p, v...)
		p = append(p, make([]byte, size-len(v))...)
	}
	word(1)
	str(b.Name)
	textCount := 0
	for _, o := range b.Objects {
		word(2)
		word(o.Type)
		if o.Type == 100 {
			textCount++
			if textCount > 8 || utf8.RuneCountInString(o.Text) > 30 || strings.ContainsRune(o.Text, 0) || !utf8.ValidString(o.Text) {
				return nil, errors.New("游戏最多 8 段文字，每段最多 30 字")
			}
			word(3)
			str(o.Text)
		}
	}
	for _, s := range sections {
		if len(b.Objects) == 0 {
			break
		}
		word(s.tag)
		word(s.kind)
		word(uint16(len(b.Objects)))
		for _, o := range b.Objects {
			switch s.tag {
			case 4:
				word(o.Flags)
			case 5:
				word(uint16(o.X))
				word(uint16(o.Y))
			case 6:
				word(uint16(o.Angle))
			case 7:
				p = append(p, o.Scale)
			case 8:
				p = append(p, o.RGBA[:]...)
			default:
				word(o.Params[s.tag-10])
			}
		}
		if s.tag == 7 && len(b.Objects)%2 != 0 {
			p = append(p, 0)
		}
	}
	word(3)
	word(1)
	word(1)
	word(b.Background)
	le.PutUint32(p[4:], uint32(len(p)-16))
	le.PutUint32(p[18:], uint32(len(p)-28))
	return p, nil
}
