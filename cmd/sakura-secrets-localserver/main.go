package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"

	"github.com/fujiwara/sakura-secrets-cli/localserver"
)

func main() {
	addr := flag.String("addr", ":8080", "listen address")
	prefix := flag.String("prefix", "/api/cloud/1.1", "URL path prefix")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	srv := &http.Server{
		Addr:    *addr,
		Handler: localserver.NewServer(*prefix),
	}

	go func() {
		<-ctx.Done()
		log.Println("shutting down...")
		srv.Shutdown(context.Background())
	}()

	host := *addr
	if host == "" || host[0] == ':' {
		host = "localhost" + host
	}
	rootURL := fmt.Sprintf("http://%s%s", host, *prefix)

	fmt.Fprintln(log.Writer(), "sakura-secrets-cli localserver is running.")
	fmt.Fprintln(log.Writer(), "")
	fmt.Fprintln(log.Writer(), "To connect sakura-secrets-cli to this server, set the following environment variables:")
	fmt.Fprintln(log.Writer(), "")
	fmt.Fprintf(log.Writer(), "  export SAKURA_API_ROOT_URL=%s\n", rootURL)
	fmt.Fprintln(log.Writer(), "  export SAKURA_ACCESS_TOKEN=dummy")
	fmt.Fprintln(log.Writer(), "  export SAKURA_ACCESS_TOKEN_SECRET=dummy")
	fmt.Fprintln(log.Writer(), "  export VAULT_ID=your-vault-id")
	fmt.Fprintln(log.Writer(), "")

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
