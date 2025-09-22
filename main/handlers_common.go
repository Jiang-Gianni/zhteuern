package main

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/a-h/templ"
	"github.com/andybalholm/brotli"
	"github.com/benbjohnson/hashfs"
	"github.com/valyala/bytebufferpool"
)

type httpHandlerFunc func(w http.ResponseWriter, r *http.Request) error

var ErrMethodNotAllowed = errors.New("method not allowed")

func (s *Server) GetNotFoundHandler() httpHandlerFunc {
	brotliHandler, err := s.BrotliTempl(ViewPageNotFound())
	if err != nil {
		log.Panicf("brotli: %v", err)
	}
	return func(w http.ResponseWriter, r *http.Request) error {
		return brotliHandler(w, r)
	}
}

type brotliOpts struct {
	level int
}

func (s *Server) BrotliTempl(t templ.Component, opts ...*brotliOpts) (httpHandlerFunc, error) {
	brotliBuf := bytebufferpool.Get()
	defer bytebufferpool.Put(brotliBuf)
	level := brotli.BestCompression
	for _, o := range opts {
		level = o.level
	}
	writer := brotli.NewWriterLevel(brotliBuf, level)
	defer func() {
		if err := writer.Close(); err != nil {
			s.Logger.Error(fmt.Sprintf("brotli.Close: %v", err))
		}
	}()
	if err := t.Render(context.Background(), writer); err != nil {
		return nil, fmt.Errorf("t.Render: %w", err)
	}
	if err := writer.Flush(); err != nil {
		return nil, fmt.Errorf("brotli.Flush: %w", err)
	}
	// copy of the contents
	contents := append([]byte{}, brotliBuf.B...)
	return func(w http.ResponseWriter, r *http.Request) error {
		w.Header().Set(ContentEncoding, "br")
		w.Header().Set(ContentType, TextHtml)
		w.Header().Set(ContentLength, strconv.Itoa(len(contents)))
		_, err := w.Write(contents)
		return err
	}, nil
}

var ErrForbidden = errors.New("request not allowed")

func (s *Server) AllowedRequest(ahh httpHandlerFunc) httpHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		site := r.Header.Get("Sec-Fetch-Site")
		mode := r.Header.Get("Sec-Fetch-Mode")
		isSameSite := (site == "" || site == "none" || site == "same-site" || site == "same-origin")
		isGETCrossSite := (mode == "navigate" && r.Method == http.MethodGet)
		if isSameSite || isGETCrossSite {
			return ahh(w, r)
		}
		return ErrForbidden
	}
}

//go:embed all:brotli
var assetsFs embed.FS
var assetFsys = hashfs.NewFS(assetsFs)

type assetList struct {
	BrotliZhteuernJs, BrotliPicoBlueCss, BrotliStyleCss string
}

var asset = &assetList{
	BrotliZhteuernJs:  "/" + assetFsys.HashName(strings.TrimPrefix(BrotliZhteuernJs, "/")),
	BrotliPicoBlueCss: "/" + assetFsys.HashName(strings.TrimPrefix(BrotliPicoBlueCss, "/")),
	BrotliStyleCss:    "/" + assetFsys.HashName(strings.TrimPrefix(BrotliStyleCss, "/")),
}

func (s *Server) GetAssetsHandler() httpHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		w.Header().Set(ContentEncoding, "br")

		if environment != EnvironmentDEV {
			hashfs.FileServer(assetFsys).ServeHTTP(w, r)
			return nil
		}

		contents, err := os.ReadFile("./main" + r.URL.Path)
		if err != nil {
			return fmt.Errorf("os.ReadFile: %w", err)
		}
		switch {
		case strings.HasSuffix(r.URL.Path, ".css"):
			w.Header().Set(ContentType, TextCss)
		case strings.HasSuffix(r.URL.Path, ".html"):
			w.Header().Set(ContentType, TextHtml)
		case strings.HasSuffix(r.URL.Path, ".js"):
			w.Header().Set(ContentType, TextJavaScript)
		}
		_, err = w.Write(contents)
		return err
	}
}

func (s *Server) HotReloadHandler() httpHandlerFunc {
	forceReload := make(chan struct{})
	var onServerStartup sync.Once
	return func(w http.ResponseWriter, r *http.Request) error {
		if r.Header.Get("Force") != "" {
			select {
			case forceReload <- struct{}{}:
			default:
			}
			return nil
		}

		sse, err := StartSSE(w, r)
		if err != nil {
			return fmt.Errorf("StartSSE: %w", err)
		}
		sse.rc.SetWriteDeadline(time.Time{})
		reloadFunc := func() {
			err = sse.WriteMessage(r.Context(), w, &SSEventMessage{
				JavaScript: "window.location.reload();",
			})
			if err != nil {
				err = fmt.Errorf("sse.WriteMessage: %w", err)
			}
		}
		onServerStartup.Do(reloadFunc)
		if err != nil {
			return err
		}

		for {
			select {
			case <-r.Context().Done():
				return nil
			case <-forceReload:
				reloadFunc()
				if err != nil {
					return err
				}
			case <-s.shutdownChan:
				/*
					manually hijacking and closing the connection to trigger
					fetch().catch() exception branch to retry again for hot relodad
				*/
				return s.HijackAndClose(w)
			}
		}
	}
}

func (s *Server) HijackAndClose(w http.ResponseWriter) error {
	hijacker, ok := w.(http.Hijacker)
	if ok {
		conn, _, _ := hijacker.Hijack()
		return conn.Close()
	}
	return nil
}
