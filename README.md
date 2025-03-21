# Fetch SRE Health Check

CLI tool written in Go to monitor HTTP endpoint availability over time, grouped by domain. Endpoints are defined via a YAML config file. The tool logs cumulative availability percentages every 15 seconds, emulating a basic synthetic monitoring system.

---

## To Get Started

### Prerequisites

- Go 1.17+ installed  
  [Install Go](https://go.dev/doc/install)

### Clone and Run

```bash
git clone https://github.com/your-username/fetch-health-check.git
cd fetch-health-check
go mod tidy
go run main.go config.yaml