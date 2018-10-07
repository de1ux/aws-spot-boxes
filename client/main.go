package main

import (
	"context"
	"github.com/de1ux/aws-spot-boxes/generated/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"log"
	"net/http"
	"net/url"
	"os/exec"

	"github.com/de1ux/aws-spot-boxes/common"

	"golang.org/x/oauth2"
)

var (
	c *common.Config
)

func init() {
	var err error
	c, err = common.GetConfig()
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	idToken := ""//getIDToken()
	conn, err := grpc.Dial(c.ClientServiceURL, grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}

	ctx := metadata.AppendToOutgoingContext(context.Background(), "id_token", idToken)

	client := api.NewAWSSpotBoxesClient(conn)

	_, err = client.StartBox(ctx, &api.StartBoxRequest{})
	if err != nil {
		log.Fatal(err)
	}
}

func getIDToken() (string) {
	c, err := common.GetConfig()
	if err != nil {
		log.Fatal(err)
	}

	srv := &http.Server{Addr: ":" + c.ClientRedirectPort}

	done := make(chan string)
	conf := &oauth2.Config{
		ClientID:     c.ClientId,
		ClientSecret: c.ClientSecret,
		Scopes:       []string{"profile", "email"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://accounts.google.com/o/oauth2/auth",
			TokenURL: "https://accounts.google.com/o/oauth2/token",
		},
		RedirectURL: "http://localhost:" + c.ClientRedirectPort,
	}

	loginURL := conf.AuthCodeURL("state", oauth2.AccessTypeOffline)

	if err := exec.Command("open", loginURL).Run(); err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		queryParts, _ := url.ParseQuery(req.URL.RawQuery)
		code := queryParts["code"][0]
		tok, err := conf.Exchange(context.Background(), code)
		if err != nil {
			log.Fatal(err)
		}
		done <- tok.Extra("id_token").(string)
	})

	go func() {
		srv.ListenAndServe()
	}()

	log.Printf("Waiting for redirect with id_token...")
	idToken := <-done
	log.Printf(
		"Waiting for redirect with id_token...got it!")

	go func() {
		srv.Shutdown(context.Background())
	}()
	return idToken
}