package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func main() {
	// parse flags
	var portString string
	portsUsage := "which http ports to listen on (https not supported)"
	flag.StringVar(&portString, "ports", "8080", portsUsage)
	flag.StringVar(&portString, "p", "8080", portsUsage+" (shorthand)")
	var host string
	hostUsage := "hostname to listen on"
	flag.StringVar(&host, "hostname", "localhost", hostUsage)
	var wait int
	waitUsage := "time (in seconds) to wait before considering server online"
	flag.IntVar(&wait, "wait", 1, waitUsage)
	flag.IntVar(&wait, "w", 1, waitUsage+" (shorthand)")
	flag.Parse()

	// process flags
	ports := []int{}
	for _, p := range strings.Split(portString, ",") {
		p := strings.Trim(p, " ")
		intp, err := strconv.Atoi(p)
		if err != nil {
			panic(err)
		}
		ports = append(ports, intp)
	}

	// validate inputs
	uniqueports := map[int]struct{}{}
	for _, p := range ports {
		before := len(uniqueports)
		uniqueports[p] = struct{}{}
		if len(uniqueports) <= before {
			panic(fmt.Sprintf("Port %d specified more than once in flags.", p))
		}
	}

	// set up routes
	http.HandleFunc("/", handleDefault)
	// http.HandleFunc("/foo", handleFoo)
	// http.HandleFunc("/bar", handleBar)

	// serve
	for _, port := range ports {
		done := make(chan struct{}, 1)
		go func(host string, port int, done chan<- struct{}) {
			s := &http.Server{
				Addr: fmt.Sprintf("%s:%d", host, port),
			}
			log.Fatal(s.ListenAndServe())
			done <- struct{}{}
			close(done)
		}(host, port, done)
		go func(host string, port int, done <-chan struct{}) {
			time.Sleep(time.Duration(wait) * time.Second)
			select {
			case <-done:
				fmt.Printf("Failed to serve on %s:%d\n", host, port)
			default:
				fmt.Printf("Serving on %s:%d\n", host, port)
			}
			<-done
			fmt.Printf("Done serving on %s:%d\n", host, port)
		}(host, port, done)
	}

	done := make(chan struct{})
	<-done
}

func handleDefault(w http.ResponseWriter, req *http.Request) {
	if err := req.ParseForm(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Parsing Error")
		return
	}
	fmt.Fprintf(w, "%s %s%s\nForm:\n%#v\n", req.Method, req.Host, req.URL, req.Form)
}
