package main

// http://pieroxy.net/blog/pages/lz-string/index.html

import (
	"bytes"
	"fmt"
	"io"
	"math"
)

type Context struct {
	dictionary         map[string]rune
	dictionaryToCreate *Hashset
	c                  string
	wc                 string
	w                  string
	enlargeIn          float64
	dictSize           int
	numBits            int
	data               *Data
	//Data data = new Data();
}

type Data struct {
	val      rune
	position int
	s        bytes.Buffer
}

type DecData struct {
	s        *bytes.Buffer
	val      rune
	position rune
	index    int64
}

func writeBit(value rune, data *Data) {
	data.val = rune((data.val << 1) | value)
	if data.position == 15 {
		data.position = 0
		data.s.WriteRune(data.val)
		data.val = 0
	} else {
		data.position++
	}
}

func writeBits(numBits int, value rune, data *Data) {
	for i := 0; i < numBits; i++ {
		writeBit((value & 1), data)
		value = (value >> 1)
	}
}

func produceW(context *Context) {
	if context.dictionaryToCreate.Contains(context.w) {
		var firstChar rune = rune(context.w[0])
		if firstChar < 256 {
			writeBits(context.numBits, 0, context.data)
			writeBits(8, firstChar, context.data)
		} else {
			writeBits(context.numBits, 1, context.data)
			writeBits(16, firstChar, context.data)
		}
		decrementEnlargeIn(context)
		context.dictionaryToCreate.Remove(context.w)
	} else {
		writeBits(context.numBits, context.dictionary[context.w], context.data)
	}
	decrementEnlargeIn(context)
}

func decrementEnlargeIn(context *Context) {
	context.enlargeIn--
	if int(context.enlargeIn) == 0 {
		context.enlargeIn = math.Pow(2, float64(context.numBits))
		context.numBits++
	}
}

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
			if err != io.EOF {
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
		if _, ok := context_dictionary[context_wc]; ok {
			context_w = context_wc
		} else {
			if context_dictionaryToCreate.Contains(context_w) {
				if int(context_w[0]) < 256 {
					for i := 0; i < context_numBits; i++ {
						context_data_val = context_data_val << 1
						if context_data_position == 15 {
							context_data_position = 0
							context_data_string.WriteString(string(context_data_val))
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
							context_data_string.WriteString(string(context_data_val))
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
							context_data_string.WriteString(string(context_data_val))
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
							context_data_string.WriteString(string(context_data_val))
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
						context_data_string.WriteString(string(context_data_val))
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
		if context_dictionaryToCreate.Contains(context_w) {
			if int(context_w[0]) < 256 {
				for i := 0; i < context_numBits; i++ {
					context_data_val = context_data_val << 1
					if context_data_position == 15 {
						context_data_position = 0
						context_data_string.WriteString(string(context_data_val))
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
						context_data_string.WriteString(string(context_data_val))
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
						context_data_string.WriteString(string(context_data_val))
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
						context_data_string.WriteString(string(context_data_val))
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
					context_data_string.WriteString(string(context_data_val))
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
			context_data_string.WriteString(string(context_data_val))
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
			context_data_string.WriteString(string(context_data_val))
			break
		} else {
			context_data_position++
		}
	}
	fmt.Println("%x", context_data_string.String())
	return context_data_string.String()
}

// https://gist.github.com/DavidVaini/10308388
func Round(f float64) float64 {
	return math.Floor(f + .5)
}

func decompress(compressed string) string {
	dictionary := make(map[int]string)
	// var enlargeIn float64 = 4
	// dictSize := 4
	numBits := 3
	// entry := ""
	result := bytes.NewBufferString("")
	// w := ""
	c := 0
	// errorCount := 0
	data := &DecData{}
	data.s = bytes.NewBufferString(compressed)
	val, size, _ := data.s.ReadRune()
	data.val = val
	fmt.Println("val ", val)
	fmt.Println(size)
	data.position = 32768
	data.index = 1

	fmt.Println(data.s.String())
	fmt.Println(uint8([]byte(data.s.String())[0]))

	for i := 0; i < 3; i += 1 {
		dictionary[i] = string(i)
	}

	fmt.Println("init position", data.position)
	next := readBits(2, data)
	switch next {
	case 0:
		fmt.Println("case 0")
		fmt.Println("before data.val %v", data.val)
		fmt.Println("before data.position %v", data.position)
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
	// w = string(c)
	result.WriteString(dictionary[3])
	fmt.Println(dictionary[3])

	for {
		//fmt.Println("numBits %v data %v", numBits, data)
		c = readBits(numBits, data)

		break

		// switch c {
		// case 0:
		// 	if errorCount > 10000 {
		// 		return "Error"
		// 	}

		// 	errorCount++
		// 	c = readBits(8, data)
		// 	fmt.Println("cc %v", c)
		// 	dictionary[dictSize] = string(c)
		// 	dictSize++
		// 	c = dictSize - 1
		// 	enlargeIn--
		// 	break
		// case 1:
		// 	c = readBits(16, data)
		// 	dictionary[dictSize] = string(c)
		// 	dictSize++
		// 	c = dictSize - 1
		// 	enlargeIn--
		// 	break
		// case 2:
		// 	return result.String()
		// }

		// if int(enlargeIn) == 0 {
		// 	enlargeIn = math.Pow(2, float64(numBits))
		// 	numBits++
		// }

		// _, ok := dictionary[c]

		// if c < len(dictionary) && ok {
		// 	entry = dictionary[c]
		// } else {
		// 	if c == dictSize {
		// 		entry = w + string(w[0])
		// 	} else {
		// 		return ""
		// 	}
		// }
		// //fmt.Println("entry %v result %v", string(entry), result)
		// result.WriteString(string(entry))

		// // Add w+entry[0] to the dictionary.
		// dictionary[dictSize] = w + string(entry[0])
		// dictSize++
		// enlargeIn--

		// w = entry

		// if int(enlargeIn) == 0 {
		// 	enlargeIn = math.Pow(2, float64(numBits))
		// 	numBits++
		// }

	}
	// return result.toString(); // Exists in JS ver, but unreachable code.*/
	return ""
}

func readBit(data *DecData) int {
	res := data.val & data.position
	fmt.Println("data.val %v", data.val)
	fmt.Println("data.position %v", data.position)
	data.position >>= 1
	if data.position == 0 {
		data.position = 32768
		str := data.s.String()
		//fmt.Println("dddd %v", data.index)
		data.val = rune(str[data.index])
		data.index++
	}
	if res > 0 {
		return 1
	} else {
		return 0
	}
}

func readBits(numBits int, data *DecData) int {
	res := 0
	maxpower := math.Pow(2, float64(numBits))
	power := 1
	// fmt.Println("max %v", maxpower)
	// fmt.Println("power %v", int(power))
	for power != int(maxpower) {
		//fmt.Println("i %v maxpo %v power %v index %v", i, int(Round(maxpower)), power, data.index)
		res |= readBit(data) * power
		//fmt.Println("power before %v", power)
		power <<= 1

		//fmt.Println("power %v", power)
	}

	fmt.Println("power %v", power)

	return res
}

func main() {
	fmt.Println(compress("test"))
	fmt.Println(decompress(compress("test")))

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
