package server

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

type WSTLSServer struct {
	http *http.Server
}

func StartWSTLSServer(addr string, h *gin.Engine) *WSTLSServer {
	caCertPEM, err := os.ReadFile("./certs/ca.crt")
	if err != nil {
		log.Fatalf("Failed to read CA cert: %v", err)
	}
	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCertPEM) {
		log.Fatal("Failed to add CA cert to pool")
	}

	tlsConfig := &tls.Config{
		ClientAuth: tls.RequireAndVerifyClientCert,
		ClientCAs:  caCertPool,
		MinVersion: tls.VersionTLS13,
	}

	srv := &http.Server{
		Addr:              addr,
		Handler:           h,
		TLSConfig:         tlsConfig,
		ReadTimeout:       60 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}

	go func() {
		log.Printf("TLS WebSocket server listening on %s ...", addr)
		if err := srv.ListenAndServeTLS("./certs/server.crt", "./certs/server.key"); err != nil && err != http.ErrServerClosed {
			log.Fatalf("WS TLS server failed: %v", err)
		}
	}()

	return &WSTLSServer{http: srv}
}

func (s *WSTLSServer) Shutdown(ctx context.Context) error {
	return s.http.Shutdown(ctx)
}
