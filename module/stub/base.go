package stub

import "webcrawler/module"

type ModuleInternal interface {
	module.Module
	IncreCalledCount()
	IncrAcceptedCount()
	IncrCompletedCount()
	IncrHandlingNumber()
	DecrHandlingNumber()
	Clear()
}
