package main

import (
	"context"
	"encoding/binary"
	"flag"
	"fmt"
	"hash/fnv"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/Snare295/architect-lab4/httptools"
	"github.com/Snare295/architect-lab4/signal"
	"golang.org/x/exp/slices"
)

var (
	port       = flag.Int("port", 8090, "load balancer port")
	timeoutSec = flag.Int("timeout-sec", 3, "request timeout time in seconds")
	https      = flag.Bool("https", false, "whether backends support HTTPs")

	traceEnabled = flag.Bool("trace", true, "whether to include tracing information into responses")
)

var (
	timeout     = time.Duration(*timeoutSec) * time.Second
	serversPool = []string{
		"server1:8080",
		"server2:8080",
		"server3:8080",
	}
)

func scheme() string {
	if *https {
		return "https"
	}
	return "http"
}

func health(dst string) bool {
	ctx, _ := context.WithTimeout(context.Background(), timeout)
	req, _ := http.NewRequestWithContext(ctx, "GET",
		fmt.Sprintf("%s://%s/health", scheme(), dst), nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false
	}
	if resp.StatusCode != http.StatusOK {
		return false
	}
	return true
}

func forward(dst string, rw http.ResponseWriter, r *http.Request) error {
	ctx, _ := context.WithTimeout(r.Context(), timeout)
	fwdRequest := r.Clone(ctx)
	fwdRequest.RequestURI = ""
	fwdRequest.URL.Host = dst
	fwdRequest.URL.Scheme = scheme()
	fwdRequest.Host = dst

	resp, err := http.DefaultClient.Do(fwdRequest)
	if err == nil {
		for k, values := range resp.Header {
			for _, value := range values {
				rw.Header().Add(k, value)
			}
		}
		if *traceEnabled {
			rw.Header().Set("lb-from", dst)
		}
		log.Println("fwd", resp.StatusCode, resp.Request.URL)
		rw.WriteHeader(resp.StatusCode)
		defer resp.Body.Close()
		_, err := io.Copy(rw, resp.Body)
		if err != nil {
			log.Printf("Failed to write response: %s", err)
		}
		return nil
	} else {
		log.Printf("Failed to get response from %s: %s", dst, err)
		rw.WriteHeader(http.StatusServiceUnavailable)
		return err
	}
}

type ServerLoad struct {
	serverName string
	conHash    []uint16
}

func hashing(s string) uint16 {
	hasher := fnv.New128()
	hasher.Write([]byte(s))
	var b []byte
	hash := binary.BigEndian.Uint16(hasher.Sum(b))
	r := uint16(hash)
	fmt.Println("Value of ", s, ",Hash of new user is:", hash)
	return r
}

type BalancerRouter struct {
	curServer   int
	servers     []ServerLoad
	liveServers []*ServerLoad
}

func (b *BalancerRouter) createServersInsts(sList []string) {
	b.servers = nil
	for _, element := range sList {
		newServ := ServerLoad{serverName: element}
		b.servers = append(b.servers, newServ)
	}
}

func (b *BalancerRouter) findServerByUrl(s string) int {
	hash := hashing(s)
	serverIndex := int(hash) % len(b.servers)
	fmt.Println("findSereve by url", serverIndex)
	serv := &b.servers[serverIndex]

	for _, element := range serv.conHash {
		if element == hash {
			return serverIndex
		}
	}

	//loop if first loop can't find in server of hesh module
	for i := 0; i < len(b.servers); i++ {
		if i == serverIndex {
			continue
		}

		server := b.servers[i]
		for _, element := range server.conHash {
			if element == hash {
				return i
			}
		}
	}

	return -1
}

func (b *BalancerRouter) putClientToServ(s string) error {
	hash := hashing(s)
	serverIndex := int(hash) % len(b.servers)

	server := &b.servers[serverIndex]
	for _, element := range b.liveServers {
		if element.serverName == server.serverName {
			server.conHash = append(server.conHash, hash)
			b.curServer = serverIndex
			return nil
		}
	}

	return nil
}

func (b *BalancerRouter) addNewLiveServer(name string, working bool) {
	for _, element := range b.servers {
		if element.serverName == name {
			pw := &element

			if working {
				for _, element := range b.liveServers {
					// println("pw", pw)
					// println("element", element)
					if pw.serverName == element.serverName {
						return
					}
				}
				b.liveServers = append(b.liveServers, pw)
				println("Add instane of server from live servers of", pw.serverName)
				return
			}

			if !working {
				for i := 0; i < len(b.liveServers); i++ {
					serv := b.liveServers[i]
					if pw.serverName == serv.serverName {
						b.liveServers = slices.Delete(b.liveServers, i, i)
						println("Delted instane of server from live servers of", pw.serverName)
						return
					}
				}
			}

		}
	}

}

func main() {
	flag.Parse()
	b := BalancerRouter{}
	b.createServersInsts(serversPool)

	// TODO: Використовуйте дані про стан сервреа, щоб підтримувати список тих серверів, яким можна відправляти ззапит.
	for _, server := range serversPool {
		server := server
		go func() {
			for range time.Tick(10 * time.Second) {
				name, health := server, health(server)
				log.Println(name, health)
				b.addNewLiveServer(server, health)
			}
		}()
	}

	frontend := httptools.CreateServer(*port, http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		// TODO: Рееалізуйте свій алгоритм балансувальника.
		b.findServerByUrl(r.URL.Path)
		forward(serversPool[0], rw, r)
	}))

	log.Println("Starting load balancer...")
	log.Printf("Tracing support enabled: %t", *traceEnabled)
	frontend.Start()
	signal.WaitForTerminationSignal()

}
