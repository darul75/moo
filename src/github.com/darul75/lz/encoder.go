package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"math"
	"unicode/utf8"
)

func main() {
	original := "hello1hello2hello3hello4hello5hello6"
	data := &Data{original, ""}

	str := data.Encode()
	fmt.Println(len(str))
	fmt.Println(str)

	end := data.Decode()

	fmt.Println("decode", end)

	// if end != original {
	// 	//t.Errorf("OMG")
	// }
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
	fmt.Println(len(data.Encoded))
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
	dictionaryToCreate *Hashset
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
	s        bytes.Buffer
}

type DecData struct {
	s        *bytes.Reader
	val      rune
	position int
	index    int
}

// MAIN METHODS

func compress(uncompressedStr string) string {
	value := 0
	context_dictionary := make(map[string]int)
	context_dictionaryToCreate := NewSet()
	context_c, context_wc, context_w := "", "", ""
	var context_enlargeIn float64 = 2.0 // Compensate for the first entry which should not count
	context_dictSize := 3
	context_numBits := 2
	context_data_string := bytes.NewBufferString("")
	context_data_val := 0
	context_data_position := 0

	uncompressed := bytes.NewBufferString(uncompressedStr)
	for {
		v, _, err := uncompressed.ReadRune()
		if err != nil {
			fmt.Println("EOF")
			if err != io.EOF {
				fmt.Println("EOF")
				return ""
			}
			break
		}
		context_c = string(v)
		if _, ok := context_dictionary[context_c]; !ok {
			context_dictionary[context_c] = int(context_dictSize)
			context_dictSize++
			context_dictionaryToCreate.Add(context_c)
		}

		context_wc = context_w + context_c
		//fmt.Println("context_wc %v", context_wc)

		if _, ok := context_dictionary[context_wc]; ok {
			fmt.Println("ok")
			context_w = context_wc
		} else {
			if context_dictionaryToCreate.Contains(context_w) {
				if int(context_w[0]) < 256 {
					for i := 0; i < context_numBits; i++ {
						context_data_val = context_data_val << 1
						if context_data_position == 15 {
							// OK
							context_data_position = 0
							// buf := []byte{context_data_val}
							// utf8.EncodeRune(buf, rune(context_data_val))
							context_data_string.WriteRune(rune(context_data_val))
							// context_data_string.WriteByte(byte(context_data_val))
							fmt.Println("1) writeString")
							fmt.Println("size %v", len(context_data_string.String()))
							context_data_val = 0
						} else {
							context_data_position++
						}
					}
					value = int(context_w[0]) // OK
					for i := 0; i < 8; i++ {
						context_data_val = (context_data_val << 1) | (value & 1)
						if context_data_position == 15 {
							// OK
							context_data_position = 0
							context_data_string.WriteRune(rune(context_data_val))
							fmt.Println("2) writeString")
							fmt.Println("size %v", len(context_data_string.String()))
							context_data_val = 0
						} else {
							context_data_position++
						}
						value = value >> 1
					}
				} else {
					fmt.Println("else %v")
					value = 1
					for i := 0; i < context_numBits; i++ {
						context_data_val = (context_data_val << 1) | value
						if context_data_position == 15 {
							context_data_position = 0
							context_data_string.WriteByte(byte(context_data_val))
							context_data_val = 0
						} else {
							context_data_position++
						}
						value = 0
					}
					value = int(context_w[0])
					for i := 0; i < 16; i++ {
						context_data_val = (context_data_val << 1) | (value & 1)
						if context_data_position == 15 {
							context_data_position = 0
							context_data_string.WriteByte(byte(context_data_val))
							context_data_val = 0
						} else {
							context_data_position++
						}
						value = value >> 1
					}
				}
				context_enlargeIn--
				if context_enlargeIn == 0 {
					context_enlargeIn = math.Pow(2, float64(context_numBits))
					context_numBits++
				}
				context_dictionaryToCreate.Remove(context_w)
			} else {
				value, ok = context_dictionary[context_w]
				for i := 0; i < context_numBits; i++ {
					context_data_val = (context_data_val << 1) | (value & 1)
					if context_data_position == 15 {
						context_data_position = 0
						context_data_string.WriteRune(rune(context_data_val))
						fmt.Println("3) writeString")
						fmt.Println("size %v", len(context_data_string.String()))
						context_data_val = 0
					} else {
						context_data_position++
					}
					value = value >> 1
				}
			}
			context_enlargeIn--
			if context_enlargeIn == 0 {
				context_enlargeIn = math.Pow(2, float64(context_numBits))
				context_numBits++
			}
			// Add wc to the dictionary.
			context_dictionary[context_wc] = int(context_dictSize)
			context_dictSize++
			context_w = context_c
		}
	}

	// Output the code for w.
	if context_w != "" {
		fmt.Println("not empty")
		if context_dictionaryToCreate.Contains(context_w) {
			fmt.Println("vv %v", int(context_w[0]))
			if int(context_w[0]) < 256 {
				for i := 0; i < context_numBits; i++ {
					context_data_val = context_data_val << 1
					if context_data_position == 15 {
						context_data_position = 0
						context_data_string.WriteRune(rune(context_data_val))
						fmt.Println("context_data_val %v", context_data_val)
						context_data_val = 0
					} else {
						context_data_position++
					}
				}
				value = int(context_w[0])
				for i := 0; i < 8; i++ {
					context_data_val = (context_data_val << 1) | (value & 1)
					if context_data_position == 15 {
						context_data_position = 0
						context_data_string.WriteRune(rune(context_data_val))
						fmt.Println("2) writeString")
						fmt.Println("size %v", len(context_data_string.String()))
						context_data_val = 0
					} else {
						context_data_position++
					}
					value = value >> 1
				}
			} else {
				value = 1
				for i := 0; i < context_numBits; i++ {
					context_data_val = (context_data_val << 1) | value
					if context_data_position == 15 {
						context_data_position = 0
						context_data_string.WriteRune(rune(context_data_val))
						context_data_val = 0
					} else {
						context_data_position++
					}
					value = 0
				}
				value = int(context_w[0])
				for i := 0; i < 16; i++ {
					context_data_val = (context_data_val << 1) | (value & 1)
					if context_data_position == 15 {
						context_data_position = 0
						context_data_string.WriteRune(rune(context_data_val))
						context_data_val = 0
					} else {
						context_data_position++
					}
					value = value >> 1
				}
			}
			context_enlargeIn--
			if context_enlargeIn == 0 {
				context_enlargeIn = math.Pow(2, float64(context_numBits))
				context_numBits++
			}
			context_dictionaryToCreate.Remove(context_w)
		} else {
			value, _ := context_dictionary[context_w]
			for i := 0; i < context_numBits; i++ {
				context_data_val = (context_data_val << 1) | (value & 1)
				if context_data_position == 15 {
					context_data_position = 0
					fmt.Println("writeString")
					context_data_string.WriteRune(rune(context_data_val))
					context_data_val = 0
				} else {
					context_data_position++
				}
				value = value >> 1
			}

		}
		context_enlargeIn--
		if context_enlargeIn == 0 {
			context_enlargeIn = math.Pow(2, float64(context_numBits))
			context_numBits++
		}
	}

	// Mark the end of the stream
	value = 2
	for i := 0; i < context_numBits; i++ {
		context_data_val = (context_data_val << 1) | (value & 1)
		if context_data_position == 15 {
			context_data_position = 0
			fmt.Println("writeString")
			context_data_string.WriteRune(rune(context_data_val))
			context_data_val = 0
		} else {
			context_data_position++
		}
		value = value >> 1
	}

	// Flush the last char
	for {
		context_data_val = (context_data_val << 1)
		if context_data_position == 15 {
			fmt.Println("writeString")
			context_data_string.WriteRune(rune(context_data_val))
			break
		} else {
			context_data_position++
		}
	}

	fmt.Println("bytes %v", []byte(context_data_string.String()))
	return context_data_string.String()
}

