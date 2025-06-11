package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
)

var (
	url       string
	list      string
	output    string
	jsCheck   bool
	deepScan  bool
	threads   int
	verbose   bool
	client    = &http.Client{Timeout: 10 * time.Second}
)

func init() {
	flag.StringVar(&url, "u", "", "Single URL to scan")
	flag.StringVar(&list, "l", "", "File containing list of URLs")
	flag.StringVar(&output, "o", "results.txt", "Output file")
	flag.BoolVar(&jsCheck, "js", false, "Enable JS file analysis")
	flag.BoolVar(&deepScan, "deep", false, "Enable deep scan")
	flag.IntVar(&threads, "concurrency", 20, "Number of threads")
	flag.BoolVar(&verbose, "verbose", false, "Show verbose output")
	flag.Parse()
}

func main() {
	displayBanner()

	if url == "" && list == "" {
		color.Red("❌ Error: No URL provided. Use -u or -l")
		flag.Usage()
		os.Exit(1)
	}

	var urls []string
	if url != "" {
		urls = append(urls, normalizeURL(url))
	} else {
		urls = readURLsFromFile(list)
	}

	color.Green("🔍 Scanning %d URLs...", len(urls))

	resultsFile, err := os.Create(output)
	if err != nil {
		color.Red("❌ Failed to create output file: %v", err)
		os.Exit(1)
	}
	defer resultsFile.Close()

	processURLs(urls, resultsFile)
	color.Green("\n✅ Scan completed! Results saved to: %s", output)
}

func displayBanner() {
	banner := `
███████╗███╗   ██╗██████╗ ██████╗ ██╗  ██╗███████╗██████╗ ██╗  ██╗███████╗██████╗ 
██╔════╝████╗  ██║██╔══██╗██╔══██╗██║ ██╔╝██╔════╝██╔══██╗██║  ██║██╔════╝██╔══██╗
█████╗  ██╔██╗ ██║██║  ██║██████╔╝█████╔╝ █████╗  ██████╔╝███████║█████╗  ██████╔╝
██╔══╝  ██║╚██╗██║██║  ██║██╔═══╝ ██╔═██╗ ██╔══╝  ██╔═══╝ ██╔══██║██╔══╝  ██╔══██╗
███████╗██║ ╚████║██████╔╝██║     ██║  ██╗███████╗██║     ██║  ██║███████╗██║  ██║
╚══════╝╚═╝  ╚═══╝╚═════╝ ╚═╝     ╚═╝  ╚═╝╚══════╝╚═╝     ╚═╝  ╚═╝╚══════╝╚═╝  ╚═╝
                                 By: amitlt2 🚀 v1.1
	`
	color.Cyan(banner)
}

func normalizeURL(rawURL string) string {
	if !strings.HasPrefix(rawURL, "http") {
		return "https://" + rawURL
	}
	return rawURL
}

func readURLsFromFile(filename string) []string {
	var urls []string
	file, err := os.Open(filename)
	if err != nil {
		color.Red("❌ Error opening file: %v", err)
		os.Exit(1)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		url := strings.TrimSpace(scanner.Text())
		if url != "" {
			urls = append(urls, normalizeURL(url))
		}
	}

	if err := scanner.Err(); err != nil {
		color.Red("❌ Error reading file: %v", err)
		os.Exit(1)
	}

	return urls
}

func processURLs(urls []string, outputFile *os.File) {
	var wg sync.WaitGroup
	urlChan := make(chan string, threads)
	resultChan := make(chan string, threads*10)

	// Start workers
	for i := 0; i < threads; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for url := range urlChan {
				scanURL(url, resultChan, outputFile)
			}
		}()
	}

	// Send URLs to workers
	go func() {
		for _, u := range urls {
			urlChan <- u
		}
		close(urlChan)
	}()

	// Print results
	go func() {
		for res := range resultChan {
			fmt.Println(res)
			outputFile.WriteString(res + "\n")
		}
	}()

	wg.Wait()
	close(resultChan)
}

func scanURL(target string, results chan<- string, outputFile *os.File) {
	color.Blue("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	color.Yellow("🔎 Target: %s", target)

	// Get initial page
	resp, err := client.Get(target)
	if err != nil {
		results <- color.RedString("❌ Failed to fetch %s: %v", target, err)
		return
	}
	defer resp.Body.Close()

	// Find JS files in page
	if jsCheck {
		findJSFiles(target, resp.Body, results)
	}

	// Extract parameters
	extractParameters(target, results)
}

func findJSFiles(baseURL string, body io.Reader, results chan<- string) {
	// Simple regex to find JS files in HTML
	jsRegex := regexp.MustCompile(`<script.*?src=["'](.*?\.js)["']`)
	content, err := io.ReadAll(body)
	if err != nil {
		results <- color.RedString("❌ Error reading response: %v", err)
		return
	}

	matches := jsRegex.FindAllStringSubmatch(string(content), -1)
	for _, match := range matches {
		if len(match) > 1 {
			jsURL := resolveRelativeURL(baseURL, match[1])
			results <- color.CyanString("📜 Found JS: %s", jsURL)
			analyzeJSFile(jsURL, results)
		}
	}
}

func analyzeJSFile(jsURL string, results chan<- string) {
	resp, err := client.Get(jsURL)
	if err != nil {
		results <- color.RedString("❌ Failed to fetch JS %s: %v", jsURL, err)
		return
	}
	defer resp.Body.Close()

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		results <- color.RedString("❌ Error reading JS: %v", err)
		return
	}

	// Simple endpoint detection in JS files
	endpointRegex := regexp.MustCompile(`["'](/[a-zA-Z0-9_\-/.]+)["']`)
	endpoints := endpointRegex.FindAllStringSubmatch(string(content), -1)

	for _, match := range endpoints {
		if len(match) > 1 {
			fullURL := resolveRelativeURL(jsURL, match[1])
			results <- color.GreenString("🔍 Found endpoint: %s", fullURL)
		}
	}
}

func extractParameters(target string, results chan<- string) {
	// Simple parameter detection
	paramRegex := regexp.MustCompile(`\?[a-zA-Z0-9_\-]+=[a-zA-Z0-9_\-]*`)
	matches := paramRegex.FindAllString(target, -1)

	for _, match := range matches {
		results <- color.YellowString("🛠️ Found parameter: %s%s", target, match)
	}
}

func resolveRelativeURL(baseURL, relativePath string) string {
	if strings.HasPrefix(relativePath, "http") {
		return relativePath
	}
	if strings.HasPrefix(relativePath, "//") {
		return "https:" + relativePath
	}
	if strings.HasPrefix(relativePath, "/") {
		u := strings.Split(baseURL, "/")
		return u[0] + "//" + u[2] + relativePath
	}
	return baseURL + "/" + relativePath
}
