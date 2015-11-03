package path

import (
	"bytes"
	"fmt"
	"math"
	"math/rand"
	"sort"
	"sync"
	"testing"

	"github.com/thatguystone/cog/check"
)

type Empty struct{}

type Pie struct {
	_ Static `path:"pie"`
	A string
	B uint8
	C string
	D uint8
	E string
	F uint8
	_ Static `path:"end"`
	g string
	h uint8
}

type PieWithInterface struct {
	Pie
}

type Everything struct {
	A int8
	B int16
	C int32
	D int64
	E uint8
	F uint16
	G uint32
	H uint64
	I float32
	J float64
	K complex64
	L complex128
	M bool
}

type EverythingPtr struct {
	A *int8
	B *int16
	C *int32
	D *int64
	E *uint8
	F *uint16
	G *uint32
	H *uint64
	I *float32
	J *float64
	K *complex64
	L *complex128
	M *bool
}

type MinMax struct {
	I int8
	U uint8
}

type NestingI struct {
	A int8
	B int8
}

type NestingU struct {
	A uint8
	B uint8
}

type Nesting struct {
	NestingI
	NestingU
}

func (p PieWithInterface) MarshalPath(b *bytes.Buffer, s Separator) error {
	s.EmitString("pie", b)
	s.EmitDelim(b)
	s.EmitString(p.A, b)
	s.EmitDelim(b)
	s.EmitUint(uint64(p.B), b, 1)
	s.EmitDelim(b)
	s.EmitString(p.C, b)
	s.EmitDelim(b)
	s.EmitUint(uint64(p.D), b, 1)
	s.EmitDelim(b)
	s.EmitString(p.E, b)
	s.EmitDelim(b)
	s.EmitUint(uint64(p.F), b, 1)
	s.EmitDelim(b)
	s.EmitString("end", b)

	return nil
}

var (
	pieWithInterfaceTag0 = []byte("pie")
	pieWithInterfaceTag1 = []byte("end")
)

func (p *PieWithInterface) UnmarshalPath(b *bytes.Buffer, s Separator) error {
	err := s.ExpectTagBytes(pieWithInterfaceTag0, b)

	if err == nil {
		err = s.ExpectDelim(b)
	}

	if err == nil {
		err = s.ExpectString(&p.A, b)
	}

	if err == nil {
		err = s.ExpectDelim(b)
	}

	if err == nil {
		err = s.ExpectUint8(&p.B, b)
	}

	if err == nil {
		err = s.ExpectDelim(b)
	}

	if err == nil {
		err = s.ExpectString(&p.C, b)
	}

	if err == nil {
		err = s.ExpectDelim(b)
	}

	if err == nil {
		err = s.ExpectUint8(&p.D, b)
	}

	if err == nil {
		err = s.ExpectDelim(b)
	}

	if err == nil {
		err = s.ExpectString(&p.E, b)
	}

	if err == nil {
		err = s.ExpectDelim(b)
	}

	if err == nil {
		err = s.ExpectUint8(&p.F, b)
	}

	if err == nil {
		err = s.ExpectDelim(b)
	}

	if err == nil {
		err = s.ExpectTagBytes(pieWithInterfaceTag1, b)
	}

	if err == nil {
		err = s.ExpectDelim(b)
	}

	return err
}

func trampoline(c *check.C, in interface{}, out interface{}, path []byte) {
	p, err := Marshal(in)
	c.MustNotError(err)

	if path != nil {
		c.Equal(path, p)
	}

	err = Unmarshal(p, out)
	c.MustNotError(err)
	c.Equal(in, out)
}

func TestEmpty(t *testing.T) {
	c := check.New(t)

	e := Empty{}
	e2 := Empty{}

	trampoline(c, &e, &e2, []byte("/"))
}

func TestNils(t *testing.T) {
	c := check.New(t)

	_, err := Marshal(nil)
	c.Error(err)

	_, err = Marshal(EverythingPtr{})
	c.Error(err)

	e := Everything{}
	err = Unmarshal(nil, &e)
	c.Error(err)

	err = Unmarshal([]byte("/test/"), nil)
	c.Error(err)
}

func TestNilPtrs(t *testing.T) {
	c := check.New(t)

	tests := []interface{}{
		(*bool)(nil),
		(*int8)(nil),
		(*int16)(nil),
		(*int32)(nil),
		(*int64)(nil),
		(*uint8)(nil),
		(*uint16)(nil),
		(*uint32)(nil),
		(*uint64)(nil),
		(*string)(nil),
		(*[]byte)(nil),
	}

	for _, t := range tests {
		c.Panic(func() {
			Marshal(t)
		})
	}
}

