package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"math"
	"strings"
	"unicode/utf8"
)

func main() {

	notworking := "hello1hello2hello3hello4hello5hello6"
	working := "hello1hello2hello3hello4hello5"

	fmt.Println("ORIGINAL %v", notworking)
	data := &Data{notworking, ""}
	data.Encode()

	fmt.Println("DECODED", data.Decode())

	fmt.Println("ORIGINAL %v", working)
	data2 := &Data{working, ""}
	data2.Encode()
	fmt.Println("DECODED", data2.Decode())

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
	dictionary         map[string]rune
	dictionaryToCreate Hashset
	c                  string
	wc                 string
	w                  string
	enlargeIn          float64
	dictSize           int
	numBits            int
	data               *EncData
}

type EncData struct {
	val      rune
	position int
	s        *bytes.Buffer
}

type DecData struct {
	//s        *bytes.Reader
	s        []rune
	val      rune
	position int
	index    int
}

// MAIN METHODS

func compress(uncompressedStr string) string {
	set := NewSet()
	ctx := &Context{}
	ctx.dictionary = make(map[string]rune)
	ctx.dictionaryToCreate = set
	ctx.c = ""
	ctx.wc = ""
	ctx.w = ""
	ctx.enlargeIn = 2.0
	ctx.dictSize = 3
	ctx.numBits = 2
	ctx.data = &EncData{0, 0, bytes.NewBufferString("")}

	//uncompressed := bytes.NewBufferString(uncompressedStr)

	runeCount := utf8.RuneCountInString(uncompressedStr)
	runes := make([]rune, runeCount)
	k := 0
	for len(uncompressedStr) > 0 {
		r, size := utf8.DecodeRuneInString(uncompressedStr)
		runes[k] = r
		uncompressedStr = uncompressedStr[size:]
		k++
	}

	for _, v := range runes {

		ctx.c = string(v)
		if _, ok := ctx.dictionary[ctx.c]; !ok {
			ctx.dictionary[ctx.c] = rune(ctx.dictSize)
			ctx.dictSize++
			ctx.dictionaryToCreate.Add(ctx.c)
		}

		ctx.wc = string(ctx.w + ctx.c)

		if _, ok := ctx.dictionary[ctx.wc]; ok {
			ctx.w = ctx.wc
		} else {
			produceW(ctx)
			// Add wc to the dictionary.
			ctx.dictionary[ctx.wc] = rune(ctx.dictSize)
			ctx.dictSize += 1
			ctx.w = string(ctx.c)
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

	return ctx.data.s.String()
}

func produceW(ctx *Context) {
	if ctx.dictionaryToCreate.Contains(ctx.w) {
		var firstChar rune = rune(ctx.w[0])
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
		writeBits(ctx.numBits, rune(ctx.dictionary[ctx.w]), ctx.data)
	}
	decrementEnlargeIn(ctx)
}

func writeBits(numBits int, value rune, data *EncData) {
	for i := 0; i < numBits; i++ {
		writeBit(rune(value&1), data)
		value = rune(value >> 1)
	}
}

func decrementEnlargeIn(ctx *Context) {
	ctx.enlargeIn--
	if int(ctx.enlargeIn) == 0 {
		ctx.enlargeIn = math.Pow(2, float64(ctx.numBits))
		ctx.numBits++
	}
}

func writeBit(value rune, data *EncData) {
	data.val = rune((data.val << 1) | value)
	if data.position == 15 {
		data.position = 0
		data.s.WriteRune(rune(data.val))
		data.val = 0
	} else {
		data.position++
	}
}

// https://gist.github.com/DavidVaini/10308388
func Round(f float64) float64 {
	return math.Floor(f + .5)
}

func decompress(compressed string) string {
	fmt.Println(compressed)

	runeCount := utf8.RuneCountInString(compressed)
	runes := make([]rune, runeCount)

	// compute rune array
	k := 0
	for len(compressed) > 0 {
		r, size := utf8.DecodeRuneInString(compressed)
		runes[k] = r
		compressed = compressed[size:]
		k++
	}

	dictionary := make(map[int]string)
	var enlargeIn float64 = 4
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
	/*fmt.Println(w)
	fmt.Println(result.String())
	fmt.Println(dictionary)*/

	i := 0

	for {
		i++
		c = readBits(numBits, data)

		switch c {
		case 0:
			c = readBits(8, data)
			dictionary[dictSize] = string(c)
			dictSize++
			c = (dictSize - 1)
			enlargeIn--
			break
		case 1:
			c = readBits(16, data)
			dictionary[dictSize] = string(c)
			dictSize++
			c = (dictSize - 1)
			enlargeIn--
			break
		case 2:
			fmt.Println("********** 2 ***********")
			return result.String()
		}

		if int(enlargeIn) == 0 {
			enlargeIn = math.Pow(2, float64(numBits))
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

		if int(enlargeIn) == 0 {
			enlargeIn = math.Pow(2, float64(numBits))
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
	res := int(data.val) & data.position
	data.position >>= 1
	fmt.Println(res)
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
	var res int = 0
	maxpower := math.Pow(2, float64(numBits))
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
