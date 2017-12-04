package helper

import (
	"crypto/md5"
	"encoding/hex"
	"strconv"
	"time"
)

func MakeGroupToken(user_token string) (token string) {
	timestamp := time.Now().Unix()
	tmpstr := user_token + strconv.FormatInt(timestamp, 10)
	h := md5.New()
	h.Write([]byte(tmpstr))
	token = hex.EncodeToString(h.Sum(nil))
	// token = Substr(token_str, 0, 16)
	return
}

func Substr(str string, start, length int) string {
	rs := []rune(str)
	rl := len(rs)
	end := 0
	if start < 0 {
		start = rl - 1 + start
	}
	end = start + length
	if start > end {
		start, end = end, start
	}
	if start < 0 {
		start = 0
	}
	if start > rl {
		start = rl
	}
	if end < 0 {
		end = 0
	}
	if end > rl {
		end = rl
	}

	return string(rs[start:end])
}
