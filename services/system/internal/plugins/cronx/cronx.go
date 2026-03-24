package cronx

import (
	"fmt"
	"sync"
)

var (
	registryMu sync.RWMutex
	registry   = make(map[string]JobHandler)
)

func Register(invokeTarget string, handler JobHandler) {
	registryMu.Lock()
	defer registryMu.Unlock()

	if invokeTarget == "" {
		panic("cron handler invoke target is empty")
	}
	if handler == nil {
		panic(fmt.Sprintf("cron handler [%s] is nil", invokeTarget))
	}
	if _, exists := registry[invokeTarget]; exists {
		panic(fmt.Sprintf("cron handler [%s] already registered", invokeTarget))
	}

	registry[invokeTarget] = handler
}

func GetRegisteredHandlers() map[string]JobHandler {
	registryMu.RLock()
	defer registryMu.RUnlock()

	result := make(map[string]JobHandler, len(registry))
	for k, v := range registry {
		result[k] = v
	}
	return result
}
