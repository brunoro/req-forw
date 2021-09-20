package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	confPath := flag.String("config", "", "path to the JSON config file")
	flag.Parse()

	if confPath == nil || *confPath == "" {
		log.Println("Missing -config command-line argument")
		return
	}

	conf, err := readConfigFile(*confPath)
	if err != nil {
		log.Println(err.Error())
		return
	}

	maxRetries := 10

	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	for _, fwdConf := range conf.Forward {
		port, target := fwdConf.Port, fwdConf.Target
		go func() {
			if err := forward(port, target, maxRetries); err != nil {
				log.Printf("ERROR initializing :%d -> %s\n%s\n",
					port, target, err.Error())
			}
			done <- true
		}()
	}

	go func() {
		sig := <-sigs
		fmt.Println()
		fmt.Println(sig)
		done <- true
	}()

	<-done
	fmt.Println("Exiting")
}