func TestNonNilPtrs(t *testing.T) {
	c := check.New(t)

	tests := []interface{}{
		new(bool),
		new(int8),
		new(int16),
		new(int32),
		new(int64),
		new(uint8),
		new(uint16),
		new(uint32),
		new(uint64),
		new(string),
	}

	for _, t := range tests {
		c.NotPanic(func() {
			_, err := Marshal(t)
			c.NotError(err)
		})
	}
}

func TestNonNils(t *testing.T) {
	c := check.New(t)

	p := EverythingPtr{
		A: new(int8),
		B: new(int16),
		C: new(int32),
		D: new(int64),
		E: new(uint8),
		F: new(uint16),
		G: new(uint32),
		H: new(uint64),
		I: new(float32),
		J: new(float64),
		K: new(complex64),
		L: new(complex128),
		M: new(bool),
	}

	*p.A = 2
	*p.B = 1
	*p.C = 6
	*p.D = 8
	*p.E = 8
	*p.K = 10 + 19i
	*p.M = true

	p2 := EverythingPtr{
		A: new(int8),
		B: new(int16),
		C: new(int32),
		D: new(int64),
		E: new(uint8),
		F: new(uint16),
		G: new(uint32),
		H: new(uint64),
		I: new(float32),
		J: new(float64),
		K: new(complex64),
		L: new(complex128),
		M: new(bool),
	}

	trampoline(c, &p, &p2, nil)
}

func TestMarshalerInterface(t *testing.T) {
	c := check.New(t)

	p := PieWithInterface{
		Pie{
			A: "apple",
			B: 1,
			C: "apple",
			D: 1,
			E: "apple",
			F: 1,
		},
	}
	p2 := PieWithInterface{}

	trampoline(c, &p, &p2, nil)
}

func TestNonStruct(t *testing.T) {
	c := check.New(t)

	i := int32(1)
	i2 := int32(0)

	trampoline(c, &i, &i2, []byte("/\x00\x00\x00\x01/"))
}

func TestBytes(t *testing.T) {
	c := check.New(t)

	i := []byte("test")
	var i2 []byte
	out := []byte("/test/")

	trampoline(c, &i, &i2, out)

	p, err := Marshal(i)
	c.MustNotError(err)
	c.Equal(out, p)
	err = Unmarshal(p, &i2)
	c.MustNotError(err)
	c.Equal(i, i2)
}

func TestNonStructPtr(t *testing.T) {
	c := check.New(t)

	var i *int
	_, err := Marshal(i)
	c.Error(err)
}

func TestPieTrampoline(t *testing.T) {
	c := check.New(t)

	pie := Pie{
		A: "apple",
		B: 1,
		C: "pumpkin",
		D: 2,
		E: "berry",
		F: 3,
		g: "nope",
		h: 123,
	}

	p, err := Marshal(pie)
	c.MustNotError(err)
	c.Equal("/pie/apple/\x01/pumpkin/\x02/berry/\x03/end/", string(p))

	p2 := Pie{}
	err = Unmarshal(p, &p2)
	c.MustNotError(err)
	c.Equal(pie.A, p2.A)
	c.Equal(pie.B, p2.B)
	c.Equal(pie.C, p2.C)
	c.Equal(pie.D, p2.D)
	c.Equal(pie.E, p2.E)
	c.Equal(pie.F, p2.F)
	c.Equal("", p2.g)
	c.Equal(0, p2.h)
}

func TestEverything(t *testing.T) {
	c := check.New(t)

	e := Everything{
		A: math.MinInt8,
		B: math.MinInt16,
		C: math.MinInt32,
		D: math.MinInt64,
		E: math.MaxUint8,
		F: math.MaxUint16,
		G: math.MaxUint32,
		H: math.MaxUint64,
		I: float32(1.2),
		J: float64(1123131.298488474),
		K: 1 + 0i,
		L: 1 + 0i,
		M: true,
	}

	e2 := Everything{}

	trampoline(c, &e, &e2, nil)
}

func TestSort(t *testing.T) {
	c := check.New(t)

	vs := []struct {
		A uint8
		B string
	}{
		{
			A: 2,
			B: "b",
		},
		{
			A: 2,
			B: "a",
		},
		{
			A: 1,
			B: "f",
		},
		{
			A: 1,
			B: "e",
		},
	}

	bs := []string{}

	for _, v := range vs {
		path, err := Marshal(v)
		c.MustNotError(err)

		bs = append(bs, string(path))
	}

	sort.Sort(sort.StringSlice(bs))

	ss := []string{
		"/\x01/e/",
		"/\x01/f/",
		"/\x02/a/",
		"/\x02/b/",
	}
	c.Equal(ss, bs)
}

