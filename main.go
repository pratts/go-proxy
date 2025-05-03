package main

import (
	"fmt"
	"goproxy/blocker"
	"goproxy/config"
	"io"
	"net"
	"net/http"
)

func main() {
	fmt.Println("Starting HTTP server on port 8080...")
	config.InitBlockList()
	http.ListenAndServe(":8080", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Serving request for URL: ", r.URL.Host, " Path: ", r.URL.Path, " Method: ", r.Method)
		if blocker.ValidateIfBlocked(r.Host) {
			fmt.Printf("Blocked request to %s\n", r.Host)
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte("Blocked host"))
			return
		}

		if r.Method == http.MethodConnect {
			fmt.Println("Received CONNECT request")
			handleHttps(w, r)
			return
		}

		newRequest, err := http.NewRequest(r.Method, r.URL.String(), r.Body)
		if err != nil {
			http.Error(w, "Failed to create new request", http.StatusInternalServerError)
			return
		}

		for key, values := range r.Header {
			for _, value := range values {
				newRequest.Header.Add(key, value)
			}
		}

		fmt.Printf("[%s] %s\n", r.Method, r.URL.String())
		resp, err := http.DefaultTransport.RoundTrip(newRequest)
		if err != nil {
			http.Error(w, "Failed to forward request", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()
		for key, values := range resp.Header {
			for _, value := range values {
				w.Header().Add(key, value)
			}
		}
		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
	}))
}

func handleHttps(w http.ResponseWriter, r *http.Request) {
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}
	fmt.Println("Hijacker: ", hijacker)

	conn, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, "Failed to hijack connection", http.StatusInternalServerError)
		return
	}
	fmt.Println("Connection hijacked: ", conn)

	targetConn, err := net.Dial("tcp", r.Host)
	if err != nil {
		http.Error(w, "Failed to connect to target host", http.StatusInternalServerError)
		return
	}
	fmt.Println("Connected to target host: ", targetConn)

	// defer targetConn.Close()
	// defer conn.Close()
	conn.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))
	fmt.Printf("HTTP/1.1 200 Connection Established\r\n\r\n")

	go io.Copy(targetConn, conn)
	go io.Copy(conn, targetConn)
}
