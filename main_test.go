
package main

import (
  "os"
  // "fmt"
  "testing"
  "io/ioutil"
)

func setup() {
  certPath = "test/output"
  acmeFile = "test/acme.json"
  os.RemoveAll(certPath)
  os.Mkdir(certPath, 0755)
}

func TestBuildCerts(t *testing.T) {
  setup()
  buildCerts()

  // Check cert
  output, err := ioutil.ReadFile("test/output/example.com.crt")
  if err != nil {
    t.Error("Error reading cert", err)
  }
  expected, _ := ioutil.ReadFile("test/example.crt")
  if string(expected) != string(output) {
    t.Error("Generated certificate does not match expected output")
  }

  // Check chain
  output, err = ioutil.ReadFile("test/output/example.com.chain.crt")
  if err != nil {
    t.Error("Error reading cert", err)
  }
  expected, _ = ioutil.ReadFile("test/example.chain.crt")
  if string(expected) != string(output) {
    t.Error("Generated chain does not match expected output")
  }

  // Check key
  output, err = ioutil.ReadFile("test/output/example.com.key")
  if err != nil {
    t.Error("Error reading key", err)
  }
  expected, _ = ioutil.ReadFile("test/example.key")
  if string(expected) != string(output) {
    t.Error("Generated key does not match expected output")
  }
  t.Log(len(output))
  t.Log(len(expected))
}

func TestCleanup(t *testing.T) {
  setup()

  // Add superflous files
  ioutil.WriteFile("test/output/remove.crt", []byte(""), 0600)
  ioutil.WriteFile("test/output/remove.chain.crt", []byte(""), 0600)
  ioutil.WriteFile("test/output/leave", []byte(""), 0600)

  // Run
  buildCerts()

  // Check
  _, err := os.Stat("test/output/remove.crt")
  if err == nil {
    t.Error("Expected test/output/remove.crt to be removed")
  }

  _, err = os.Stat("test/output/remove.chain.crt")
  if err == nil {
    t.Error("Expected test/output/remove.chain.crt to be removed")
  }

  _, err = os.Stat("test/output/leave")
  if os.IsNotExist(err) {
    t.Error("Expected test/output/remove.crt not to be removed")
  } else if err != nil {
    t.Error(err)
  }
}
