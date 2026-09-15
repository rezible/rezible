package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/dexidp/dex/api/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatal("usage: dex-client <wait|ensure|delete>")
	}

	port := os.Getenv("DEX_GRPC_PORT")
	if port == "" {
		port = "7015"
	}
	conn, err := grpc.NewClient("localhost:"+port, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := api.NewDexClient(conn)
	switch os.Args[1] {
	case "wait":
		err = wait(client)
	case "ensure":
		err = ensure(client)
	case "delete":
		err = deleteClient(client)
	default:
		log.Fatalf("unknown operation %q", os.Args[1])
	}
	if err != nil {
		log.Fatal(err)
	}
}

func wait(client api.DexClient) error {
	for range 30 {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		_, err := client.GetVersion(ctx, &api.VersionReq{})
		cancel()
		if err == nil {
			return nil
		}
		time.Sleep(time.Second)
	}
	return fmt.Errorf("Dex management API is not ready")
}

func ensure(client api.DexClient) error {
	id, secret, appURL := os.Getenv("OIDC_CLIENT_ID"), os.Getenv("OIDC_CLIENT_SECRET"), os.Getenv("APP_URL")
	if id == "" || secret == "" || appURL == "" {
		return fmt.Errorf("OIDC_CLIENT_ID, OIDC_CLIENT_SECRET, and APP_URL are required")
	}
	name := "Rezible " + os.Getenv("WORKSPACE_ID")
	redirectURIs := []string{appURL + "/api/auth/callback"}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	created, err := client.CreateClient(ctx, &api.CreateClientReq{Client: &api.Client{
		Id: id, Secret: secret, Name: name, RedirectUris: redirectURIs,
	}})
	if err != nil {
		return fmt.Errorf("create client: %w", err)
	}
	if !created.AlreadyExists {
		return nil
	}

	updated, err := client.UpdateClient(ctx, &api.UpdateClientReq{Id: id, Name: name, RedirectUris: redirectURIs})
	if err != nil {
		return fmt.Errorf("update client: %w", err)
	}
	if updated.NotFound {
		return fmt.Errorf("client %q disappeared during update", id)
	}
	return nil
}

func deleteClient(client api.DexClient) error {
	id := os.Getenv("OIDC_CLIENT_ID")
	if id == "" {
		return fmt.Errorf("OIDC_CLIENT_ID is required")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := client.DeleteClient(ctx, &api.DeleteClientReq{Id: id})
	if err != nil {
		return fmt.Errorf("delete client: %w", err)
	}
	return nil
}
