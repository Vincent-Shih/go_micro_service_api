package helper

import (
	"sync"
	"time"
)

type MachineID int64

const (
	epoch         = 1672531200_000_000 // Custom epoch (e.g., Jan 1, 2023)
	machineIDBits = 10                 // Number of bits for machine ID
	sequenceBits  = 12                 // Number of bits for sequence
)

type Snowflake struct {
	mu        sync.Mutex
	lastTime  int64
	sequence  int64
	machineID int64
}

func NewMachineID() MachineID {
	return 1 // Hardcoded machine ID
}

func NewSnowflake(machineID MachineID) *Snowflake {
	return &Snowflake{
		mu:        sync.Mutex{},
		lastTime:  time.Now().UnixMicro(),
		sequence:  sequenceBits,
		machineID: int64(machineID),
	}
}

func (sf *Snowflake) NextID() int64 {
	sf.mu.Lock()
	defer sf.mu.Unlock()

	now := time.Now().UnixMicro()
	if now == sf.lastTime {
		sf.sequence = (sf.sequence + 1) & ((1 << sequenceBits) - 1) // Cycle sequence
	} else {
		sf.sequence = 0 // Reset sequence if timestamp changes
	}
	sf.lastTime = now

	// Combine all parts to create the ID
	return ((now - epoch) << (machineIDBits + sequenceBits)) | (sf.machineID << sequenceBits) | sf.sequence
}
