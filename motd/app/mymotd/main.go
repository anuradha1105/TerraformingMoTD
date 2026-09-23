package main

import (
"bufio"
"bytes"
"encoding/json"
"fmt"
"io"
"net/http"
"os"
"path/filepath"
"strings"
)

type config struct {
Token string
Host  string
Port  string
}

func loadConfig(path string) (config, error) {
var cfg config

f, err := os.Open(path)
if err != nil {
return cfg, err
}
defer f.Close()

scanner := bufio.NewScanner(f)
for scanner.Scan() {
line := strings.TrimSpace(scanner.Text())
if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
continue
}

parts := strings.SplitN(line, "=", 2)
if len(parts) != 2 {
continue
}

key := strings.TrimSpace(parts[0])
value := strings.TrimSpace(parts[1])

switch key {
case "token":
cfg.Token = value
case "host":
cfg.Host = value
case "port":
cfg.Port = value
}
}

if err := scanner.Err(); err != nil {
return cfg, err
}

return cfg, nil
}

func configPath() (string, error) {
home, err := os.UserHomeDir()
if err != nil {
return "", err
}
return filepath.Join(home, ".config", "motd.ini"), nil
}

func main() {
if len(os.Args) != 2 {
fmt.Fprintf(os.Stderr, "usage: %s \"new message\"\n", os.Args[0])
os.Exit(1)
}
message := os.Args[1]

path, err := configPath()
if err != nil {
fmt.Fprintf(os.Stderr, "could not determine home directory: %v\n", err)
os.Exit(1)
}

cfg, err := loadConfig(path)
if err != nil {
fmt.Fprintf(os.Stderr, "could not read %s: %v\n", path, err)
os.Exit(1)
}

if cfg.Token == "" || cfg.Host == "" || cfg.Port == "" {
fmt.Fprintf(os.Stderr, "%s must set token, host, and port\n", path)
os.Exit(1)
}

body, err := json.Marshal(map[string]string{
"token":   cfg.Token,
"message": message,
})
if err != nil {
fmt.Fprintf(os.Stderr, "could not encode request: %v\n", err)
os.Exit(1)
}

url := fmt.Sprintf("http://%s:%s/message", cfg.Host, cfg.Port)

resp, err := http.Post(url, "application/json", bytes.NewReader(body))
if err != nil {
fmt.Fprintf(os.Stderr, "request failed: %v\n", err)
os.Exit(1)
}
defer resp.Body.Close()

respBody, _ := io.ReadAll(resp.Body)

if resp.StatusCode != http.StatusOK {
fmt.Fprintf(os.Stderr, "server returned %s: %s\n", resp.Status, strings.TrimSpace(string(respBody)))
os.Exit(1)
}

fmt.Println("message updated")
}
