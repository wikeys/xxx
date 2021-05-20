package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/tjfoc/gmsm/sm4"
)

func main() {

	key := []byte("1234567890abcdef")              // 生成密钥对
	src := "https://127.0.0.1/download/beacon.bin" //你的bin
	var msg []byte
	var CL http.Client
	resp, err := CL.Get(src)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {

		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		msg = bodyBytes
	}

	ecbDec, err := sm4.Sm4Ecb(key, msg, true) //sm4Ecb模式pksc7填充加密
	fmt.Println(err)
	ioutil.WriteFile("testsbin", ecbDec, 0644)

}
