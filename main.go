
package main

import (
  "os"
  "fmt"
  "strings"
  "io/ioutil"
  "encoding/json"
  "encoding/base64"

  "github.com/fsnotify/fsnotify"
)

// Structs

type CertificateDomain struct {
  Main string `json:"Main"`
}

type Certificate struct {
  Domain CertificateDomain `json:"Domain"`
  Certificate string `json:"Certificate"`
  Key string `json:"Key"`
}

type Acme struct {
  Certificates []Certificate `json:"Certificates"`
}

// Find given name in certificate array
func findDomain(name string, certs []Certificate) bool {
  for _, cert := range certs {
    if cert.Domain.Main == name {
      return true
    }
  }
  return false
}

// Vars

var certPath, acmeFile string

// Extract and write certificates
func buildCerts() {
  before, err := ioutil.ReadDir(certPath)
  if err != nil {
    fmt.Println(err)
  }

  f, err := ioutil.ReadFile(acmeFile)
  if err != nil {
    fmt.Println("Error", err)
    return
  }

  var acme Acme
  json.Unmarshal(f, &acme)
  for _, cert := range acme.Certificates {
    // Decode
    decoded, err := base64.StdEncoding.DecodeString(cert.Certificate)
    if err != nil {
      fmt.Println("Unable to decode certificate", cert.Domain.Main)
      continue
    }

    // Write chain
    name := fmt.Sprintf("%s/%s.chain.crt", certPath, cert.Domain.Main)
    fmt.Println("Writing file", name)
    err = ioutil.WriteFile(name, decoded, 0644)
    if err != nil {
      fmt.Println("Error writing file", name)
    }

    // Write cert
    name = fmt.Sprintf("%s/%s.crt", certPath, cert.Domain.Main)
    fmt.Println("Writing file", name)
    parts := strings.Split(string(decoded), "\n\n")
    err = ioutil.WriteFile(name, []byte(parts[0]), 0644)
    if err != nil {
      fmt.Println("Error writing file", name)
    }

    // Decode key
    decoded, err = base64.StdEncoding.DecodeString(cert.Key)
    if err != nil {
      fmt.Println("Unable to decode certificate", cert.Domain.Main)
      continue
    }

    // Write key
    name = fmt.Sprintf("%s/%s.key", certPath, cert.Domain.Main)
    fmt.Println("Writing file", name)
    err = ioutil.WriteFile(name, []byte(decoded), 0644)
    if err != nil {
      fmt.Println("Error writing file", name)
    }
  }

  // Remove any no longer in use
  for _, f := range before {
    var domain string
    l := len(f.Name())
    if l >= 10 && f.Name()[l - 10:] == ".chain.crt" {
      domain = f.Name()[:l - 10]
    } else if l >= 4 && f.Name()[l - 4:] == ".crt" {
      domain = f.Name()[:l - 4]
    } else {
      continue
    }

    if !findDomain(domain, acme.Certificates) {
      fmt.Println("Removing file", f.Name())
      os.Remove(fmt.Sprintf("%s/%s", certPath, f.Name()))
    }
  }

  fmt.Println("Wrote certificates")
}

// Main
func main() {
  var exists bool
  certPath, exists = os.LookupEnv("CERT_PATH")
  if !exists {
    certPath = "/certs"
  }

  acmePath, _ := os.LookupEnv("ACME_PATH")
  acmeFile = fmt.Sprintf("%s/acme.json", acmePath)

  // Setup watcher
  watcher, err := fsnotify.NewWatcher()
  if err != nil {
    fmt.Println("Error", err)
    return
  }

  defer watcher.Close()
  done := make(chan bool)

  go func () {
    for {
      select {
      case ev := <-watcher.Events:
        if ev.Op & fsnotify.Write == fsnotify.Write {
          buildCerts()
        }
      case err := <-watcher.Errors:
        fmt.Printf("Error: %#v\n", err)
      }
    }
  }()

  // Watch acme.json
  err = watcher.Add(acmeFile)
  if err != nil {
    fmt.Println("Error", err)
  }
  fmt.Println("Watching", acmeFile)

  // One for kick off
  buildCerts()

  <-done
}
