---
layout: default
title: Hot Reload — confkit
---

# Hot Reload

confkit supports file watching and hot reloading. When a config file changes, your application can reload it without restarting.

## Basic Hot Reload

Use `LoadWithWatcher()` to set up file watching:

```go
import (
    "log"
    "time"
    "github.com/MimoJanra/confkit"
)

type Config struct {
    Port int    `env:"PORT" default:"8080"`
    Host string `env:"HOST" default:"localhost"`
}

func main() {
    cfg, watcher, err := confkit.LoadWithWatcher[Config]("config.yaml",
        confkit.FromYAML("config.yaml"),
        confkit.FromEnv(),
    )
    if err != nil {
        log.Fatal(err)
    }

    // Listen for config changes
    watcher.AddListener(func(oldCfg, newCfg any, err error) {
        if err != nil {
            log.Printf("Config reload failed: %v", err)
            return
        }
        log.Println("Config reloaded successfully")
        cfg = newCfg.(Config)
    })

    // Start watching
    watcher.Start()
    defer watcher.Stop()

    // ... rest of your application
}
```

## Listener Function

The listener receives:

- `oldCfg` — Previous config (before reload)
- `newCfg` — New config (after reload)
- `err` — Error if reload failed, `nil` on success

Both configs are cast to `any`, so you must type-assert them:

```go
watcher.AddListener(func(oldCfg, newCfg any, err error) {
    if err != nil {
        log.Printf("Reload error: %v", err)
        return
    }
    
    old := oldCfg.(Config)
    new := newCfg.(Config)
    
    log.Printf("Port changed from %d to %d", old.Port, new.Port)
})
```

## Multiple Listeners

Add multiple listeners — all are called on reload:

```go
watcher.AddListener(func(old, new any, err error) {
    if err != nil {
        log.Printf("Error: %v", err)
        return
    }
    log.Println("Listener 1: Config changed")
})

watcher.AddListener(func(old, new any, err error) {
    if err != nil {
        return
    }
    log.Println("Listener 2: Reloading services...")
    // Update services with new config
})

watcher.Start()
```

All listeners are called sequentially.

## Poll Interval

Control how often the file is checked for changes:

```go
watcher.SetPollInterval(2 * time.Second)  // Check every 2 seconds
watcher.Start()
```

Default: 500ms

Shorter intervals detect changes faster but use more CPU. Longer intervals save CPU but delay reload.

## Graceful Reload Failure

If reload fails (invalid config, validation error), the old config is kept:

```go
type Config struct {
    Port int `validate:"min=1,max=65535"`
}

// config.yaml initially has Port: 8080 (valid)
cfg, watcher, _ := confkit.LoadWithWatcher[Config]("config.yaml", ...)

watcher.AddListener(func(old, new any, err error) {
    if err != nil {
        log.Printf("Reload failed, keeping old config: %v", err)
        return  // config stays at Port: 8080
    }
    log.Println("Config updated successfully")
})

watcher.Start()

// Later, someone edits config.yaml: Port: 99999 (invalid)
// Listener receives error, old config is preserved
```

## Watching Multiple Files

You can only watch one file at a time with `LoadWithWatcher()`, but you can combine multiple sources:

```go
cfg, watcher, err := confkit.LoadWithWatcher[Config]("config.yaml",
    confkit.FromYAML("config.yaml"),     // watched
    confkit.FromYAML("defaults.yaml"),   // not watched (static)
    confkit.FromEnv(),                   // not watched (env)
)
```

Only changes to `config.yaml` trigger reloads. `defaults.yaml` and env vars are re-evaluated each time.

## Reload with Defaults and Validation

Reloads re-apply defaults and validation:

```go
type Config struct {
    Port int `env:"PORT" default:"8080" validate:"min=1,max=65535"`
}

cfg, watcher, _ := confkit.LoadWithWatcher[Config]("config.yaml", ...)

watcher.AddListener(func(old, new any, err error) {
    if err != nil {
        // Could be:
        // - Parse error (invalid YAML)
        // - Validation error (Port out of range)
        // - IO error (file deleted)
        log.Printf("Reload failed: %v", err)
        return
    }
    log.Println("Config reloaded and validated")
})
```

## Safe Config Access

Use a mutex to avoid race conditions:

