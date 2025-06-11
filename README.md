# EndpointXplorer 🔍
## Advanced Hidden Endpoint Discovery Tool
### Author: amitlt2

A powerful Go-based tool to discover hidden endpoints, parameters, and URLs using OSINT techniques, Wayback Machine data, and JavaScript file analysis.

## ✨ Features
* Multi-Source URL Discovery (Wayback Machine, Common Crawl)
* JavaScript File Analysis (Extracts hidden endpoints & API routes)
* Smart Parameter Extraction (Finds ?param=, /api/, /v1/, etc.)
* Concurrent Scanning (Fast processing with configurable threads)
* Colorful CLI Output (Easy-to-read results)
* Save Results to File (JSON/TXT formats)

# 🚀 Installation
## 1. Install from Source
### Clone the repo
```
git clone https://github.com/amitlt2/EndpointXplorer
cd EndpointXplorer
```

### Build the tool
```
go build -o endpointxplorer
```

### Move to PATH (optional)

```
sudo mv endpointxplorer /usr/local/bin/
```

## 2. Install via Go

```
go install github.com/amitlt2/EndpointXplorer@latest
```

## Commands
### 1. Basic Scan (Single URL)
```
./endpointxplorer -u https://example.com -js -o results.txt
```

### 2. Bulk Scan (List of URLs)
```
./endpointxplorer -l urls.txt -js -deep -o results.json
```

### 3. Full Deep Scan
```
./endpointxplorer -u https://example.com -js -deep -concurrency 100 -verbose
```

## Flags
Flag	Description	Example
```
-u	Single URL to scan	-u "https://example.com"
-l	File containing list of URLs	-l urls.txt
-o	Output file (supports .txt/.json)	-o results.json
-js	Enable JS file analysis	-js
-deep	Deep scan (slower but thorough)	-deep
-concurrency	Threads (default: 20)	-concurrency 50
-verbose	Show verbose output	-verbose
```

## 🛡 Legal Disclaimer
This tool is for authorized security testing and research only.
Misuse for illegal activities is strictly prohibited.


