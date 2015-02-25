package main

import (
	"bytes"
	"code.google.com/p/intmath/intgr"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"strings"
	"unicode/utf8"
)

func main() {

	notworking := "hello1hello2hello3hello4hello5hello6"
	//	working := "hello1hello2hello3hello4hello5"

	fmt.Println("ORIGINAL %v", notworking)
	data := &Data{notworking, ""}
	data.Encode()

	/*[ 1413,
	  12342,
	  24822,
	  2250,
	  6402,
	  25492,
	  3276,
	  51205,
	  40197,
	  26432,
	  55424 ]*/

	fmt.Println("DECODED", data.Decode())

	/*	fmt.Println("ORIGINAL %v", working)
		data2 := &Data{working, ""}
		data2.Encode()
		fmt.Println("DECODED", data2.Decode())*/

}

// USEFUL DOC
// http://www.goinggo.net/2014/03/exportedunexported-identifiers-in-go.html
// http://pieroxy.net/blog/pages/lz-string/index.html

// INTERFACE

type Lz interface {
	Encode(value string) string
	Encode64(value string) string
	Decode(value string) string
	Decode64(value string) string
}

// WRAPPER

type Data struct {
	Value   string
	Encoded string
}

func (data *Data) Encode() string {
	data.Encoded = compress(data.Value)
	return data.Encoded
}

func (data *Data) Encode64() string {

	data.Encoded = compress64(data.Value)
	return data.Encoded
}

func (data *Data) Decode() string {
	return decompress(data.Encoded)
}

func (data *Data) Decode64() string {
	return decompress64(data.Encoded)
}

// LZ ALGO STRUCT

type Context struct {
	dictionary         map[string]uint16
	dictionaryToCreate Hashset
	c                  string
	wc                 string
	w                  string
	enlargeIn          int
	dictSize           uint16
	numBits            uint16
	data               *EncData
}

type EncData struct {
	val      uint16
	position uint16
	s        *bytes.Buffer
}

type DecData struct {
	//s        *bytes.Reader
	s        []rune
	val      rune
	position uint16
	index    int
}

// MAIN METHODS

func compress(uncompressedStr string) string {
	fmt.Println("len uncompressed %v", len(uncompressedStr))
	set := NewSet()
	ctx := &Context{}
	ctx.dictionary = make(map[string]uint16)
	ctx.dictionaryToCreate = set
	ctx.c = ""
	ctx.wc = ""
	ctx.w = ""
	ctx.enlargeIn = 2
	ctx.dictSize = 3
	ctx.numBits = 2
	ctx.data = &EncData{0, 0, bytes.NewBufferString("")}

	//newUtf16Bytes := Encode(uncompressedStr)
	/*r, _ := utf8.DecodeRune(newUtf16Bytes)
	fmt.Println(string(r))	*/

	//uncompressed := bytes.NewBufferString(uncompressedStr)
	/*l := 0
	for l < len(newUtf16Bytes) {
		r, size := utf8.DecodeRune(newUtf16Bytes[l : l+2])
		l += 2
	}*/

	runeCount := utf8.RuneCountInString(uncompressedStr)
	runes := make([]rune, runeCount)
	/*k := 0
	for len(uncompressedStr) > 0 {
		r, size := utf8.DecodeRuneInString(uncompressedStr)
		runes[k] = r
		uncompressedStr = uncompressedStr[size:]
		k++
	}*/

	newUtf16Bytes := Encode(uncompressedStr)
	fmt.Println(newUtf16Bytes)
	l := 0
	k := 0
	length := len(newUtf16Bytes) / 2
	for k < length {
		//fmt.Println(newUtf16Bytes[l : l+2])
		r, _ := utf8.DecodeRune(newUtf16Bytes[l : l+2])
		runes[k] = r
		//fmt.Println(r)
		l += 2
		k++
	}

	for _, v := range runes {

		ctx.c = string(v)
		//fmt.Println(string(v))
		if _, ok := ctx.dictionary[ctx.c]; !ok {
			ctx.dictionary[ctx.c] = ctx.dictSize
			ctx.dictSize++
			ctx.dictionaryToCreate.Add(ctx.c)
		}

		ctx.wc = ctx.w + ctx.c

		if _, ok := ctx.dictionary[ctx.wc]; ok {
			ctx.w = ctx.wc
		} else {
			produceW(ctx)
			// Add wc to the dictionary.
			ctx.dictionary[ctx.wc] = ctx.dictSize
			ctx.dictSize += 1
			ctx.w = ctx.c
		}
	}

	// Output the code for w.
	if ctx.w != "" {
		produceW(ctx)
	}

	fmt.Println("ctx.numBits %v", ctx.data.val)
	fmt.Println("ctx.numBits %v", ctx.data.position)
	fmt.Println("ctx.numBits %v", ctx.data.s.String())

	// Mark the end of the stream
	writeBits(ctx.numBits, 2, ctx.data)

	for {
		//fmt.Println("val %v", ctx.data.val)
		if ctx.data.val <= 0 {
			break
		}

		writeBit(0, ctx.data)

	}

	fmt.Println("compressed bytes %v", []byte(ctx.data.s.String()))

	return ctx.data.s.String()
}

