package app

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"
)

const defaultWorkbenchAddr = "127.0.0.1:0"

func (s Service) Serve(ctx context.Context, request ServeRequest) error {
	outputRoot, err := s.resolveOutputRoot("")
	if err != nil {
		return err
	}

	handler, err := s.workbenchHandler(outputRoot)
	if err != nil {
		return err
	}

	addr := strings.TrimSpace(request.Addr)
	if addr == "" {
		addr = defaultWorkbenchAddr
	}

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen on %q: %w", addr, err)
	}
	defer listener.Close()

	server := &http.Server{
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	if request.Ready != nil {
		request.Ready(ServeResult{
			URL:        workbenchURL(listener.Addr()),
			OutputRoot: outputRoot,
		})
	}

	go func() {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()

	err = server.Serve(listener)
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("serve workbench: %w", err)
	}
	return nil
}

func workbenchURL(address net.Addr) string {
	host, port, err := net.SplitHostPort(address.String())
	if err != nil {
		return "http://" + address.String()
	}

	host = strings.TrimSpace(host)
	switch host {
	case "", "0.0.0.0", "::", "[::]":
		host = "127.0.0.1"
	}

	return "http://" + net.JoinHostPort(host, port)
}
