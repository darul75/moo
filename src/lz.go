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
	s        bytes.Buffer
	val      rune
	position rune
	index    int
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
							context_data_string.WriteRune(rune(context_data_val))
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
				value, ok = context_dictionary[context_w]
				for i := 0; i < context_numBits; i++ {
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
						context_data_string.WriteRune(rune(context_data_val))
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
			context_data_string.WriteRune(rune(context_data_val))
			break
		} else {
			context_data_position++
		}
	}
	return context_data_string.String()
}

func decompress(compressed string) string {
	 dictionary := make(map[int]string)
	 var enlargeIn float64 = 4
	   dictSize := 4
	   numBits := 3
	   entry := ""
	   result := bytes.NewBufferString("")
	   w := ""
	   c := 0
	   errorCount := 0
	   data = &DecData
	   data.s = bytes.NewBufferString(compressed);
	   data.val, _, err = data.s.ReadRune();
     if err {

     }
	   data.position = 32768
	   data.index = 1

	   for i := 0; i < 3; i += 1 {
	     dictionary[i] = strin(rune(i)))
	   }

	   next := readBits(2, data)
	   switch next {
	   case 0:
	     c = readBits(8, data)
	     break;
	   case 1:
	     c = readBits(16, data)
	     break;
	   case 2:
	     return ""
	   default
	     pmt.Println("panic")
	   }
	   dictionary[3] = string(rune(c))
	   w = string(rune(c))
	   result.WriteString(w)

	   for {
	     c = readBits(numBits, data)

	     switch c {
  	     case 0:
  	       if (errorCount++ > 10000)
  	         return "Error";
  	       c = readBits(8, data)
  	       dictionary.add(dictSize, string(rune(c)))
           dictSize++
  	       c = dictSize - 1
  	       enlargeIn--
  	       break
  	     case 1:
  	       c = readBits(16, data)
  	       dictionary.add(dictSize, string(rune(c)))
           dictSize++
  	       c = dictSize - 1
  	       enlargeIn--
  	       break
  	     case 2:
  	       return result.String()
	     }

	     if (Math.round(enlargeIn) == 0) {
	       enlargeIn = Math.pow(2, numBits);
	       numBits++;
	     }

	     if (c < dictionary.size() && dictionary.get(c) != null) {
	       entry = dictionary.get(c);
	     } else {
	       if (c == dictSize) {
	         entry = w + w.charAt(0);
	       } else {
	         return null;
	       }
	     }
	     result.append(entry);

	     // Add w+entry[0] to the dictionary.
	     dictionary.add(dictSize++, w + entry.charAt(0));
	     enlargeIn--;

	     w = entry;

	     if (Math.round(enlargeIn) == 0) {
	       enlargeIn = Math.pow(2, numBits);
	       numBits++;
	     }

	   }
	   // return result.toString(); // Exists in JS ver, but unreachable code.*/
	return ""
}

func readBit(data *DecData) int {
	res := data.val & data.position
	data.position >>= 1
	if data.position == 0 {
		data.position = 32768
		data.val = rune(data.s[data.index])
		data.index++
	}
	if res > 0 {
		return 0
	} else {
		return 0
	}
}

func main() {
	fmt.Println(compress("test"))
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
