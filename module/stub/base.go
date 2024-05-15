package stub

import "webcrawler/module"

type ModuleInternal interface {
	module.Module
	IncrCalledCount()
	IncrAcceptedCount()
	IncrCompletedCount()
	IncrHandlingNumber()
	DecrHandlingNumber()
	Clear()
}
