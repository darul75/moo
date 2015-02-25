package main

import (
	"bytes"
	"code.google.com/p/intmath/intgr"
	"encoding/base64"
	//"encoding/binary"
	"fmt"
	"strings"
	//"unicode/utf16"
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
	val      int
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
	ctx.dictSize = uint16(3)
	ctx.numBits = 2
	ctx.data = &EncData{0, 0, bytes.NewBufferString("")}

	runeCount := utf8.RuneCountInString(uncompressedStr)
	runes := make([]uint16, runeCount)
	k := 0
	for len(uncompressedStr) > 0 {
		r, size := utf8.DecodeRuneInString(uncompressedStr)
		runes[k] = uint16(r)
		uncompressedStr = uncompressedStr[size:]
		k++
	}

	for _, v := range runes {

		ctx.c = string(v)
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

	// Mark the end of the stream
	writeBits(ctx.numBits, 2, ctx.data)

	for {
		if ctx.data.val <= 0 {
			break
		}

		writeBit(0, ctx.data)
	}

	fmt.Println("compressed string length %v", len(ctx.data.s.String()))
	fmt.Println("compressed string %q", ctx.data.s.String())

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

func writeBit(value uint16, data *EncData) {
	data.val = (data.val << 1) | value
	if data.position == 15 {
		data.position = 0
		fmt.Println("written rune %v", data.val)
		var l, h uint8 = uint8(data.val >> 8), uint8(data.val & 0xff)
		buf := make([]byte, 2)
		buf[0] = l
		buf[1] = h
		// TRY WRITE THESE 2 BYTES INSTEAD ABOVE...
		//data.s.Write(buf)
		// TRY UTF 16 ?
		/*r1, r2 := utf16.EncodeRune(rune(data.val))
		data.s.WriteRune(r1)
		data.s.WriteRune(r2)		*/
		// NOW
		data.s.WriteRune(rune(data.val))
		data.val = 0
	} else {
		data.position++
	}
}

func decrementEnlargeIn(ctx *Context) {
	ctx.enlargeIn--
	if ctx.enlargeIn == 0 {
		ctx.enlargeIn = intgr.Pow(2, int(ctx.numBits))
		ctx.numBits++
	}
}

func decompress(compressed string) string {
	fmt.Println("compressed bytes in decomp %v", []byte(compressed))
	fmt.Println("rune count %v", utf8.RuneCountInString(compressed))

	runes := make([]rune, len(compressed))

	// SEE WHAT INSIDE
	j := 0
	for index, runeValue := range compressed {
		fmt.Printf("%#U starts at byte position %d\n", runeValue, index)
		runes[j] = runeValue
		j++
	}
	/*for k, w := 0, 0; k < len(compressed); k += w {
		runeValue, width := utf8.DecodeRuneInString(compressed[k:])
		fmt.Printf("%#U starts at byte position %d\n", runeValue, k)
		w = width
	}*/

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
	data.val = int(runes[0])
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
		size++

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
		data.val = int(data.s[data.index])
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