func produceW(ctx *Context) {
	if ctx.dictionaryToCreate.Contains(ctx.w) {
		var firstChar uint16 = uint16(ctx.w[0])
		if firstChar < 256 {
			writeBits(ctx.numBits, 0, ctx.data)
			writeBits(8, firstChar, ctx.data)
		} else {
			writeBits(ctx.numBits, 1, ctx.data)
			writeBits(16, firstChar, ctx.data)
		}
		decrementEnlargeIn(ctx)
		ctx.dictionaryToCreate.Remove(ctx.w)
	} else {
		writeBits(ctx.numBits, ctx.dictionary[ctx.w], ctx.data)
	}
	decrementEnlargeIn(ctx)
}

func writeBits(numBits uint16, value uint16, data *EncData) {
	for i := uint16(0); i < numBits; i++ {
		writeBit(value&1, data)
		value = value >> 1
	}
}

func decrementEnlargeIn(ctx *Context) {
	ctx.enlargeIn--
	if ctx.enlargeIn == 0 {
		ctx.enlargeIn = intgr.Pow(2, int(ctx.numBits))
		ctx.numBits++
	}
}

func writeBit(value uint16, data *EncData) {
	data.val = (data.val << 1) | value
	if data.position == 15 {
		data.position = 0
		b := make([]byte, 2)
		binary.LittleEndian.PutUint16(b, data.val)
		data.s.Write(b)
		data.val = 0
	} else {
		data.position++
	}
}

