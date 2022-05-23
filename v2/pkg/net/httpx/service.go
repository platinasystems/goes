// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package httpx

import (
	"context"
	"crypto/tls"
	"log"
	"net"
	"net/http"
	"sync"
)

var Elog = log.Println

type Service struct {
	http.Server
	done chan error
}

func NewService(cfg *tls.Config, mux http.Handler) *Service {
	return &Service{
		Server: http.Server{
			Addr:      ":https",
			Handler:   mux,
			TLSConfig: cfg,
		},
		done: make(chan error),
	}
}

func (svc *Service) Shutdown(ctx context.Context) {
	if err := svc.Server.Shutdown(ctx); err != nil {
		Elog(err)
	}
	<-svc.done
}

func (svc *Service) Start(
	wg *sync.WaitGroup,
	ln net.Listener,
	certfn, keyfn string,
) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		svc.done <- svc.ServeTLS(ln, certfn, keyfn)
	}()
}