// https://gist.github.com/DavidVaini/10308388
func Round(f float64) float64 {
	return math.Floor(f + .5)
}

func decompress(compressed string) string {
	fmt.Println("bytes %v", []byte(compressed))
	runeCount := utf8.RuneCountInString(compressed)
	fmt.Println(runeCount)
	dictionary := make(map[string]string)
	var enlargeIn float64 = 4
	dictSize := 4
	numBits := 3
	entry := ""
	result := bytes.NewBufferString("")
	w := ""
	var c int

	data := &DecData{}
	data.s = bytes.NewReader([]byte(compressed))
	data.position = 32768
	val, _, err := data.s.ReadRune()

	originalValue := val

	if err == io.EOF {
		fmt.Println(err) // EOF
		return ""
	}
	data.val = val
	data.index = 1
	fmt.Println("data.val %v", data.val)

	dictionary["0"] = "0"
	dictionary["1"] = "1"
	dictionary["2"] = "2"

	// test := data.s.UnreadRune()
	// if test != nil {
	// 	fmt.Println("ok")
	// }

	fmt.Println("init c %v", originalValue)

	next := readBits(2, data)
	fmt.Println("next", string(next))
	switch next {
	case 0:
		c = readBits(8, data)
		fmt.Println("ccedille %v", c)
		break
	case 1:
		c = readBits(16, data)
		break
	case 2:
		fmt.Println("case 2")
		return ""
	default:
		fmt.Println("panic")
	}

	// test2 := data.s.UnreadRune()
	// if test2 != nil {
	// 	fmt.Println("ok")
	// }

	dictionary["3"] = string(c)
	w = string(c)
	result.WriteString(w)
	fmt.Println(dictionary)
	i := 0

	for {
		i++
		c = readBits(numBits, data)
		fmt.Println("c %v", byte(c))

		switch c {
		case 0:
			c = readBits(8, data)
			dictionary[string(dictSize)] = string(c)
			dictSize++
			c = (dictSize - 1)
			enlargeIn--
			break
		case 1:
			c = readBits(16, data)
			dictionary[string(dictSize)] = string(c)
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

		_, ok := dictionary[string(c)]

		if ok {
			entry = dictionary[string(c)]
			//fmt.Println("contains %v", entry)
		} else {
			// fmt.Println("else")
			fmt.Println("c %v", c)
			fmt.Println("len %v", len(dictionary))
			if int(c) == len(dictionary) {
				entry = w + string(w[0])
			} else {
				fmt.Println("i %v", i)
				return ""
			}
		}
		//fmt.Println("entry %v result %v", string(entry), result)
		result.WriteString(string(entry))

		// Add w+entry[0] to the dictionary.
		dictionary[string(dictSize)] = w + string(entry[0])
		dictSize++
		enlargeIn--

		w = entry

		fmt.Println("dictionary %v", w+string(entry[0]))

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

func decrementEnlargeIn(context *Context) {
	context.enlargeIn--
	if int(context.enlargeIn) == 0 {
		context.enlargeIn = math.Pow(2, float64(context.numBits))
		context.numBits++
	}
}

// BIT EXTRACTION

func readBit(data *DecData) int {
	res := int(data.val) & data.position
	data.position >>= 1

	if data.position == 0 {
		data.position = 32768
		val, size, _ := data.s.ReadRune()
		size++ // not used...
		data.val = val
		//data.index += size
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
