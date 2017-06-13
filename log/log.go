package log

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"time"
)

func checkFileIsExist(filename string) bool {
	var exist = true
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		exist = false
	}
	return exist
}

func NewLog(comment string, data interface{}) {
	var f *os.File
	var err error
	log_file := "log/" + strconv.Itoa(time.Now().Year()) + time.Now().Month().String() + ".log"
	if checkFileIsExist(log_file) { //如果文件存在
		f, err = os.OpenFile(log_file, os.O_APPEND|os.O_WRONLY, os.ModeAppend) //打开文件
		if err != nil {
			panic(err)
		}
		// fmt.Println("文件存在")
	} else {
		f, err = os.Create(log_file) //创建文件
		if err != nil {
			panic(err)
		}
		// fmt.Println("文件不存在")
	}
	io.WriteString(f, "[@@@"+time.Now().String()+"@@@-###"+comment+"###]------"+fmt.Sprint(data)+"\r\n") //写入文件(字符串)
	// fmt.Printf("写入 %d 个字节n", n)
}
