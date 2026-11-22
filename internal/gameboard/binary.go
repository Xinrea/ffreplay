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

type section struct{ tag, kind uint16 }

var sections = []section{{4, 1}, {5, 3}, {6, 1}, {7, 0}, {8, 2}, {10, 1}, {11, 1}, {12, 1}}

// The game writes one section per attribute, not one record per object.
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

	p := bytes.TrimRight(r.take(n), "\x00")
	if !utf8.Valid(p) || bytes.IndexByte(p, 0) >= 0 {
		r.bad = true
	}

	return string(p)
}

func (r *reader) objects() ([]Object, error) {
	var objects []Object

	for len(r.p) >= 4 && le.Uint16(r.p) == 2 {
		r.word()

		o := Object{Type: r.word()}
		if o.Type == 100 {
			if r.word() != 3 {
				return nil, invalid
			}

			o.Text = r.str()
		}

		if len(objects) >= MaxObjects {
			return nil, errors.New("游戏战术板最多 50 个对象")
		}

		objects = append(objects, o)
	}

	return objects, nil
}

func (r *reader) readValue(tag uint16, o *Object) {
	switch tag {
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
		o.Params[tag-10] = r.word()
	}
}

func (r *reader) readSections(objects []Object) bool {
	if len(objects) == 0 && len(r.p) == 8 {
		return true
	}

	for _, s := range sections {
		if r.word() != s.tag || r.word() != s.kind || int(r.word()) != len(objects) {
			return false
		}

		for i := range objects {
			r.readValue(s.tag, &objects[i])
		}

		if s.tag == 7 && len(objects)%2 != 0 {
			r.take(1)
		}
	}

	return !r.bad
}

func validHeader(data []byte) bool {
	return len(data) >= 36 && le.Uint32(data) == 2 &&
		int(le.Uint32(data[4:])) == len(data)-16 && int(le.Uint32(data[18:])) == len(data)-28
}

func unpack(data []byte) (Board, error) {
	if !validHeader(data) {
		return Board{}, invalid
	}

	r := reader{p: data[24:]}
	if r.word() != 1 {
		return Board{}, invalid
	}

	name := r.str()

	objects, err := r.objects()
	if err != nil {
		return Board{}, err
	}

	if !r.readSections(objects) || !r.footer() {
		return Board{}, invalid
	}

	bg := r.word()
	if r.bad || len(r.p) != 0 || bg < 1 || bg > 7 {
		return Board{}, invalid
	}

	return Board{Name: name, Objects: objects, Background: bg}, nil
}

func (r *reader) footer() bool { return r.word() == 3 && r.word() == 1 && r.word() == 1 }

func validText(s string) bool { return utf8.ValidString(s) && !strings.ContainsRune(s, 0) }
func (b Board) validate() error {
	if len(b.Objects) > MaxObjects {
		return errors.New("游戏战术板最多 50 个对象（含连线）")
	}

	if len(b.Name) > 20 || !validText(b.Name) {
		return errors.New("游戏板名最多 20 个 UTF-8 字节，请缩短步骤名称")
	}

	if b.Background < 1 || b.Background > 7 {
		return invalid
	}

	return b.validateTexts()
}

func (b Board) validateTexts() error {
	count := 0

	for _, o := range b.Objects {
		if o.Type != 100 {
			continue
		}

		count++
		if count > 8 || utf8.RuneCountInString(o.Text) > 30 || !validText(o.Text) {
			return errors.New("游戏最多 8 段文字，每段最多 30 字")
		}
	}

	return nil
}

type writer []byte

func (w *writer) word(v uint16) { *w = le.AppendUint16(*w, v) }
func (w *writer) str(v string) {
	n := (len(v) + 4) &^ 3
	w.word(uint16(n))
	*w = append(*w, v...)
	*w = append(*w, make([]byte, n-len(v))...)
}

func (w *writer) value(tag uint16, o Object) {
	switch tag {
	case 4:
		w.word(o.Flags)
	case 5:
		w.word(uint16(o.X))
		w.word(uint16(o.Y))
	case 6:
		w.word(uint16(o.Angle))
	case 7:
		*w = append(*w, o.Scale)
	case 8:
		*w = append(*w, o.RGBA[:]...)
	default:
		w.word(o.Params[tag-10])
	}
}

func (w *writer) sections(objects []Object) {
	if len(objects) == 0 {
		return
	}

	for _, s := range sections {
		w.word(s.tag)
		w.word(s.kind)
		w.word(uint16(len(objects)))

		for _, o := range objects {
			w.value(s.tag, o)
		}

		if s.tag == 7 && len(objects)%2 != 0 {
			*w = append(*w, 0)
		}
	}
}

func pack(b Board) ([]byte, error) {
	err := b.validate()
	if err != nil {
		return nil, err
	}

	w := make(writer, 24)
	le.PutUint32(w, 2)
	w.word(1)
	w.str(b.Name)

	for _, o := range b.Objects {
		w.word(2)
		w.word(o.Type)

		if o.Type == 100 {
			w.word(3)
			w.str(o.Text)
		}
	}

	w.sections(b.Objects)
	w.word(3)
	w.word(1)
	w.word(1)
	w.word(b.Background)
	le.PutUint32(w[4:], uint32(len(w)-16))
	le.PutUint32(w[18:], uint32(len(w)-28))

	return w, nil
}
