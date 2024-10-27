package tools

import (
	"crypto/x509"
	"log"
	"sync"
	"sync/atomic"
)

var (
	systemCertificatePool atomic.Pointer[x509.CertPool]
	once                  sync.Once
)

func InitialCerts() {
	once.Do(func() {
		// Get the system's CA certificate pool
		certPool, err := x509.SystemCertPool()
		if err != nil {
			// Handle the error (e.g., log it or panic)
			log.Fatalf("Error getting system CA pool:%+v", err)
		}
		systemCertificatePool.Store(certPool)
	})
}

func SystemCertPool() *x509.CertPool {
	load := systemCertificatePool.Load()
	return load
}
