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
