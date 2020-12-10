package main

import (
	"flag"
	"io"
	"log"
	"net"
	"strings"
	"time"
)

type DNS struct {
	Addr string
	Time time.Time
}

var DNSCache = make(map[string]DNS, 0)

func main() {
	var listen = flag.String("listen", ":8080", "Listen address for proxy")
	flag.Parse()
	l, err := net.Listen("tcp", *listen)
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()
	for {
		conn, err := l.Accept() // Wait for a connection.
		if err != nil {
			continue
		}

		go func(c net.Conn) {
			defer c.Close()
			buf_size := 1024
			buf := make([]byte, buf_size) // simple buffer for incoming requests
			hostport := ""
			get_line := ""

			for i := 0; i < buf_size; i++ { // Read one charater at a time
				if _, err := c.Read(buf[i : i+1]); err != nil {
					break
				}
				if buf[i] == 0xa { // New line to parse...
					s := string(buf[0 : i+1])
					if strings.HasPrefix(s, "CONNECT ") {
						parts := strings.SplitN(s, " ", 3)
						hostport = parts[1]
					} else if strings.HasPrefix(s, "GET ") || strings.HasPrefix(s, "POST ") {
						parts := strings.SplitN(s, " ", 3)
						hostport = parts[1]
						if strings.HasPrefix(hostport, "http://") {
							hostport = strings.SplitN(hostport[7:], "/", 2)[0]
						}
						paths := strings.SplitN(parts[1], "/", 4)
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
				if val, ok := DNSCache[host]; ok && val.Time.After(time.Now().Add(-1*time.Minute)) {
					addr = val.Addr
					DNSCache[host] = val
				} else {
					addrs, err := net.LookupHost(host)
					if err != nil {
						return
					}
					addr = addrs[0]
					DNSCache[host] = DNS{Addr: addr, Time: time.Now()}
				}
				if remote, err := net.Dial("tcp", net.JoinHostPort(addr, port)); err == nil {
					defer remote.Close()
					if get_line == "" {
						c.Write([]byte("HTTP/1.1 200 OK\n\n"))
					} else {
						remote.Write([]byte(get_line))
					}
					go io.Copy(remote, c)
					io.Copy(c, remote)
				}
			}
		}(conn)
	}
}
