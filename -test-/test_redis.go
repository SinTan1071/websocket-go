package main

import (
	// "encoding/json"
	"fmt"
	"github.com/garyburd/redigo/redis"
)

func main() {
	rs, _ := redis.Dial("tcp", "localhost:6379")
	rs.Do("SELECT", 1)
	defer rs.Close()
	key := "badanmu"
	value := []string{"lalal", "xixixi", "cccccc"}
	// value := "lalal"
	// _, err := rs.Do("DEL", key)
	// for i := 0; i < len(value); i++ {
	// 	_, err := rs.Do("LPUSH", key, value[i])
	// 	fmt.Println("error: ", err)
	// }
	reply, err := redis.Values(rs.Do("LRANGE", key, 0, len(value)))
	res, _ := redis.Strings(reply, err)
	for i := 0; i < len(res); i++ {
		fmt.Println("结果: ", res[i])
	}
}
