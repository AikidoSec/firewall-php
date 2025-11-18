package instance

import (
	"runtime"
	"sync"
	"unsafe"
)

// Global instance storage - unified for both NTS and ZTS
// Key: Thread ID
//
//   - NTS: Always use threadID = 0 (single process-wide instance)
//   - ZTS: Use pthread_self() (unique per thread)
var (
	instances      = make(map[uint64]*RequestProcessorInstance)
	instancesMutex sync.RWMutex
	pinners        = make(map[uint64]runtime.Pinner) // Keeps instances pinned for CGO
)

// CreateInstance creates and stores a new instance
//
// threadID:
//
//   - For NTS: pass 0 (creates/reuses single instance)
//   - For ZTS: pass pthread_self() (creates per-thread instance)
//
// isZTS:
//
//   - true if running in Franken PHP (ZTS mode)
//   - false if running in standard PHP (NTS mode)
//
// Returns: unsafe.Pointer to the instance (for C++ to store)
func CreateInstance(threadID uint64, isZTS bool) unsafe.Pointer {
	instancesMutex.Lock()
	defer instancesMutex.Unlock()

	// Check if instance already exists for this thread/process
	if existingInstance, exists := instances[threadID]; exists {
		return unsafe.Pointer(existingInstance)
	}

	// Create new instance
	instance := NewRequestProcessorInstance(isZTS)
	instances[threadID] = instance

	// Pin the instance to prevent garbage collection while C++ holds pointer
	var pinner runtime.Pinner
	pinner.Pin(instance)
	pinners[threadID] = pinner

	return unsafe.Pointer(instance)
}

// GetInstance retrieves an instance by its pointer
func GetInstance(instancePtr unsafe.Pointer) *RequestProcessorInstance {
	if instancePtr == nil {
		return nil
	}
	return (*RequestProcessorInstance)(instancePtr)
}

// DestroyInstance removes an instance from storage
//
// threadID:
//
//   - For NTS: pass 0
//   - For ZTS: pass pthread_self()
func DestroyInstance(threadID uint64) {
	instancesMutex.Lock()
	defer instancesMutex.Unlock()

	// Unpin the instance to allow garbage collection
	if pinner, exists := pinners[threadID]; exists {
		pinner.Unpin()
		delete(pinners, threadID)
	}

	delete(instances, threadID)
}

// GetAllInstances returns all active instances (for testing/debugging)
func GetAllInstances() []*RequestProcessorInstance {
	instancesMutex.RLock()
	defer instancesMutex.RUnlock()

	result := make([]*RequestProcessorInstance, 0, len(instances))
	for _, instance := range instances {
		result = append(result, instance)
	}
	return result
}

// GetInstanceCount returns the number of active instances
func GetInstanceCount() int {
	instancesMutex.RLock()
	defer instancesMutex.RUnlock()
	return len(instances)
}
