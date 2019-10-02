# Parts
- HTTP - Yes
- HTTPS - 50/50
- Repeater - Yes

# Proxy usage
Config is `config/sproxy.json`

    go run sproxy.go
    
# Repeater usage
Config is `config/repeates.json`

    go run repeater.go
    
Then open `localhost:8899` with get param `?id={id of request}`. Id you can find in `proxy.sqlite`.
