package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/Jiang-Gianni/zhteuern/browser"
	"github.com/a-h/templ"
	"github.com/valyala/bytebufferpool"
)

type SSE struct {
	m  *sync.Mutex
	rc *http.ResponseController
}

func StartSSE(w http.ResponseWriter, r *http.Request) (*SSE, error) {
	rc := http.NewResponseController(w)
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Content-Type", "text/event-stream")
	if r.ProtoMajor == 1 {
		w.Header().Set("Connection", "keep-alive")
	}
	if err := rc.Flush(); err != nil {
		return nil, fmt.Errorf("flush: %w", err)
	}
	return &SSE{m: &sync.Mutex{}, rc: rc}, nil
}

type SSEventMessage struct {
	Data           []byte
	Templ          templ.Component
	JavaScript     string
	BrowserUpdates []*browser.Update
	Event          SSEvent
	Selector       string
}

func (sse *SSE) WriteMessage(ctx context.Context, w io.Writer, d *SSEventMessage) (err error) {
	if sse.m != nil {
		sse.m.Lock()
		defer sse.m.Unlock()
	}

	buf := bytebufferpool.Get()
	defer bytebufferpool.Put(buf)

	if d.Templ != nil {
		templBuf := bytebufferpool.Get()
		defer bytebufferpool.Put(templBuf)
		if err := d.Templ.Render(ctx, templBuf); err != nil {
			return fmt.Errorf("t.Render: %w", err)
		}
		d.Data = append([]byte{}, templBuf.B...)
	}

	if len(d.JavaScript) > 0 {
		d.Event = EventAppend
		d.Data = append(d.Data, "<script>document.currentScript.remove();"...)
		d.Data = append(d.Data, d.JavaScript...)
		d.Data = append(d.Data, "</script>"...)
		d.Selector = "body"
	}

	if len(d.BrowserUpdates) > 0 {
		d.Event = EventDOM
		d.Data, err = json.Marshal(d.BrowserUpdates)
		if err != nil {
			return fmt.Errorf("json.Marshal: %w", err)
		}
	}

	if d.Selector != "" {
		buf.B = append(buf.B, "id: "...)
		buf.B = append(buf.B, d.Selector...)
		buf.B = append(buf.B, "\n"...)
	}
	if len(d.Event) > 0 {
		buf.B = append(buf.B, d.Event...)
	}

	if len(d.Data) > 0 {
		start := 0
		for i := 0; i <= len(d.Data); i++ {
			if i == len(d.Data) || d.Data[i] == '\n' {
				line := d.Data[start:i]
				buf.B = append(buf.B, "data: "...)
				buf.B = append(buf.B, line...)
				buf.B = append(buf.B, '\n')
				start = i + 1
			}
		}
	}

	buf.B = append(buf.B, []byte("\n")...)

	_, err = w.Write(buf.B)
	if err != nil {
		return fmt.Errorf("w.Write: %w", err)
	}
	if fw, ok := w.(interface{ Flush() error }); ok {
		err := fw.Flush()
		if err != nil {
			return fmt.Errorf("fw.Flush: %w", err)
		}
	}
	if sse.rc != nil {
		err = sse.rc.Flush()
		if err != nil {
			return fmt.Errorf("sse.rc.Flush: %w", err)
		}
	}
	return nil
}

type SSEvent []byte

var (
	EventDOM     SSEvent = []byte("event: dom\n")
	EventPrepend SSEvent = []byte("event: prepend\n")
	EventAppend  SSEvent = []byte("event: append\n")
	EventBefore  SSEvent = []byte("event: before\n")
	EventAfter   SSEvent = []byte("event: after\n")
	EventReplace SSEvent = []byte("event: replace\n")
)
