package main

import (
	"fmt"
	"flag"
	"net/http"
	"io/ioutil"
	"strings"
	"sync"
	"time"
	"net/url"
	"os"
	"regexp"
)

// Wacky shit dont ask why
var (
	prlist    []string
	prlistMu  sync.Mutex

	good      int
	bad       int
	total     int
	goodls    []string
)

// Go routines are very fucking confusing
func makeRequestWithProxy(testurl string, proxy string, resultChan chan string, timeout float64) {
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(
				&url.URL{
					Scheme: "http",
					Host: proxy,
				},
			),
		},
		Timeout: time.Duration(timeout) * time.Second,
	}


	resp, err := client.Get(testurl)
	if err != nil {
		bad++
		total++
	 	fmt.Printf("[ERROR] [T/G/B] - [%d / %d / %d] Failed to connect to %s with proxy %s: %s\n", total, good, bad, testurl, proxy, err)
		return
	}
	defer resp.Body.Close()
	total++

	if resp.StatusCode == 200 {
		good++
		resultChan <- fmt.Sprintf("[+] %s: Success | [T/G/B] - %d / %d / %d", proxy, total, good, bad)
		goodls = append(goodls, proxy)
	} else {
		bad++
		resultChan <- fmt.Sprintf("[-] %s: Failed", proxy)
	}

}

// Scuffed shit, thank you chatgpt
func makeRequest(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func fetchProxies() []string {
	rawproxies := []string{
		"https://raw.githubusercontent.com/clarketm/proxy-list/master/proxy-list-raw.txt",
		"https://spys.me/proxy.txt",
		"https://free-proxy-list.net",
		"https://www.us-proxy.org/",
		"https://www.sslproxies.org",
		"https://hidemy.name/en/proxy-list/",
	}


	ipexp := regexp.MustCompile("^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?):[0-9]+$")

	var wg sync.WaitGroup
	wg.Add(len(rawproxies))

	for _, url := range rawproxies {
		go func(url string) {
			defer wg.Done()

			body, err := makeRequest(url)
			if err != nil {
				return
			}

			lines := strings.Split(body, "\n")
			for _, line := range lines {
				if ipexp.MatchString(line) {
					prlistMu.Lock()
					prlist = append(prlist, line)
					prlistMu.Unlock()
				}
			}
		}(url)
	}

	wg.Wait()


	fmt.Printf("[:] Finished scraping proxies\n")
	fmt.Printf("[:] Total proxies: %d\n", len(prlist))

	time.Sleep(3 * time.Second)

	return prlist
}

func writeProxies(proxies []string, filen string) error {
	err := ioutil.WriteFile(filen, []byte(strings.Join(proxies, "\n")), 0644)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	
	// Flags
	customSite := flag.String("url", "https://example.com", "Custom site URL")
	customTimeout := flag.Float64("timeout", 15, "Custom request timeout")
	outFile := flag.String("out", "good.txt", "Custom output file")
	saveAll := flag.Bool("saveall", false, "Save scraped proxies to file")
	saveAllFile := flag.String("sa-file", "scraped.txt", "Output file for unchecked proxies")
	help := flag.Bool("help", false, "Display the help menu")
	flag.Parse()
	
	// Help Menu
	if *help {
		fmt.Println("Usage: ./proximal [options]")
		fmt.Println("Options:")
		flag.PrintDefaults()
		os.Exit(0)
	}
	
	fmt.Printf("[I] Check url: %s\n", *customSite)
	fmt.Printf("[I] Timeout: %f\n", *customTimeout)
	fmt.Printf("[I] Output file: %s\n", *outFile)
	fmt.Printf("[I] Save scraped: %t\n", *saveAll)
	if *saveAll {
		fmt.Printf("[I] Scraped file: %s\n", *customSite)
	} else {
		fmt.Printf("[I] Scraped file: NONE\n")
	}

	start := time.Now()
	proxies := fetchProxies()
	if *saveAll {
		writeProxies(proxies, *saveAllFile)
	}
	resultChan := make(chan string)
	var wg sync.WaitGroup

	for _, proxy := range proxies {
        wg.Add(1)
        go func(proxy string) {
			defer wg.Done()
			makeRequestWithProxy(*customSite, proxy, resultChan, *customTimeout)
		}(proxy)
    }

	go func() {
		for i := 0; i < len(proxies); i++ {
			result := <-resultChan
			fmt.Println(result)
		}
		close(resultChan)
	}()

	wg.Wait()

	writeProxies(goodls, *outFile)

	elapsed := time.Since(start)
	fmt.Printf("[G] Checking completed | Elapsed time: %s\n", elapsed)

	os.Exit(0)
}
