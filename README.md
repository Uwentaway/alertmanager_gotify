# Alertmanager to Gotify webhook

An [Alertmanager](https://prometheus.io/docs/alerting/latest/alertmanager/)-aware webhook endpoint that converts messages to [Gotify](https://gotify.net/) API calls.

## Installation

### Binaries

Download the already existing [binaries](https://github.com/Uwentaway/alertmanager_gotify/releases) for your platform:

```bash
export GOTIFY_URL="http://<gotify_address>:<port>/message"
export GOTIFY_TOKEN="<gotify_token>"


./alertmanager_gotify 
```

### From source

Using the standard `go install` (you must have [Go](https://golang.org/) already installed in your local machine):

```bash
go install github.com/DRuggeri/alertmanager_gotify
./alertmanager_gotify
```

Alertmanager YAML Example:
```YAML
receivers:
- name: storage
  webhook_configs:
  - url: http://127.0.0.1:9110/webhook
    send_resolved: true
```

## Docker
updating...
