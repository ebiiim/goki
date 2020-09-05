# GOKI

[![Build Status](https://travis-ci.org/ebiiim/goki.svg?branch=master)](https://travis-ci.org/ebiiim/goki)
[![Go Report Card](https://goreportcard.com/badge/github.com/ebiiim/goki)](https://goreportcard.com/report/github.com/ebiiim/goki)
[![MIT license](https://img.shields.io/badge/License-MIT-blue.svg)](https://lbesson.mit-license.org)

## What is this?

A website that counts the number of cockroaches you've rid this year.

👉[ゴキブリやっつけた！](https://goki.nullpo-t.net)

## Deploy

Prepare `config.json`.

```json
{
  "server": {
    "scheme": "https",
    "address": "goki.nullpo-t.net",
    "base_path": "/"
  },
  "web": {
    "template_dir": "./views",
    "static_dir": "./static",
    "serve_static": true
  },
  "session": {
    "key": "goki"
  },
  "twitter": {
    "key": "TWITTER_CONSUMER_KEY",
    "secret": "TWITTER_CONSUMER_SECRET",
    "request_url": "https://api.twitter.com/oauth/request_token",
    "authorize_url": "https://api.twitter.com/oauth/authorize",
    "token_request_url": "https://api.twitter.com/oauth/access_token",
    "callback_path": "/login/twitter/callback"
  }
}
```

Run the server application.

```sh
./goki
```

### Environment Variables

- `GOKI_CONFIG`: Path to config file. (default `./config.json`)
- `TWITTER_CONSUMER_KEY`: Twitter consumer key. (override the value loaded from `./config.json`)
- `TWITTER_CONSUMER_SECRET`: Twitter consumer secret. (override the value loaded from `./config.json`)

## Third Party Notice

### Libraries

- see `go.mod`

### Pictures

- [developer.twitter.com](https://developer.twitter.com/en/docs/basics/authentication/guides/log-in-with-twitter)
