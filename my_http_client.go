package main

import (
	"fmt"
	"github.com/hashicorp/consul/api"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
)

var httpServerAddress string = "http://"

func httpServerDiscovery() {
	var lastIndex uint64
	config := api.DefaultConfig()
	config.Address = "127.0.0.1:8500" //consul server

	client, err := api.NewClient(config)
	if err != nil {
		log.Fatal("api new client is failed, err:", err)
		return
	}
	services, metainfo, err := client.Health().Service("httpServer", "httpServer", true, &api.QueryOptions{
		WaitIndex: lastIndex, // 同步点，这个调用将一直阻塞，直到有新的更新
	})
	if err != nil {
		logrus.Warn("error retrieving instances from Consul: %v", err)
	}
	lastIndex = metainfo.LastIndex

	//addrs := map[string]struct{}{}
	for _, service := range services {
		log.Println("service.Service.Address:", service.Service.Address, "service.Service.Port:", service.Service.Port)
		httpServerAddress = httpServerAddress + service.Service.Address + ":" + strconv.Itoa(service.Service.Port)
		//addrs[net.JoinHostPort(service.Service.Address, strconv.Itoa(service.Service.Port))] = struct{}{}
	}
}

func main() {
	httpServerDiscovery()
	res, err := http.PostForm(httpServerAddress+"/register", url.Values{"name": {"a"}, "age": {"10"}, "profession": {"软件工程"}})
	if err != nil {
		log.Fatal(err)
	}
	robots, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s\n", robots)

	res2, err := http.PostForm(httpServerAddress+"/query", url.Values{"id": {"1353252099"}})
	if err != nil {
		log.Fatal(err)
	}
	robots2, err := ioutil.ReadAll(res2.Body)
	res2.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s\n", robots2)

	res3, err := http.PostForm(httpServerAddress+"/alterProfession", url.Values{"id": {"1353252099"}, "profession": {"计算机科学与技术"}})
	if err != nil {
		log.Fatal(err)
	}
	robots3, err := ioutil.ReadAll(res3.Body)
	res3.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s\n", robots3)

	res4, err := http.PostForm(httpServerAddress+"/delete", url.Values{"id": {"1353252099"}})
	if err != nil {
		log.Fatal(err)
	}
	robots4, err := ioutil.ReadAll(res4.Body)
	res4.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s\n", robots4)

	res5, err := http.PostForm(httpServerAddress+"/queryList", url.Values{})
	if err != nil {
		log.Fatal(err)
	}
	robots5, err := ioutil.ReadAll(res5.Body)
	res5.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s\n", robots5)
}
