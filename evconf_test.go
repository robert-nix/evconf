package evconf

import (
  "github.com/Mischanix/applog"
  "os"
  "testing"
  "time"
)

var config struct {
  StringKey string `json:"string_key"`
}

// This isn't the most deterministic of tests because of I/O latency, but it's
// a test nonetheless.
func TestEvconf(t *testing.T) {
  // Log to stdout
  applog.SetOutput(os.Stdout)
  applog.Level = applog.DebugLevel
  // Create config_test.json
  f, _ := os.Create("config_test.json")
  f.WriteString("{\n  \"string_key\": \"I'm Cool!\"\n}\n")
  f.Close()

  c := New("config_test.json", &config)
  numCalls := 0
  c.OnLoad(func() {
    numCalls++
    if numCalls == 1 && config.StringKey != "I'm Cool!" {
      t.Errorf(
        "OnLoad#1 expected \"I'm Cool!\", got %v",
        config.StringKey,
      )
    }
    if numCalls == 2 && config.StringKey != "I'm Cooler!" {
      t.Errorf(
        "OnLoad#2 expected \"I'm Cooler!\", got %v",
        config.StringKey,
      )
    }
    if numCalls == 3 && config.StringKey != "I'm Cooler!" {
      t.Errorf(
        "OnLoad#3 expected \"I'm Cooler!\", got %v",
        config.StringKey,
      )
    }
  })
  c.Ready()

  <-time.After(1 * time.Millisecond)

  // Update config_test.json
  f, _ = os.Create("_config_test.json")
  f.WriteString("{\n  \"string_key\": \"I'm Cooler!\"\n}\n")
  f.Close()
  os.Remove("config_test.json")
  os.Rename("_config_test.json", "config_test.json")

  <-time.After(1 * time.Millisecond)

  // Update config_test.json with invalid data -- don't overwrite valid configs
  f, _ = os.Create("_config_test.json")
  f.WriteString("{\n  \"not_string_key\": \"I'm Coolest!\"\n}\n")
  f.Close()
  os.Remove("config_test.json")
  os.Rename("_config_test.json", "config_test.json")

  <-time.After(1 * time.Millisecond)
  os.Remove("config_test.json")
  if numCalls != 3 {
    t.Errorf(
      "Expected 3 calls for OnLoad, but got %d",
      numCalls,
    )
  }
}
