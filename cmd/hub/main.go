package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"campus-device-hub/internal/adapters"
	"campus-device-hub/internal/dashboard"
	"campus-device-hub/internal/service"
	"campus-device-hub/internal/storage"
)

func main() {
	path := filepath.Join(os.TempDir(), "campus-device-hub.db")
	store, err := storage.Open(path)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer store.Close()
	clock := service.NewFixedClock(time.Date(2026, 1, 15, 8, 0, 0, 0, time.UTC))
	syncer := service.NewSyncService(store, clock, adapters.DefaultFixtures())
	if _, err := syncer.SyncAll(); err != nil {
		fmt.Println(err)
		return
	}
	view := dashboard.NewView(store)
	summary, err := view.Summary()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("campus device status: %d devices, %d healthy, %d failed\n", summary.TotalDevices, summary.HealthyDevices, summary.FailedSyncs)
}
