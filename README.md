# mss
Music Source Separation

Master's Thesis Solution Repository - Grgur Crnogorac

Requirements:
- Go
- Python with depenedencies installed - version 3.10 was used during development

How to run

1. Clone the repository
2. Install the requirements - `pip install -r requirements.txt`
3. Install go dependencies - `go mod tidy`
3. Run the application - `go run main.go`
4. Open the application in a browser - http://localhost:8080

Notes:
- currently only tested on MacOS
- depending on OS and Python version, some troubleshooting might be necessary
- in my case, there were backwards compatibility problems with some Python packages and binaries