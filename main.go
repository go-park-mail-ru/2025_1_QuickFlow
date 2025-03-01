package main

import (
    "log"
    "quickflow/internal"
)

func main() {
    if err := internal.Run(); err != nil {
        log.Fatalf("failed to start QuickFlow: %v", err)
    }
}
