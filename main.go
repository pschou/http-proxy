package main

import (
	"crypto/rand"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

type DNS struct {
	Addr string
	Time time.Time
}

var DNSCache = make(map[string]DNS, 0)
var DNSCacheLock sync.Mutex
var debug = false

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Simple HTTP-Proxy, written by Paul Schou github@paulschou.com in December 2020\nAll rights reserved, personal use only, provided AS-IS -- not responsible for loss.\nUsage implies agreement.  For requests or support, please contact above.\n\n Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}
	var listen = flag.String("listen", ":8080", "Listen address for proxy")
	var cert = flag.String("cert", "/etc/pki/server.pem", "File to load with CERT when TLS is enabled")
	var key = flag.String("key", "/etc/pki/server.pem", "File to load with KEY when TLS is enabled")
	var tls_enabled = flag.Bool("tls", false, "Enable TLS on listening port (default -tls=false)")
	var Debug = flag.Bool("debug", false, "Turn on debug logging (default -debug=false)")
	flag.Parse()
	debug = *Debug

	var l net.Listener
	if *tls_enabled {
		cert, err := tls.LoadX509KeyPair(*cert, *key)
		if err != nil {
			log.Fatalf("server: loadkeys: %s", err)
		}
		config := tls.Config{Certificates: []tls.Certificate{cert}, InsecureSkipVerify: true}
		config.Rand = rand.Reader
		if l, err = tls.Listen("tcp", *listen, &config); err != nil {
			log.Fatal(err)
		}
	} else {
		var err error
		if l, err = net.Listen("tcp", *listen); err != nil {
			log.Fatal(err)
		}
	}

	defer l.Close()
	for {
		conn, err := l.Accept() // Wait for a connection.
		if err != nil {
			continue
		}

		go func(c net.Conn) {
			if debug {
				log.Println("new connection", c)
			}
			defer c.Close()
			buf_size := 3 * 1024 * 1024
			buf := make([]byte, buf_size) // simple buffer for incoming requests
			hostport := ""
			get_line := ""

			for i := 0; i < buf_size-1; i++ { // Read one charater at a time
				if _, err := c.Read(buf[i : i+1]); err != nil {
					break
				}
				if buf[i] == 0xa { // New line to parse...
					s := string(buf[0 : i+1])
					if strings.HasPrefix(s, "CONNECT ") {
						parts := strings.SplitN(s, " ", 3)
						if len(parts) < 2 {
							break
						}
						hostport = parts[1]
						if debug {
							log.Println("CONNECT request", hostport)
						}
					} else if strings.HasPrefix(s, "GET ") || strings.HasPrefix(s, "POST ") ||
						strings.HasPrefix(s, "HEAD ") || strings.HasPrefix(s, "OPTIONS ") {
						parts := strings.SplitN(s, " ", 3)
						if len(parts) < 3 || len(parts[1]) < 8 {
							break
						}
						hostport = parts[1]
						if strings.HasPrefix(hostport, "http://") {
							hostport = strings.SplitN(hostport[7:], "/", 2)[0]
						}
						paths := strings.SplitN(parts[1], "/", 4)
						if len(parts) < 4 {
							hostport = ""
							break
						}
						get_line = parts[0] + " /" + paths[3] + " " + parts[2]
						break
					} else if i <= 1 { // end of connect request!
						break
					}

					i = -1 // reset the buffer scanner to 0
				}
			}

			if hostport != "" { // if any request was passed, parse it
				host, port, err := net.SplitHostPort(hostport)
				addr := ""
				if err != nil {
					host = hostport
					port = "80"
				}
				// use a cache for all requests under a minute
				if val, ok := DNSCache[host]; ok && val.Time.After(time.Now().Add(-1*time.Minute)) {
					addr = val.Addr
					DNSCacheLock.Lock()
					DNSCache[host] = val
					DNSCacheLock.Unlock()
				} else {
					addrs, err := net.LookupHost(host)
					if err != nil {
						return
					}
					addr = addrs[0]
					DNSCacheLock.Lock()
					DNSCache[host] = DNS{Addr: addr, Time: time.Now()}
					DNSCacheLock.Unlock()
				}
				con_str := net.JoinHostPort(addr, port)
				if debug {
					log.Println("  Dialing", con_str)
				}
				if remote, err := net.Dial("tcp", con_str); err == nil {
					defer remote.Close()
					if get_line == "" {
						c.Write([]byte("HTTP/1.1 200 OK\n\n"))
					} else {
						remote.Write([]byte(get_line))
					}
					if debug {
						log.Println("  Read/Write mode for", con_str)
					}
					go io.Copy(remote, c)
					io.Copy(c, remote)
					if debug {
						log.Println("  closed", con_str)
					}
				}
			}
		}(conn)
	}
}
