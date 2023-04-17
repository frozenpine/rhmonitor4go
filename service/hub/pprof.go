package hub

import (
	"fmt"
	"log"
	"net/http"
	"net/http/pprof"
	"sort"
	"strings"
)

type profiles struct {
	name           string
	baseURI        string
	profileNames   []string
	extraHandlFunc []string
}

func (p *profiles) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	handler := []string{}
	handler = append(handler, p.profileNames...)
	handler = append(handler, p.extraHandlFunc...)
	sort.Strings(handler)

	w.Write([]byte(`<ul>`))

	for _, name := range handler {
		w.Write([]byte(fmt.Sprintf(`<li><a href="%s">%s</a></li>`, p.GetPprofURI(name), name)))
	}

	w.Write([]byte(`</ul>`))
}

func (p *profiles) GetBaseURI() string {
	uri := strings.Join([]string{p.baseURI, p.name}, "/")

	if !strings.HasPrefix(uri, "/") {
		uri = "/" + uri
	}

	return uri
}

func (p *profiles) GetPprofURI(name string) string {
	for _, pName := range p.profileNames {
		if pName == name {
			return p.GetBaseURI() + "/" + name
		}
	}

	for _, pName := range p.extraHandlFunc {
		if pName == name {
			return p.GetBaseURI() + "/" + name
		}
	}

	return ""
}

var pprofDefine *profiles

func handleFunc(name string) http.HandlerFunc {
	switch name {
	case "profile":
		return pprof.Profile
	case "cmdline":
		return pprof.Cmdline
	case "symbol":
		return pprof.Symbol
	case "trace":
		return pprof.Trace
	}

	return nil
}

// CollectPProfStatics collect pprof statics & expose on uri
func CollectPProfStatics(uri string) error {
	pprofDefine.baseURI = uri

	if err := registerHandler(pprofDefine.GetBaseURI(), pprofDefine); err != nil {
		return err
	}

	for _, profile := range pprofDefine.profileNames {
		if err := registerHandler(
			pprofDefine.GetPprofURI(profile),
			pprof.Handler(profile),
		); err != nil {
			return err
		} else {
			log.Print("Register pprof profile at:", pprofDefine.GetPprofURI(profile))
		}
	}

	for _, name := range pprofDefine.extraHandlFunc {
		if err := registerHandleFunc(
			pprofDefine.GetPprofURI(name),
			handleFunc(name),
		); err != nil {
			return err
		} else {
			log.Print("Register pprof extra functions at:", pprofDefine.GetPprofURI(name))
		}
	}

	log.Printf("Pprof URI registered, expose stats at \"%s\"", pprofDefine.GetBaseURI())

	return nil
}

func init() {
	pprofDefine = &profiles{
		name:           "pprof",
		profileNames:   []string{"goroutine", "threadcreate", "heap", "allocs", "block", "mutex"},
		extraHandlFunc: []string{"profile", "cmdline", "symbol", "trace"},
	}
}