func decompress(compressed string) string {
	fmt.Println("decompressed %v", []byte(compressed))
	fmt.Println("len %v", len(compressed))

	length := len(compressed) / 2

	//runeCount := utf8.RuneCountInString(compressed)
	runes := make([]rune, len([]byte(compressed)))

	// compute rune array
	/*k := 0
	for len(compressed) > 0 {
		r, size := utf8.DecodeRuneInString(compressed)
		runes[k] = r
		compressed = compressed[size:]
		k++
	}*/
	newUtf16Bytes := []byte(compressed) // Encode(compressed)
	//fmt.Println("decompressed entry bytes %v", newUtf16Bytes)
	l, idx := 0, 0
	for idx < length {

		fmt.Println("l %v idx %v value %s", l, idx, newUtf16Bytes[l:l+2])
		//r, _ := utf8.DecodeRune(newUtf16Bytes[l : l+2])
		//r := uint16(newUtf16Bytes[l : l+2])
		//buf := bytes.NewReader(newUtf16Bytes[l : l+2])
		runes[idx], _ = utf8.DecodeRuneInString(Decode(newUtf16Bytes[l : l+2]))
		fmt.Println(runes[idx])
		//fmt.Println(utf16.Decode(newUtf16Bytes[l : l+2]))
		//binary.BigEndian.Uint16(newUtf16Bytes[l : l+2])
		l += 2
		idx++
	}

	//fmt.Println(runes)

	dictionary := make(map[int]string)
	enlargeIn := 4
	dictSize := 4
	numBits := 3
	entry := ""
	result := bytes.NewBufferString("")
	w := ""
	var c int

	data := &DecData{}
	data.s = runes
	data.position = 32768
	data.val = runes[0]
	data.index = 1
	//data.index = utf8.RuneLen(data.val)

	dictionary[0] = "0"
	dictionary[1] = "1"
	dictionary[2] = "2"

	next := readBits(2, data)
	switch next {
	case 0:
		c = readBits(8, data)
		break
	case 1:
		c = readBits(16, data)
		break
	case 2:
		return ""
	default:
		fmt.Println("panic")
	}

	dictionary[3] = string(c)
	w = string(c)
	result.WriteString(w)
	fmt.Println(w)
	fmt.Println(result.String())
	fmt.Println(dictionary)

	i := 0

	for {
		i++
		c = readBits(numBits, data)

		switch c {
		case 0:
			c = readBits(8, data)
			dictionary[dictSize] = string(c)
			dictSize++
			c = dictSize - 1
			enlargeIn--
			break
		case 1:
			c = readBits(16, data)
			dictionary[dictSize] = string(c)
			dictSize++
			c = dictSize - 1
			enlargeIn--
			break
		case 2:
			fmt.Println("********** 2 ***********")
			return result.String()
		}

		if enlargeIn == 0 {
			enlargeIn = intgr.Pow(2, numBits)
			numBits++
		}

		_, ok := dictionary[int(c)]

		if ok {
			entry = dictionary[int(c)]
		} else {
			if c == dictSize {
				wOne, size, _ := strings.NewReader(w).ReadRune()
				size++
				entry = string(string(w) + string(wOne))
			} else {
				return ""
			}
		}

		result.WriteString(string(entry))

		// Add w+entry[0] to the dictionary.
		entryOne, size, _ := strings.NewReader(entry).ReadRune()
		fmt.Println("size %v", size)

		dictionary[dictSize] = w + string(entryOne)
		dictSize++
		enlargeIn--

		w = entry

		if enlargeIn == 0 {
			enlargeIn = intgr.Pow(2, numBits)
			numBits++
		}

	}

	return result.String()
}

const (
	_keyStr = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/="
)

// BASE 64 / 16
func compress64(input string) string {
	if input == "" {
		return ""
	}
	input = base64.StdEncoding.EncodeToString([]byte(input))

	fmt.Println(input)

	return compress(input)
}

func decompress64(input string) string {
	if input == "" {
		return ""
	}
	bytes, _ := base64.StdEncoding.DecodeString(input)

	fmt.Println(string(bytes))

	return decompress(string(bytes))
}

// WRITERS

// BIT EXTRACTION

func readBit(data *DecData) int {
	res := uint16(data.val) & data.position
	data.position >>= 1
	if data.position == 0 {
		data.position = 32768
		//val, size, _ := data.s.ReadRune()
		data.val = data.s[data.index]
		//size++ // not used...
		//data.val = val
		data.index += 1
	}
	if res > 0 {
		return 1
	} else {
		return 0
	}
}

func readBits(numBits int, data *DecData) int {
	res := 0
	maxpower := intgr.Pow(2, numBits)
	power := 1
	for power != int(maxpower) {
		res |= readBit(data) * power
		power <<= 1
	}

	return res
}

// UTILS

type Hashset map[string]bool

func NewSet() Hashset {
	return make(Hashset)
}

func (this Hashset) Add(value string) {
	this[value] = true
}

func (this Hashset) Remove(value string) {
	delete(this, value)
}

func (this Hashset) Contains(value string) bool {
	_, ok := this[value]
	return ok
}

func (this Hashset) Length() int {
	return len(this)
}

func (this Hashset) Union(that Hashset) Hashset {
	ns := NewSet()
	for k, v := range this {
		ns[k] = v
	}
	for k, _ := range that {
		if _, ok := this[k]; !ok {
			ns[k] = true
		}
	}
	return ns
}

