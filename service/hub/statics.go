package hub

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/pkg/errors"
)

const (
	DefaultServerPort int    = 9810
	DefaultListenAddr string = "127.0.0.1"
)

var (
	staticsServer   *http.Server
	uriHandlerCache map[string]httpHandler
)

type httpHandler struct {
	handler    http.Handler
	handleFunc http.HandlerFunc
}

// registerHandler register handler for specified uri
func registerHandler(uri string, handler http.Handler) error {
	if uri == "" {
		return errors.New("uri can not be empty")
	}

	if handler == nil {
		return errors.New("handler can not be empty")
	}

	if !strings.HasPrefix(uri, "/") {
		uri = "/" + uri
	}

	if _, exist := uriHandlerCache[uri]; exist {
		return fmt.Errorf("handler for URI[%s] is already exist", uri)
	}

	uriHandlerCache[uri] = httpHandler{handler: handler}

	return nil
}

// registerHandleFunc register handle function for uri
func registerHandleFunc(uri string, handleFunc http.HandlerFunc) error {
	if uri == "" {
		return errors.New("uri can not be empty")
	}

	if handleFunc == nil {
		return errors.New("handler can not be empty")
	}

	if !strings.HasPrefix(uri, "/") {
		uri = "/" + uri
	}

	if _, exist := uriHandlerCache[uri]; exist {
		return fmt.Errorf("handler for URI[%s] is already exist", uri)
	}

	uriHandlerCache[uri] = httpHandler{handleFunc: handleFunc}

	return nil
}

// StartStaticsServer start statics server
func StartStaticsServer(ctx context.Context, addr string, port int) (err error) {
	if addr == "" || port <= 0 {
		addr = DefaultListenAddr
		port = DefaultServerPort
	}

	listen := fmt.Sprintf("%s:%d", addr, port)

	staticsServer = &http.Server{Addr: listen}
	staticsRouter := http.NewServeMux()
	staticsServer.Handler = staticsRouter

	for uri, h := range uriHandlerCache {
		if h.handler != nil {
			staticsRouter.Handle(uri, h.handler)
			continue
		}

		if h.handleFunc != nil {
			staticsRouter.HandleFunc(uri, h.handleFunc)
			continue
		}

		log.Printf("Registered handler for [%s] is invalid.", uri)
	}

	log.Println("Starting statics server on:", "http://"+listen)

	errCh := make(chan error)

	go func() {
		errCh <- staticsServer.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
	case err = <-errCh:
	}

	return
}

// ShutStaticsServer shutdown server
func ShutStaticsServer(ctx context.Context) error {
	return staticsServer.Shutdown(ctx)
}

// StartDefaultStaticsServer start statics server with default listener
func StartDefaultStaticsServer(ctx context.Context) error {
	return StartStaticsServer(ctx, "", 0)
}

func init() {
	uriHandlerCache = make(map[string]httpHandler)
}
