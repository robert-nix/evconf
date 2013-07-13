// Package evconf provides a simple JSON-based configuration system with
// automatic runtime updating of config data on config file changes.
package evconf

import (
  "encoding/json"
  "github.com/Mischanix/applog"
  "github.com/howeyc/fsnotify"
  "os"
  "path/filepath"
  "sync"
)

type Config struct {
  path    string
  once    *sync.Once
  onload  func()
  watcher *fsnotify.Watcher
  data    interface{}
}

// New returns a new Config object.  Once a client has bound callbacks via
// OnLoad, it may call Ready.
func New(path string, data interface{}) (c *Config) {
  w, err := fsnotify.NewWatcher()
  if err != nil {
    applog.Error("evconf.InitConfig: NewWatcher failed: %v", err)
  }
  c = &Config{path, &sync.Once{}, nil, w, data}

  return c
}

// Only one OnLoad func may be bound per Config object.
func (c *Config) OnLoad(fn func()) {
  c.onload = fn
}

// Ready will only be called once in the lifetime of a Config object.
func (c *Config) Ready() {
  c.once.Do(func() {
    // Always call loadConfig from a separate goroutine for consistency
    go func() {
      c.loadConfig()
    }()

    go func() {
      for c.watcher != nil {
        select {
        case ev := <-c.watcher.Event:
          if filepath.Base(ev.Name) == filepath.Base(c.path) && !ev.IsDelete() {
            // Synchronous file operation to block up channels while we use it.
            c.loadConfig()
          }
        case err := <-c.watcher.Error:
          applog.Error("evconf.watcher: watcher.Error: %v", err)
        }
      }
    }()

    err := c.watcher.Watch(filepath.Dir(c.path))
    if err != nil {
      applog.Error("evconf.ready: Watch failed: %v", err)
    }
  })
}

// StopWatching stops watching the config file for changes.  No further load
// events will be emitted except by the ready event.
func (c *Config) StopWatching() {
  c.watcher = nil
}

// loadConfig loads the config from path and emits a load event.
func (c *Config) loadConfig() {
  file, err := os.Open(c.path)
  if err != nil {
    applog.Error("evconf.loadConfig: Open failed: %v", err)
    return
  }

  d := json.NewDecoder(file)
  if err = d.Decode(c.data); err != nil {
    applog.Error("evconf.loadConfig: json Decode failed: %v", err)
    return
  }

  file.Close()

  if c.onload != nil {
    c.onload()
  }
}
