package main

import (
	"context"
	"fmt"

	firecracker "github.com/firecracker-microvm/firecracker-go-sdk"
	"github.com/firecracker-microvm/firecracker-go-sdk/client/models"
	"github.com/go-openapi/swag"
	"github.com/sirupsen/logrus"
)

func main() {
	// 1) Path to your Firecracker API socket (flintlock or raw firecracker)
	socketPath := "/var/lib/flintlock/vm/01JZZY95PG2X3YQEP0R3D03M7E/firecracker.sock"

	// 2) Set up a logger for the SDK
	logger := logrus.New()
	entry := logrus.NewEntry(logger)

	// 3) Construct the API client
	fcClient := firecracker.NewClient(socketPath, entry, false)

	ctx := context.Background()

	ddfdf, dfdf := fcClient.GetInstanceInfo(ctx)

	fmt.Println("Instance ID:", dfdf)
	fmt.Println("Instance State:", ddfdf.Payload)
	// firecracker.NewMachine()

	// 4) Pause the VM
	pause := &models.InstanceActionInfo{ActionType: swag.String("Pause")}
	if _, err := fcClient.CreateSyncAction(ctx, pause); err != nil {
		panic(fmt.Errorf("pause failed: %w", err))
	}
	fmt.Println("VM is now paused")

	// 5) Take a snapshot (memory + disk)
	snap := &models.SnapshotCreateParams{
		SnapshotPath: swag.String("snapshot.tar"),
		MemFilePath:  swag.String("snapshot.mem"),
	}
	if _, err := fcClient.CreateSnapshot(ctx, snap); err != nil {
		panic(fmt.Errorf("snapshot failed: %w", err))
	}
	fmt.Println("Snapshot written to snapshot.tar + snapshot.mem")

	// 6) Resume the VM
	resume := &models.InstanceActionInfo{ActionType: swag.String("Resume")}
	if _, err := fcClient.CreateSyncAction(ctx, resume); err != nil {
		panic(fmt.Errorf("resume failed: %w", err))
	}
	fmt.Println("VM resumed")
}
