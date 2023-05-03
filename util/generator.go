package util

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

// LeftPad prepends the data with the padding value until the total length is reached
func LeftPad(data, padding string, totalLength int) string {
	currLength := len(data)
	for i := 0; i < (totalLength-currLength)/len(padding); i++ {
		data = padding + data
	}
	return data
}

// RightPad appends the data with the padding value until the total length is reached
func RightPad(data, padding string, totalLength int) string {
	currLength := len(data)
	for i := 0; i < (totalLength-currLength)/len(padding); i++ {
		data += padding
	}
	return data
}

// StrToHex converts the data string into multiples of mod bits and prepends the start index of data
func StrToHex(data string, mod int) string {
	dataLen := len(data)
	x := dataLen % mod
	reqdBits := mod - x
	if x != 0 {
		fillLen := dataLen + reqdBits
		data = LeftPad(data, "0", fillLen)
		reqdBits += mod
	}
	reqdBitsStr := strconv.Itoa(reqdBits)
	dummy := LeftPad(reqdBitsStr, "0", mod)
	retData := dummy + data
	return retData
}

// CreateUUID create a v4 UUID string without any "-" (hyphen)
func CreateUUID() string {
	id, err := uuid.NewUUID()
	if err != nil {
		panic(err)
	}
	return strings.Replace(id.String(), "-", "", -1)
	//return strings.Replace(id, "-", "", -1)
}

// returns a random number(string) of fixed length (n)
func GenerateRandomNumberOnlyString(n int) (str string, err error) {
	var num int64
	var nBig *big.Int
	for i := 0; i < n; i++ {
		nBig, err = rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			err = fmt.Errorf("error while creating random number: %s", err)
			return
		}
		num = nBig.Int64()
		str += strconv.Itoa(int(num))
	}
	return
}
