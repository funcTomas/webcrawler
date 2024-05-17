package module

import (
	"fmt"
	"sync"
	"webcrawler/errors"
)

type Registrar interface {
	Register(module Module) (bool, error)
	Unregister(mid MID) (bool, error)
	Get(moduleType Type) (Module, error)
	GetAllByType(moduleType Type) (map[MID]Module, error)
	GetAll() map[MID]Module
	Clear()
}

func NewRegistrar() Registrar {
	return &myRegistrar{
		moduleTypeMap: map[Type]map[MID]Module{},
	}
}

type myRegistrar struct {
	moduleTypeMap map[Type]map[MID]Module
	rwlock        sync.RWMutex
}

func (registrar *myRegistrar) Register(module Module) (bool, error) {
	if module == nil {
		return false, errors.NewIllegalParameterError("nil module instance")
	}
	mid := module.ID()
	parts, err := SplitMID(mid)
	if err != nil {
		return false, err
	}
	moduleType := legalLetterTypeMap[parts[0]]
	if !CheckType(moduleType, module) {
		errMsg := fmt.Sprintf("incorrect module type %s", moduleType)
		return false, errors.NewIllegalParameterError(errMsg)
	}
	registrar.rwlock.Lock()
	defer registrar.rwlock.Unlock()
	modules := registrar.moduleTypeMap[moduleType]
	if modules == nil {
		modules = map[MID]Module{}
	}
	if _, has := modules[mid]; has {
		return false, nil
	}
	modules[mid] = module
	registrar.moduleTypeMap[moduleType] = modules
	return true, nil
}

func (registrar *myRegistrar) Unregister(mid MID) (bool, error) {
	parts, err := SplitMID(mid)
	if err != nil {
		return false, err
	}

	moduleType := legalLetterTypeMap[parts[0]]
	var deleted bool
	registrar.rwlock.Lock()
	defer registrar.rwlock.Unlock()
	if modules, ok := registrar.moduleTypeMap[moduleType]; ok {
		if _, has := modules[mid]; has {
			delete(modules, mid)
			deleted = true
		}
	}
	return deleted, nil
}

func (registrar *myRegistrar) Get(moduleType Type) (Module, error) {
	modules, err := registrar.GetAllByType(moduleType)
	if err != nil {
		return nil, err
	}
	minScore := uint64(0)
	var selectedModule Module
	for _, module := range modules {
		SetScore(module)
		score := module.Score()
		if minScore == 0 || score < minScore {
			selectedModule = module
			minScore = score
		}
	}
	return selectedModule, nil

}

func (registrar *myRegistrar) GetAllByType(moduleType Type) (map[MID]Module, error) {
	if !LegalType(moduleType) {
		errMsg := fmt.Sprintf("illegal module type: %s", moduleType)
		return nil, errors.NewIllegalParameterError(errMsg)
	}
	registrar.rwlock.Lock()
	defer registrar.rwlock.Unlock()
	modules := registrar.moduleTypeMap[moduleType]
	if len(modules) == 0 {
		return nil, ErrNotFoundModuleInstance
	}
	result := map[MID]Module{}
	for mid, module := range modules {
		result[mid] = module
	}
	return result, nil
}

func (registrar *myRegistrar) GetAll() map[MID]Module {
	result := map[MID]Module{}
	registrar.rwlock.RLock()
	defer registrar.rwlock.RUnlock()
	for _, modules := range registrar.moduleTypeMap {
		for mid, module := range modules {
			result[mid] = module
		}
	}
	return result
}

func (registrar *myRegistrar) Clear() {
	registrar.rwlock.Lock()
	defer registrar.rwlock.Unlock()
	registrar.moduleTypeMap = map[Type]map[MID]Module{}
}
