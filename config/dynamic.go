package config

import "sync"

type DynamicConf_T uint32

const (
	// 允许公开注册
	EnablePublicRegister DynamicConf_T = iota
)

type DynamicConfig struct {
	conf map[DynamicConf_T]any
	lock sync.RWMutex
}

func (dc *DynamicConfig) EnablePublicRegister() bool {
	dc.lock.RLock()
	defer dc.lock.RUnlock()
	if val, ok := dc.conf[EnablePublicRegister]; ok {
		if enabled, ok := val.(bool); ok {
			return enabled
		}
		return false
	}
	return false
}

func (dc *DynamicConfig) Set(key DynamicConf_T, value any) {
	dc.lock.Lock()
	defer dc.lock.Unlock()
	dc.conf[key] = value
}

var DynamicConf *DynamicConfig = &DynamicConfig{
	conf: make(map[DynamicConf_T]any),
	lock: sync.RWMutex{},
}
