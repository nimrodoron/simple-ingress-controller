package server

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
)

type ReverseProxy struct {
	port            string
	pathUrlMap      map[string]string
	namePathsMap    map[string][]string
	ruleMux         sync.Mutex
	serviceResolver ServiceResolver
}

type ProxyOperation int

const (
	ADD    = iota
	REMOVE = iota
)

type ProxyRuleOperation struct {
	Name      string
	Rules     map[string]string
	Operation ProxyOperation
}

type Service struct {
	Address string
	Port    string
}

type ServiceResolver interface {
	GetService(name string) (*Service, error)
}

func NewReverseProxy(port string, serviceResolver ServiceResolver) *ReverseProxy {
	return &ReverseProxy{
		port:            port,
		pathUrlMap:      make(map[string]string),
		namePathsMap:    make(map[string][]string),
		serviceResolver: serviceResolver,
	}
}

func (p *ReverseProxy) Run(operationCh <-chan *ProxyRuleOperation) error {

	go p.handleRuleOperations(operationCh)

	// start server
	http.HandleFunc("/", p.handleRequestAndRedirect)
	if err := http.ListenAndServe(":"+p.port, nil); err != nil {
		return err
	}

	return nil
}

func (p *ReverseProxy) handleRuleOperations(operationCh <-chan *ProxyRuleOperation) {
	for o := range operationCh {
		go p.handleRuleOperation(o)
	}
}

func (p *ReverseProxy) handleRuleOperation(ruleOp *ProxyRuleOperation) {
	p.ruleMux.Lock()
	defer p.ruleMux.Unlock()
	for _, path := range p.namePathsMap[ruleOp.Name] {
		delete(p.pathUrlMap, path)
	}
	delete(p.namePathsMap, ruleOp.Name)
	if ruleOp.Operation == ADD {
		keys := make([]string, 0, len(ruleOp.Rules))
		for k, v := range ruleOp.Rules {
			keys = append(keys, k)
			p.pathUrlMap[k] = v
		}
		p.namePathsMap[ruleOp.Name] = keys
	}
}

// Given a request send it to the appropriate url
func (p *ReverseProxy) handleRequestAndRedirect(res http.ResponseWriter, req *http.Request) {
	p.ruleMux.Lock()
	serviceName := p.pathUrlMap[req.URL.Path]
	p.ruleMux.Unlock()
	service, err := p.serviceResolver.GetService(serviceName)
	if err == nil {
		serveReverseProxy(service.Address+":"+service.Port, res, req)
	} else {

	}

	/*	url := getProxyUrl(requestPayload.ProxyCondition)

		logRequestPayload(requestPayload, url)

		serveReverseProxy(url, res, req)*/
}

// Serve a reverse proxy for a given url
func serveReverseProxy(target string, res http.ResponseWriter, req *http.Request) {
	// parse the url
	url, _ := url.Parse(target)

	// create the reverse proxy
	proxy := httputil.NewSingleHostReverseProxy(url)

	// Update the headers to allow for SSL redirection
	req.URL.Host = url.Host
	req.URL.Scheme = url.Scheme
	req.Header.Set("X-Forwarded-Host", req.Header.Get("Host"))
	req.Host = url.Host

	// Note that ServeHttp is non blocking and uses a go routine under the hood
	proxy.ServeHTTP(res, req)
}
