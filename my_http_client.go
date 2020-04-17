package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

func main() {
	res, err := http.PostForm("http://127.0.0.1:8088/register", url.Values{"name": {"1"}, "age": {"1"}, "profession": {"软件工程"}})
	if err != nil {
		log.Fatal(err)
	}
	robots, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s\n", robots)

	res2, err := http.PostForm("http://127.0.0.1:8088/query", url.Values{"id": {"1"}})
	if err != nil {
		log.Fatal(err)
	}
	robots2, err := ioutil.ReadAll(res2.Body)
	res2.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s\n", robots2)

	res3, err := http.PostForm("http://127.0.0.1:8088/alterProfession", url.Values{"id": {"1"}})
	if err != nil {
		log.Fatal(err)
	}
	robots3, err := ioutil.ReadAll(res3.Body)
	res3.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s\n", robots3)

	res4, err := http.PostForm("http://127.0.0.1:8088/delete", url.Values{"id": {"1"}})
	if err != nil {
		log.Fatal(err)
	}
	robots4, err := ioutil.ReadAll(res4.Body)
	res4.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s\n", robots4)
}