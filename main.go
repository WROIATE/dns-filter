package main

import (
	"flag"
	"log"
	"sync"

	"github.com/miekg/dns"
)

var UpstreamAddress = flag.String("s", "1.1.1.1:53", "upstream dns server")
var ListenAddress = flag.String("l", "0.0.0.0:5367", "listen address")

func main() {
	flag.Parse()

	dns.HandleFunc(".", HandleDNSRequest)

	server := &dns.Server{Addr: *ListenAddress, Net: "udp"}
	log.Printf("Starting DNS server on %s, upstream %s\n", server.Addr, *UpstreamAddress)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Failed to start DNS server: %v\n", err)
	}
}

func HandleDNSRequest(writer dns.ResponseWriter, req *dns.Msg) {
	c := new(dns.Client)
	c.Net = "udp"

	var upstreamDNS = *UpstreamAddress

	resp, _, err := c.Exchange(req, upstreamDNS)

	if err != nil {
		log.Printf("Query upstream DNS error: %v\n", err)
		return
	}

	defer writer.WriteMsg(resp)

	if len(resp.Answer) == 0 {
		return
	}

	recordMap := make(map[string]struct{})

	wg := sync.WaitGroup{}

	for index, question := range req.Question {
		// If query type is AAAA, try to check A record
		if question.Qtype == dns.TypeAAAA {
			wg.Add(1)
			go func(k int, question dns.Question) {
				defer wg.Done()
				m := new(dns.Msg)
				m.SetQuestion(question.Name, dns.TypeA)
				r, _, _ := c.Exchange(m, upstreamDNS)
				if r != nil && len(r.Answer) != 0 {
					// Mark domain name
					recordMap[question.Name] = struct{}{}
				}
			}(index, question)
		}
	}

	wg.Wait()

	// CNAME
	temp := resp.Answer[:0]

	for _, answer := range resp.Answer {
		if answer == nil {
			continue
		}
		if answer.Header().Rrtype == dns.TypeCNAME {
			if _, ok := recordMap[answer.Header().Name]; ok {
				// log.Printf("Block cname record: %s %s\n", answer.Header().Name, answer.(*dns.CNAME).Target)
				recordMap[answer.(*dns.CNAME).Target] = struct{}{}
				continue
			}
		}
		temp = append(temp, answer)
	}

	temp = resp.Answer[:0]

	// AAAA
	for _, answer := range resp.Answer {
		if answer == nil {
			continue
		}
		if answer.Header().Rrtype == dns.TypeAAAA {
			if _, ok := recordMap[answer.Header().Name]; ok {
				// log.Printf("Block v6 record: %s %s\n", answer.Header().Name, answer.(*dns.AAAA).AAAA.String())
				continue
			}
		}
		temp = append(temp, answer)
	}

	resp.Answer = temp
}
