package scheduler

type Args interface {
	Check() error
}

type RequestArgs struct {
	AcceptedDomains []string `json:"accepted_primary_domains"`
	MaxDepth        uint32   `json:"max_depth"`
}

func (args *RequestArgs) Check() error
