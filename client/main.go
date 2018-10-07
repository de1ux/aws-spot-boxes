package main

import (
	"context"
	"log"
	"net/http"
	"net/url"
	"os/exec"

	"github.com/kelseyhightower/envconfig"
	"golang.org/x/oauth2"
)

type config struct {
	ClientId string
	ClientSecret string
	ClientRedirectPort string
}

func main() {
	println(getIDToken())
}

func getIDToken() (string) {
	c := &config{}
	if err := envconfig.Process("awsspotboxes", c); err != nil {
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
		if err := srv.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()

	idToken := <-done
	srv.Shutdown(context.Background())

	return idToken
}