func TestFuzz(t *testing.T) {
	c := check.New(t)

	wg := sync.WaitGroup{}
	for i := 0; i < 2048; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			e := Everything{
				A: int8(rand.Int63()),
				B: int16(rand.Int63()),
				C: int32(rand.Int63()),
				D: rand.Int63(),
				E: uint8(rand.Uint32()),
				F: uint16(rand.Uint32()),
				G: rand.Uint32(),
				H: uint64(rand.Uint32())<<32 | uint64(rand.Uint32()),
				I: rand.Float32(),
				J: rand.Float64(),
				K: complex(rand.Float32(), rand.Float32()),
				L: complex(rand.Float64(), rand.Float64()),
				M: rand.Int()%2 == 0,
			}

			e2 := Everything{}

			trampoline(c, &e, &e2, nil)
		}()
	}

	wg.Wait()
}

func TestNesting(t *testing.T) {
	c := check.New(t)

	n := Nesting{
		NestingI: NestingI{
			A: 1,
			B: 2,
		},
		NestingU: NestingU{
			A: 3,
			B: 4,
		},
	}

	n2 := Nesting{}

	trampoline(c, &n, &n2, []byte("/\x01/\x02/\x03/\x04/"))
}

func TestMinMax(t *testing.T) {
	c := check.New(t)

	i8 := int8(math.MinInt8)
	u8 := uint8(0)

	wg := sync.WaitGroup{}
	for i := 0; i <= math.MaxUint8; i++ {
		wg.Add(1)
		go func(i8 int8, u8 uint8) {
			wg.Done()
			mm := MinMax{
				I: i8,
				U: u8,
			}

			mm2 := MinMax{}

			trampoline(c, &mm, &mm2, nil)
		}(i8, u8)

		i8++
		u8++
	}

	wg.Wait()
}

func TestArray(t *testing.T) {
	c := check.New(t)

	in := [3]int32{1, 2, 3}
	out := [3]int32{}

	trampoline(c, &in, &out, nil)
}

func TestSeparator(t *testing.T) {
	c := check.New(t)

	pie := Pie{
		A: "apple",
		B: 1,
		C: "pumpkin",
		D: 2,
		E: "berry",
		F: 3,
	}

	s := NewSeparator('#')

	p, err := s.Marshal(pie)
	c.MustNotError(err)

	p2 := Pie{}
	err = s.Unmarshal(p, &p2)
	c.MustNotError(err)
	c.Equal(pie, p2)
}

func TestMarshalInto(t *testing.T) {
	c := check.New(t)

	b := bytes.Buffer{}
	pie := Pie{}
	MustMarshalInto(pie, &b)
	err := MarshalInto(pie, &b)
	c.MustNotError(err)

	p2 := Pie{}
	b2 := bytes.NewBuffer(b.Bytes())
	MustUnmarshalFrom(b2, &p2)
	err = UnmarshalFrom(&b, &p2)
	c.MustNotError(err)

	c.Equal(pie, p2)
}

func TestUnsupportedMarshalTypes(t *testing.T) {
	c := check.New(t)

	tests := []interface{}{
		make(chan int),
		[]int{1, 2, 3},
		struct{ Ch chan int }{},
		[...]chan struct{}{make(chan struct{}), make(chan struct{})},
	}

	for i, t := range tests {
		_, err := Marshal(t)
		c.Error(err, "%d did not error", i)
	}
}

func TestUnsupportedUnmarshalTypes(t *testing.T) {
	c := check.New(t)

	tests := []interface{}{
		&[]int{1, 2, 3},
		&[...]int{1, 2, 3},
		&struct{ Ch chan int }{},
		&[...]chan struct{}{make(chan struct{}), make(chan struct{})},
	}

	for i, t := range tests {
		err := Unmarshal([]byte("/test/"), t)
		c.Error(err, "%d did not error", i)
	}
}

func TestInvalidString(t *testing.T) {
	c := check.New(t)

	tests := []interface{}{
		"string/with/slashes",
		[]byte("string/with/slashes"),
	}

	for i, t := range tests {
		_, err := Marshal(t)
		c.Error(err, "%d did not error", i)
	}
}

