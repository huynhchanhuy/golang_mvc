package stringer

import(
	"io"
	"crypto/rand"
)

var alphaString = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
var alnumString = []byte("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
var lowalnumString = []byte("0123456789abcdefghijklmnopqrstuvwxy")
var numericString = []byte("0123456789")
var nozeroString = []byte("123456789")

func RandomStr(length int, types string) string {
	if types == "nozero" {
		return rand_char(length, nozeroString)
	} else if types == "alnum" {
		return rand_char(length, alnumString)
	} else if types == "lowalnum" {
		return rand_char(length, lowalnumString)
	} else if types == "numeric" {
		return rand_char(length, numericString)
	} else {
		return rand_char(length, alphaString)
    }
}

func rand_char(length int, chars []byte) string {
    new_pword := make([]byte, length)
    random_data := make([]byte, length+(length/4)) // storage for random bytes.
    clen := byte(len(chars))
    maxrb := byte(256 - (256 % len(chars)))
    i := 0
    for {
        if _, err := io.ReadFull(rand.Reader, random_data); err != nil {
            panic(err)
        }
        for _, c := range random_data {
            if c >= maxrb {
                continue
            }
            new_pword[i] = chars[c%clen]
            i++
            if i == length {
                return string(new_pword)
            }
        }
    }
    panic("unreachable")
}