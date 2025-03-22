package main

import (
    "log"

    "quickflow/config"
    "quickflow/config/cors"
    "quickflow/internal"
)

func main() {
    cfg, err := config.Parse()
    if err != nil {
        log.Fatalf("failed to load QuickFlow configuration: %v", err)
    }

    corsCfg, err := cors.ParseCORS()
    if err != nil {
        log.Fatalf("failed to load CORS configuration: %v", err)
    }

    if err = internal.Run(cfg, corsCfg); err != nil {
        log.Fatalf("failed to start QuickFlow: %v", err)
    }
}
