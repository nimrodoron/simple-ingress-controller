package server

type ReverseProxy struct {
	port  int
	rules map[string]string
}

type ProxyConditionOperation int

const (
	ADD    = iota
	REMOVE = iota
)

type ProxyRuleOperation struct {
	Path        string
	ServiceName string
	operation   ProxyConditionOperation
}

func NewReverseProxy(port int) *ReverseProxy {
	return &ReverseProxy{port: port,
		rules: make(map[string]string)}
}

func (p *ReverseProxy) Run(oc <-chan *ProxyRuleOperation) error {

}