func (this Hashset) Intersection(that Hashset) Hashset {
	ns := NewSet()
	for k, _ := range that {
		if _, ok := this[k]; ok {
			ns.Add(k)
		}
	}
	for k, _ := range this {
		if _, ok := that[k]; ok {
			ns.Add(k)
		}
	}
	return ns
}

// UTILS

const (
	replacementChar = '\uFFFD'     // Unicode replacement character
	maxRune         = '\U0010FFFF' // Maximum valid Unicode code point.
)

const (
	// 0xd800-0xdc00 encodes the high 10 bits of a pair.
	// 0xdc00-0xe000 encodes the low 10 bits of a pair.
	// the value is those 20 bits plus 0x10000.
	surr1 = 0xd800
	surr2 = 0xdc00
	surr3 = 0xe000

	surrSelf = 0x10000
)

// IsSurrogate returns true if the specified Unicode code point
// can appear in a surrogate pair.
func IsSurrogate(r rune) bool {
	return surr1 <= r && r < surr3
}

// DecodeRune returns the UTF-16 decoding of a surrogate pair.
// If the pair is not a valid UTF-16 surrogate pair, DecodeRune returns
// the Unicode replacement code point U+FFFD.
func DecodeRune(r1, r2 rune) rune {
	if surr1 <= r1 && r1 < surr2 && surr2 <= r2 && r2 < surr3 {
		return (rune(r1)-surr1)<<10 | (rune(r2) - surr2) + 0x10000
	}
	return replacementChar
}

// EncodeRune returns the UTF-16 surrogate pair r1, r2 for the given rune.
// If the rune is not a valid Unicode code point or does not need encoding,
// EncodeRune returns U+FFFD, U+FFFD.
func EncodeRune(r rune) (r1, r2 rune) {
	if r < surrSelf || r > maxRune || IsSurrogate(r) {
		return replacementChar, replacementChar
	}
	r -= surrSelf
	return surr1 + (r>>10)&0x3ff, surr2 + r&0x3ff
}

// Encode returns the UTF-16 encoding of the specified string str.
func Encode(s string) []byte {
	n := len(s)
	for _, v := range s {
		if v >= surrSelf {
			n++
		}
	}

	a := make([]byte, n*2)
	n = 0
	for _, v := range s {
		switch {
		case v < 0, surr1 <= v && v < surr3, v > maxRune:
			v = replacementChar
			fallthrough
		case v < surrSelf:
			a[n] = byte(v)
			a[n+1] = byte(v >> 8)
			n += 2
		default:
			r1, r2 := EncodeRune(v)
			a[n] = byte(r1)
			a[n+1] = byte(r1 >> 8)
			a[n+2] = byte(r2)
			a[n+3] = byte(r2 >> 8)
			n += 4
		}
	}
	return a[0:n]
}

// Decode returns the string represented by the UTF-16 encoding s.
func Decode(s []byte) string {
	a := make([]rune, len(s)/2)
	n := 0
	for i := 0; i < len(s); i += 2 {
		switch r := MakeUint16(s[i], s[i+1]); {
		case surr1 <= r && r < surr2 && i+3 < len(s) &&
			surr2 <= MakeUint16(s[i+2], s[i+3]) && MakeUint16(s[i+2], s[i+3]) < surr3:
			// valid surrogate sequence
			a[n] = DecodeRune(rune(r), rune(MakeUint16(s[i+2], s[i+3])))
			i++
			n++
		case surr1 <= r && r < surr3:
			// invalid surrogate sequence
			a[n] = replacementChar
			n++
		default:
			// normal rune
			a[n] = rune(r)
			n++
		}
	}
	return string(a[0:n])
}

func MakeUint16(l, h byte) uint16 {
	return uint16(l) + (uint16(h) << 8)
}

func SplitUint16(v uint16) (byte, byte) {
	return byte(v), byte(v >> 8)
}
