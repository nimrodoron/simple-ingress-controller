package server

import "net/http"

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
	Operation   ProxyConditionOperation
}

func NewReverseProxy(port int) *ReverseProxy {
	return &ReverseProxy{port: port,
		rules: make(map[string]string)}
}

func (p *ReverseProxy) Run(oc <-chan *ProxyRuleOperation) error {
	// start server
	//http.HandleFunc("/", handleRequestAndRedirect)
	if err := http.ListenAndServe(":"+string(p.port), nil); err != nil {
		return err
	}

	return nil
}

/*// Given a request send it to the appropriate url
func handleRequestAndRedirect(res http.ResponseWriter, req *http.Request) {
	requestPayload := parseRequestBody(req)
	url := getProxyUrl(requestPayload.ProxyCondition)

	logRequestPayload(requestPayload, url)

	serveReverseProxy(url, res, req)
}*/
