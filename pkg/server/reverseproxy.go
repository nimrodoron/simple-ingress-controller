package server

import (
	"fmt"
	"k8s.io/apimachinery/pkg/util/proxy"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
)

type ReverseProxy struct {
	port            int32
	pathUrlMap      map[string]string
	namePathsMap    map[string][]string
	ruleMux         sync.Mutex
	serviceResolver ServiceResolver
	proxyUrl        *url.URL
	proxy           *httputil.ReverseProxy
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
	Port    int32
}

type ServiceResolver interface {
	GetService(name string) (*Service, error)
}

func NewReverseProxy(port int32, serviceResolver ServiceResolver) *ReverseProxy {
	url, _ := url.Parse("/")
	return &ReverseProxy{
		port:            port,
		pathUrlMap:      make(map[string]string),
		namePathsMap:    make(map[string][]string),
		serviceResolver: serviceResolver,
		proxyUrl:        url,
		proxy:           httputil.NewSingleHostReverseProxy(url),
	}
}

func (p *ReverseProxy) Run(operationCh <-chan *ProxyRuleOperation) error {

	go p.handleRuleOperations(operationCh)

	// start server
	http.HandleFunc("/", p.handleRequestAndRedirect)
	if err := http.ListenAndServe(":"+fmt.Sprint(p.port), nil); err != nil {
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
		p.serveReverseProxy(service.Address+":"+fmt.Sprint(service.Port), res, req)
	}
}

// Serve a reverse proxy for a given url
func (p *ReverseProxy) serveReverseProxy(target string, res http.ResponseWriter, req *http.Request) {

	// Update the headers to allow for SSL redirection
	req.URL.Host = p.proxyUrl.Host
	req.URL.Scheme = p.proxyUrl.Scheme
	req.Header.Set("X-Forwarded-Host", req.Header.Get("Host"))
	req.Host = p.proxyUrl.Host

	// Note that ServeHttp is non blocking and uses a go routine under the hood
	p.proxy.ServeHTTP(res, req)
}
