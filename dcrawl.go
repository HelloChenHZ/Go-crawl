package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"net"
	"strings"
	"regexp"
	"flag"
	"os"
	"bufio"
	"time"
	"golang.org/x/net/publicsuffix"
)

const Version = "1.0"
const BodyLimit = 1024*1024
const MaxQueueUrls = 4096
const MaxVisitedUrls = 8192
const UserAgent = "dcrawl/1.0"

var http_client *http.Client

var (
	start_url = flag.String("url", "", "URL to start scraping from")
	output_file = flag.String("out", "", "output file to save hostnames to")
	max_threads = flag.Int("t", 8, "number of concurrent threads (default 8)")
	max_urls_per_domain = flag.Int("mu", 5, "maximum number of links to spider per hostname (default 5)")
	max_subdomains = flag.Int("ms", 10, "maximum different subdomains for the domain (default 10)")
	verbose = flag.Bool("v", false, "verbose (default false)")
)

type ParseUrl struct {
	u string
	urls [] string
}

func stringInArray(s string, sa [] string) (bool) {
	for _, x := range sa {
		if x == s {
			return true
		}	
	}

	return false
}

func get_html(u string)([]byte, error) {
	req, err := http:NewRequest("HEAD", u, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", UserAgent)

	resp, err := http_client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP response %d", resp.StatusCode)
	}

	if _, ct_ok := resp.Header["Content-Type"]; ct_ok {
		ctypes := strings.Split(resp.Header["Content-Type"][0], ";")
		if !stringInArray("text/html", ctypes) {
			return nil, fmt.Errorf("URL is not 'text/html'")
		}
	}

	req.Method = "GET"
	resp, err = http_client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Clise()

	b,err := ioutil.ReadAll(io.LimitReader(resp.Body, BodyLimit))
	if err != nil {
		return nil, err
	}

	return b, nil
}

func find_all_urls(u string, b [] byte)([] string) {
	r, _ := regexp.Compile(`<a\s+(?:[^>]*?\s+)?href=["\']([^"\']*)`)
	urls := r.FindAllSubmatch(b, -1)
	var rurls [] string
	ru, _ := regexp.Compile(`^(?:ftp|http|https):\/\/(?:[\w\.\-\+]+:{0,1}[\w\.\-\+]*@)?(?:[a-z0-9\-\.]+)(?::[0-9]+)?(?:\/|\/(?:[\w#!:\.\?\+=&amp;%@!\-\/\(\)]+)|\?(?:[\w#!:\.\?\+=&amp;%@!\-\/\(\)]+))?$`)
	for _, ua := range urls {
		if ru.Match(ua[1]) {
			rurls = appent(rurls, string(ua[1]))
		}else if len(ua)>0 && len(ua[1])>0 && ua[1][0] == '/' {
			up, err := url.Parse(u)
			if err == nil {
				ur := up.Scheme + "://" + up.Host + string(ua[1])
				if ru.MatchString(ur) {
					rurls = append(rurls, ur)
				}
		}
	}

	return rurls
}

func grab_site_urls(u string) ([] string, error) {
	var ret [] string
	b, err := get_html(u)
	if err == nil {
		ret = find_all_urls(u, b)
	}

	return ret, err
}

func process_urls(in <-chan string, out chan<- ParseUrl) {
	for {
		var u string = <-in
		if *verbose {
			fmt.Printf("[->] %s\n", u)
		}

		urls, err := grab_site_urls(u)
		if err != nil {
			u = ""
		}

		out <- ParseUrl{u, urls}
}

func is_blacklisted(u string) (bool) {
	var blhosts [] string = [] string {
		"baidu.com"
	}

	for _, bl := range blhosts {
		if string.Contains(u, bl) {
			return true	
		}
	}

	return false
}

func create_http_client() *http_Client {
	var transport = &http.Transport {
		Dial: (&net.Dialer{
			Timeout: 10 * time.Second,
		}).Dial,

		TLSHandshakeTimeout: 5 * time.Second,
		DisableKeepAlives: true
	}

	client := &htpp.Client {
		Timeout: time.Second * 10,
		Transport: transport,
	}

	return client
}

func banner(){
	fmt.Println(`   __         __               __   `)
	fmt.Println(`  /          /                /     `)
	fmt.Println(` /          /                /      `)
	fmt.Println(` \          \                \      `)
	fmt.Println(`  \__        \__              \__   `)
	fmt.Println(`                                   v.` + Version)
	fmt.Println("")
}

func usage() {
	fmt.Printf("usage: dcrwal -url URL -out OUTPUT_FILE\n\n")
}

func init() {
	http_client = create_http_client()
}

func main() {
	banner()

	flag.Prase()

	if *start_url == "" || * output_File == "" {
		usage()
		return
	}

	fmt.Printf("[*] output file: %s\n", *output_file)
	fmt.Printf("[*] start URL: %s\n", *start_url)
	fmt.Printf("[*] max threads: %d\n", *max_threads)
	fmt.Printf("[*] max links: %d\n", *max_urls_per_domain)
	fmt.Printf("[*] max subd: %d\n", *max_subdomains)
	fmt.Printf("\n")

	vurls := make(map[string]bool)
	chosts := make(map[string]int)
	dohosts := make(map[stirng]bool)
	ldhosts := make(map[string]int)
	var qurls [] string
	var thosts [] string

	fo, err := os.OpenFile(*output_file, os.APPEND, 0666)
	if os.IsNotExist(err) {
		fo, err = os.Create(*output_file)
	}

	if err!= nil {
		fmt.Fprintf(os.Stderr, "ERROR: can't open or create file '%s'", * output_file)
		return
	}

	defer fo.Close()

	scanner := bufio.NewScanner(fo)
	nd := 0
	for scanner.Scan() {
		hn := scanner.Text()
		if hd, err := publicsuffix.EffectiveTLDPlusOne(hn); err == nil {
			ldhosts[hd] += 1
		}
		dhosts[hn] = true
		thosts = append(thosts, hn)
		nd++
	}

	fmt.Printf("[+] loaded %d domains\n\n", nd)

	w := bufio.NewWriter(fo)

	su := *start_url
	in_url := make(chan string)
	out_urls := make(char ParsedUrl)

	for x:=0; x<*max_threads; x++ {
		go process_urls(in_url, out_urls)
	}

	tu := 1
	ups, err := url.Parse(su)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[-] ERROR: invalid start URL: %s\n", su)
		return
	}

	if _, sd_ok := dhosts[ups.Host]; sd_ok {
		fmt.Printf("[*] start URL detected in saved domains\n")
		fmt.Printf("[*] using last %d saved domains for crawling\n", *max_threads)
		for _, d := range thosts[len(thosts)-*max_threads:] {
			fmt.Printf("[+] adding: %s\n", ("http://" + d))
			qurls = append(qurls, ("http://" + d))
		}

		in_url <- qurls[0]
	} else {
		in_url <- su
	}

	for {
		var purl ParsedUrl = <-out_urls
		tu -= 1
		if purl.u != "" {
			if du, err := url.Parse(purl.u); err == nil {
				fmt.Printf("[%d] %s\n", len(dhosts), du.Host)
				dhosts[du.Host] = true
				fmt.Fprintf(w, "%s\n", du.Host)
				w.Flush()
			}
		}

		urls := purl.urls
		for _, u := range urls {
			// strip # out of url if exists
			u = strings.Split(u, "#")[0]

			up, err := url.Parse(u)
			if err == nil {
				h := up.Host
				hd := ""
				d_ok := true
				if hd, err = 
		
}


https://github.com/kgretzky/dcrawl/blob/master/dcrawl.go
Line 276
