package cronx

import (
	"fmt"
	"sync"
)

var (
	registryMu sync.RWMutex
	registry   = make(map[string]JobHandler)
	registered = make(map[string]string)
)

func Register(invokeTarget string, invokeName string, handler JobHandler) {
	registryMu.Lock()
	defer registryMu.Unlock()

	if invokeTarget == "" {
		panic("cron handler invoke target is empty")
	}
	if invokeName == "" {
		panic("cron handler invoke name is empty")
	}
	if handler == nil {
		panic(fmt.Sprintf("cron handler [%s] is nil", invokeTarget))
	}
	if _, exists := registry[invokeTarget]; exists {
		panic(fmt.Sprintf("cron handler [%s] already registered", invokeTarget))
	}

	registry[invokeTarget] = handler
	registered[invokeTarget] = invokeName
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

func GetRegisteredNames() map[string]string {
	registryMu.RLock()
	defer registryMu.RUnlock()

	result := make(map[string]string, len(registered))
	for k, v := range registered {
		result[k] = v
	}
	return result
}