func TestTruncatedInts(t *testing.T) {
	c := check.New(t)

	tests := []interface{}{
		new(int8),
		new(int16),
		new(int32),
		new(int64),
		new(uint8),
		new(uint16),
		new(uint32),
		new(uint64),
		struct{ A uint64 }{},
	}

	for i, t := range tests {
		err := Unmarshal([]byte("/"), t)
		c.Error(err, "%d did not fail", i)
	}
}

func TestUnmarshalWithoutPointer(t *testing.T) {
	c := check.New(t)
	v := struct{ A int }{}

	err := Unmarshal(nil, v)
	c.Error(err)
}

func TestWrongTag(t *testing.T) {
	c := check.New(t)

	in := struct {
		_ Static `path:"in"`
	}{}
	out := struct {
		_ Static `path:"out"`
	}{}

	b, err := Marshal(in)
	c.MustNotError(err)

	err = Unmarshal(b, &out)
	c.Error(err)
}

func TestErrors(t *testing.T) {
	c := check.New(t)

	pie := Pie{
		A: "apple",
		B: 1,
		C: "pumpkin",
		D: 2,
		E: "berry",
		F: 3,
	}

	p, err := Marshal(pie)
	c.MustNotError(err)

	err = nil
	p2 := Pie{}
	for i := range p {
		err = UnmarshalFrom(bytes.NewBuffer(p[0:i]), &p2)
		c.Error(err)
	}

	err = UnmarshalFrom(bytes.NewBuffer(p), &p2)
	c.NotError(err)
}

func TestExpectTagCoverage(t *testing.T) {
	c := check.New(t)
	c.Error(defSep.ExpectTag("test", bytes.NewBufferString("/hai/")))
	c.Error(defSep.ExpectTagBytes([]byte("test"), bytes.NewBufferString("hai")))
	c.Error(defSep.ExpectTagBytes([]byte("test"), bytes.NewBufferString("/hai/")))
}

func BenchmarkMarshal(b *testing.B) {
	pie := Pie{
		A: "apple",
		B: 1,
		C: "apple",
		D: 1,
		E: "apple",
		F: 1,
	}

	buff := bytes.Buffer{}
	for i := 0; i < b.N; i++ {
		buff.Reset()
		MustMarshalInto(pie, &buff)
	}
}

func BenchmarkMarshalPath(b *testing.B) {
	pie := PieWithInterface{
		Pie{
			A: "apple",
			B: 1,
			C: "apple",
			D: 1,
			E: "apple",
			F: 1,
		},
	}

	buff := bytes.Buffer{}
	for i := 0; i < b.N; i++ {
		buff.Reset()
		MustMarshalInto(pie, &buff)
	}
}

func BenchmarkUnmarshal(b *testing.B) {
	c := check.New(b)

	pie := Pie{
		A: "apple",
		B: 1,
		C: "apple",
		D: 1,
		E: "apple",
		F: 1,
	}

	pb, err := Marshal(pie)
	c.MustNotError(err)

	for i := 0; i < b.N; i++ {
		p2 := Pie{}
		MustUnmarshal(pb, &p2)
	}
}

func BenchmarkUnmarshalPath(b *testing.B) {
	c := check.New(b)

	pie := PieWithInterface{
		Pie{
			A: "apple",
			B: 1,
			C: "apple",
			D: 1,
			E: "apple",
			F: 1,
		},
	}

	pb, err := Marshal(pie)
	c.MustNotError(err)

	for i := 0; i < b.N; i++ {
		p2 := PieWithInterface{}
		MustUnmarshal(pb, &p2)
	}
}

func Example_usage() {
	v := struct {
		Type         string
		Index        uint32
		Desirability uint64
	}{
		Type:         "apple",
		Index:        123,
		Desirability: 99828,
	}

	// This path isn't human-readable, so printing it anywhere it pointless
	path := MustMarshal(v)

	// Reset everything, just in case
	v.Type = ""
	v.Index = 0
	v.Desirability = 0

	MustUnmarshal(path, &v)
	fmt.Println("Type:", v.Type)
	fmt.Println("Index:", v.Index)
	fmt.Println("Desirability:", v.Desirability)

	// Output:
	// Type: apple
	// Index: 123
	// Desirability: 99828
}

func ExampleStatic() {
	v := struct {
		_     Static `path:"pie"` // Prefix this path with "/pie/"
		Type  string
		Taste string
	}{
		Type:  "apple",
		Taste: "fantastic",
	}

	path := MustMarshal(v)
	fmt.Println(string(path))

	// Output:
	// /pie/apple/fantastic/
}