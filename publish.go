package main

import (
	"bytes"
	"image/jpeg"
	"image/png"

	"github.com/bluesky-social/indigo/api/atproto"
	bsky "github.com/bluesky-social/indigo/api/bsky"
	lexutil "github.com/bluesky-social/indigo/lex/util"
	"github.com/bluesky-social/indigo/util"
	"github.com/bluesky-social/indigo/xrpc"

	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

func loadAuthFromEnv(req bool) (*xrpc.AuthInfo, error) {
	val := os.Getenv("ATP_AUTH_FILE")
	if val == "" {
		if req {
			return nil, fmt.Errorf("no auth env present, ATP_AUTH_FILE not set")
		}

		return nil, nil
	}

	var auth xrpc.AuthInfo
	if err := json.Unmarshal([]byte(val), &auth); err != nil {
		return nil, err
	}

	return &auth, nil
}

func NewHttpClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}
}

func GetXrpcClient(host string, authreq bool) (*xrpc.Client, error) {
	auth, err := loadAuthFromEnv(authreq)
	if err != nil {
		return nil, fmt.Errorf("loading auth: %w", err)
	}

	return &xrpc.Client{
		Client: NewHttpClient(),
		Host:   host,
		Auth:   auth,
	}, nil
}

func uploadBlob(ctx context.Context, xrpcc *xrpc.Client, imagefile string) (*lexutil.LexBlob, error) {
	data, err := os.ReadFile(imagefile)
	if err != nil {
		return nil, err
	}
	var contentType = "image/png"
	if _, err := png.Decode(bytes.NewBuffer(data)); err != nil {
		if _, err := jpeg.Decode(bytes.NewBuffer(data)); err != nil {
			return nil, err
		}
		contentType = "image/jpeg"
	}
	log.Printf("Content-type: %s", contentType)
	ref, err := atproto.RepoUploadBlob(ctx, xrpcc, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	return ref.Blob, nil
}

func publishFeedGen( ctx context.Context, host string, handle string, password string, cfg *FeedConfig ) error {
	desc := cfg.PublishConfig.ServiceDescription
	did  := cfg.PublishConfig.ServiceDID
	rkey := cfg.PublishConfig.ServiceShortName
	name := rkey
	img  := cfg.PublishConfig.ServiceIcon
	if cfg.PublishConfig.ServiceHumanName != "" {
		name = cfg.PublishConfig.ServiceHumanName
	}

	xrpcc, err := GetXrpcClient(host, false)
	if err != nil {
		return err
	}
	_, err = atproto.ServerCreateSession(ctx, xrpcc, &atproto.ServerCreateSession_Input{
		Identifier: handle,
		Password:   password,
	})
	if err != nil {
		return err
	}

	rec := &lexutil.LexiconTypeDecoder{Val: &bsky.FeedGenerator{
		CreatedAt:   time.Now().Format(util.ISO8601),
		Description: &desc,
		Did:         did,
		DisplayName: name,
	}}

	if img != "" {
		avRef, err := uploadBlob(ctx, xrpcc, img)
		if err == nil && avRef != nil {
			rec.Val.(*bsky.FeedGenerator).Avatar = avRef
		}
	}

	ex, err := atproto.RepoGetRecord(ctx, xrpcc, "", "app.bsky.feed.generator", xrpcc.Auth.Did, rkey)
	if err == nil {
		resp, err := atproto.RepoPutRecord(ctx, xrpcc, &atproto.RepoPutRecord_Input{
			SwapRecord: ex.Cid,
			Collection: "app.bsky.feed.generator",
			Repo:       xrpcc.Auth.Did,
			Rkey:       rkey,
			Record:     rec,
		})
		if err != nil {
			return err
		}

		fmt.Println(resp.Uri)
	} else {
		resp, err := atproto.RepoCreateRecord(ctx, xrpcc, &atproto.RepoCreateRecord_Input{
			Collection: "app.bsky.feed.generator",
			Repo:       xrpcc.Auth.Did,
			Rkey:       &rkey,
			Record:     rec,
		})
		if err != nil {
			return err
		}

		fmt.Println(resp.Uri)
	}

	return nil
}
