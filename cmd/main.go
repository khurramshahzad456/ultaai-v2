package main

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"log"
	"net/http"
	"ultahost-ai-gateway/internal/config"
	"ultahost-ai-gateway/internal/server"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

func main() {
	// Load environment variables
	config.LoadConfig()

	// Initialize server
	s := server.NewServer()
	go ws_tls_conn()

	// Register all routes
	server.RegisterRoutes(s.Engine)

	log.Printf(" Server starting on port %s...\n", config.AppConfig.Port)
	if err := s.Start(); err != nil {
		log.Fatalf(" Server failed to start: %v", err)
	}

}

// Upgrader with CheckOrigin allowing all origins (for demo; restrict in production)
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func ws_tls_conn() {

	// Load CA cert to verify client certs (mutual TLS)
	caCertPEM, err := ioutil.ReadFile("./certs/ca.crt")
	if err != nil {
		log.Fatalf("Failed to read CA cert: %v", err)
	}
	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCertPEM) {
		log.Fatal("Failed to add CA cert to pool")
	}

	// Setup TLS config for server
	tlsConfig := &tls.Config{
		ClientAuth: tls.RequireAndVerifyClientCert, // Require client cert and verify
		ClientCAs:  caCertPool,                     // Client certs must be signed by this CA
		MinVersion: tls.VersionTLS13,
	}

	// Setup Gin router
	r := gin.Default()

	r.GET("/agent/connect", func(c *gin.Context) {
		// Check client certificate info
		if c.Request.TLS == nil || len(c.Request.TLS.PeerCertificates) == 0 {
			c.String(http.StatusUnauthorized, "Client certificate required")
			return
		}

		clientCert := c.Request.TLS.PeerCertificates[0]
		log.Printf("-------Agent connected: CN=%s, Serial=%s", clientCert.Subject.CommonName, clientCert.SerialNumber)

		// Upgrade to WebSocket
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Printf("WebSocket upgrade error: %v", err)
			return
		}
		defer conn.Close()

		// Simple echo loop
		for {
			mt, message, err := conn.ReadMessage()
			if err != nil {
				log.Printf("Read error: %v", err)
				break
			}
			log.Printf("Received from %s: %s", clientCert.Subject.CommonName, string(message))

			err = conn.WriteMessage(mt, message)
			if err != nil {
				log.Printf("Write error: %v", err)
				break
			}
		}
	})

	// Create HTTPS server with TLS config
	server := &http.Server{
		Addr:      ":8443",
		Handler:   r,
		TLSConfig: tlsConfig,
	}

	log.Println("Starting TLS WebSocket server on https://localhost:8443 ...")
	err = server.ListenAndServeTLS("./certs/server.crt", "./certs/server.key")
	if err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
