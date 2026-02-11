package instance

import (
	"runtime"
	"sync"
	"unsafe"
)

// Stores instances keyed by thread ID:
// - NTS (standard PHP): single instance
// - ZTS (FrankenPHP): 	 one per thread
var (
	instances      = make(map[uint64]*RequestProcessorInstance)
	instancesMutex sync.RWMutex
	pinners        = make(map[uint64]runtime.Pinner) // Prevents GC while C++ holds pointers
)

// CreateInstance creates or reuses an instance for the given thread.
// Returns an unsafe.Pointer for C++ to store.
func CreateInstance(threadID uint64, isZTS bool) unsafe.Pointer {
	instancesMutex.Lock()
	defer instancesMutex.Unlock()

	if existingInstance, exists := instances[threadID]; exists {
		return unsafe.Pointer(existingInstance)
	}

	instance := NewRequestProcessorInstance(threadID, isZTS)
	instances[threadID] = instance

	// Pin to prevent GC while C++ holds the pointer
	var pinner runtime.Pinner
	pinner.Pin(instance)
	pinners[threadID] = pinner

	return unsafe.Pointer(instance)
}

func GetInstance(instancePtr unsafe.Pointer) *RequestProcessorInstance {
	if instancePtr == nil {
		return nil
	}
	return (*RequestProcessorInstance)(instancePtr)
}

func DestroyInstance(threadID uint64) {
	instancesMutex.Lock()
	defer instancesMutex.Unlock()

	if pinner, exists := pinners[threadID]; exists {
		pinner.Unpin()
		delete(pinners, threadID)
	}

	delete(instances, threadID)
}

// GetAllInstances is used for testing and debugging
func GetAllInstances() []*RequestProcessorInstance {
	instancesMutex.RLock()
	defer instancesMutex.RUnlock()

	result := make([]*RequestProcessorInstance, 0, len(instances))
	for _, instance := range instances {
		result = append(result, instance)
	}
	return result
}

func GetInstanceCount() int {
	instancesMutex.RLock()
	defer instancesMutex.RUnlock()
	return len(instances)
}
