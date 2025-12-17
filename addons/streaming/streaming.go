package streaming

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/buildwithgo/amaro"
)

// StreamContext wraps the standard Context and adds a Flusher.
// It allows modifying the response before or during the stream.
type StreamContext struct {
	amaro.Context
	flusher http.Flusher
}

func (c *StreamContext) Write(data []byte) (int, error) {
	if c.flusher != nil {
		c.flusher.Flush()
	}
	return c.Context.Writer.Write(data)
}

// Stream initiates a basic streaming response.
// It sets proper headers (Connection: keep-alive) and provides a StreamContext
// to the callback for flushing data manually.
func Stream(ctx *amaro.Context, call func(StreamContext)) error {
	if ctx == nil {
		return errors.New("context cannot be nil")
	}

	ctx.Writer.Header().Set("Cache-Control", "no-cache")
	ctx.Writer.Header().Set("Connection", "keep-alive")

	flusher, ok := ctx.Writer.(http.Flusher)
	if !ok {
		return errors.New("streaming not supported by the response writer")
	}

	streamCtx := StreamContext{
		Context: *ctx,
		flusher: flusher,
	}

	call(streamCtx)

	return nil
}

type StreamTextContext struct {
	sc *StreamContext
}

func (s *StreamTextContext) WriteLn(line string, args ...any) (int, error) {
	return s.sc.Write([]byte(fmt.Sprintf(line, args...) + "\n"))
}

// StreamText initiates a text/plain streaming response with chunked transfer encoding.
func StreamText(ctx *amaro.Context, call func(*StreamTextContext)) error {
	if ctx == nil {
		return errors.New("context cannot be nil")
	}

	ctx.Writer.Header().Set("Content-Type", "text/plain")
	ctx.Writer.Header().Set("Cache-Control", "no-cache")
	ctx.Writer.Header().Set("Connection", "keep-alive")
	ctx.Writer.Header().Set("Transfer-Encoding", "chunked")
	ctx.Writer.Header().Set("X-Content-Type-Options", "nosniff")

	flusher, ok := ctx.Writer.(http.Flusher)
	if !ok {
		return errors.New("streaming not supported by the response writer")
	}

	streamCtx := &StreamContext{
		Context: *ctx,
		flusher: flusher,
	}

	streamText := &StreamTextContext{sc: streamCtx}
	call(streamText)

	return nil
}

type SSEMessage struct {
	Data  string  `json:"data"`
	Event *string `json:"event,omitempty"`
	ID    *string `json:"id,omitempty"`
	Retry *int    `json:"retry,omitempty"`
}

type StreamSSEContext struct {
	sc *StreamContext
}

func (s *StreamSSEContext) Send(msg SSEMessage) error {
	if msg.Data == "" {
		return errors.New("data field cannot be empty")
	}

	if msg.Event != nil {
		_, err := s.sc.Write([]byte(fmt.Sprintf("event: %s\n", *msg.Event)))
		if err != nil {
			return err
		}
	}

	if msg.ID != nil {
		_, err := s.sc.Write([]byte(fmt.Sprintf("id: %s\n", *msg.ID)))
		if err != nil {
			return err
		}
	}

	if msg.Retry != nil {
		_, err := s.sc.Write([]byte(fmt.Sprintf("retry: %d\n", *msg.Retry)))
		if err != nil {
			return err
		}
	}

	_, err := s.sc.Write([]byte(fmt.Sprintf("data: %s\n\n", msg.Data)))
	return err
}

func StreamSSE(c *amaro.Context, call func(StreamSSEContext)) error {
	if c == nil {
		return errors.New("context cannot be nil")
	}

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		return errors.New("streaming not supported by the response writer")
	}

	streamCtx := StreamContext{
		Context: *c,
		flusher: flusher,
	}

	sseCtx := StreamSSEContext{sc: &streamCtx}
	call(sseCtx)

	return nil
}