```go
import "sync"

type Server struct {
    mu  sync.RWMutex
    cfg Config
}

func (s *Server) GetConfig() Config {
    s.mu.RLock()
    defer s.mu.RUnlock()
    return s.cfg
}

func (s *Server) SetConfig(cfg Config) {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.cfg = cfg
}

func (s *Server) SetupWatcher(watcher *confkit.ConfigWatcher) {
    watcher.AddListener(func(old, new any, err error) {
        if err != nil {
            log.Printf("Reload failed: %v", err)
            return
        }
        s.SetConfig(new.(Config))
    })
}
```

## Stop Watching

Stop the watcher when your application shuts down:

```go
watcher.Stop()  // Stops polling
```

Use with defer:

```go
cfg, watcher, err := confkit.LoadWithWatcher[Config]("config.yaml", ...)
if err != nil {
    log.Fatal(err)
}

watcher.Start()
defer watcher.Stop()  // Stops on exit

// ... rest of application
```

## Real-World Example: Hot-Reload with HTTP Server

```go
package main

import (
    "fmt"
    "log"
    "net/http"
    "sync"
    "time"
    "github.com/MimoJanra/confkit"
)

type Config struct {
    Port    int    `env:"PORT" default:"8080"`
    Host    string `env:"HOST" default:"localhost"`
    LogLevel string `env:"LOG_LEVEL" default:"info"`
}

type Server struct {
    mu  sync.RWMutex
    cfg Config
}

func (s *Server) Config() Config {
    s.mu.RLock()
    defer s.mu.RUnlock()
    return s.cfg
}

func (s *Server) HandleStatus(w http.ResponseWriter, r *http.Request) {
    cfg := s.Config()
    fmt.Fprintf(w, "Server running on %s:%d\n", cfg.Host, cfg.Port)
}

func main() {
    cfg, watcher, err := confkit.LoadWithWatcher[Config]("config.yaml",
        confkit.FromYAML("config.yaml"),
        confkit.FromEnv(),
    )
    if err != nil {
        log.Fatal(err)
    }

    server := &Server{cfg: cfg}

    // Listen for reloads
    watcher.AddListener(func(old, new any, err error) {
        if err != nil {
            log.Printf("Config reload failed: %v", err)
            return
        }
        
        newCfg := new.(Config)
        oldCfg := old.(Config)
        
        server.mu.Lock()
        server.cfg = newCfg
        server.mu.Unlock()
        
        if oldCfg.LogLevel != newCfg.LogLevel {
            log.Printf("LogLevel changed: %s → %s", oldCfg.LogLevel, newCfg.LogLevel)
        }
        
        log.Println("Config reloaded successfully")
    })

    watcher.SetPollInterval(1 * time.Second)
    watcher.Start()
    defer watcher.Stop()

    http.HandleFunc("/status", server.HandleStatus)
    
    cfg := server.Config()
    addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
    log.Printf("Starting server on %s", addr)
    log.Fatal(http.ListenAndServe(addr, nil))
}
```

## Watching Non-File Sources

For sources like env vars, Vault, or Kubernetes, you can manually reload:

```go
// Manually reload without file watching
cfg, err := confkit.Load[Config](
    confkit.FromEnv(),
    vault.FromVault(...),
)

// To reload, call Load again
newCfg, err := confkit.Load[Config](
    confkit.FromEnv(),
    vault.FromVault(...),
)
```

For automated reloading of non-file sources, use a goroutine with a ticker:

```go
go func() {
    ticker := time.NewTicker(10 * time.Second)
    defer ticker.Stop()
    
    for range ticker.C {
        newCfg, err := confkit.Load[Config](vault.FromVault(...))
        if err != nil {
            log.Printf("Reload failed: %v", err)
            continue
        }
        
        server.mu.Lock()
        server.cfg = newCfg
        server.mu.Unlock()
        
        log.Println("Config reloaded from Vault")
    }
}()
```

## Best Practices

1. **Always use defer to stop the watcher**
   ```go
   watcher.Start()
   defer watcher.Stop()
   ```

2. **Handle reload errors gracefully**
   ```go
   watcher.AddListener(func(old, new any, err error) {
       if err != nil {
           log.Printf("Reload failed: %v", err)
           return
       }
       // Update config
   })
   ```

3. **Use mutexes for thread-safe access**
   ```go
   type Server struct {
       mu  sync.RWMutex
       cfg Config
   }
   ```

4. **Log config changes**
   ```go
   log.Printf("Port changed from %d to %d", old.Port, new.Port)
   ```

5. **Choose appropriate poll interval**
   ```go
   watcher.SetPollInterval(2 * time.Second)  // Balance CPU vs latency
   ```

## Next Steps

- **[Sources](./sources.md)** — All available configuration sources
- **[Recipes](../recipes/)** — Real-world hot-reload examples
