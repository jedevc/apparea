---
title: 'Protocols'
weight: 3
summary: |
    Explore the different forwarding protocols that AppArea supports.
---

AppArea supports forwarding a number of different protocols.

## HTTP

HTTP forwarding is the recommended way to use AppArea, wherever possible.

By using it, you get an automatic domain name for your site and automatic
HTTPS exposed to the end-user (depending on server configuration).

To cast an HTTP port:

```bash
> apparea http 8000
>>> Listening on http://user.apparea.dev
```

### HTTP server

A common use case for AppArea is to serve a static directory, either for
sharing files, or for serving a static site.

The client helper includes a handy subcommand for serving the current
directory and automatically proxying it to AppArea.

```bash
> apparea serve-http
>>> Listening on http://user.apparea.dev
```

## HTTPS

HTTPS forwarding is a variation on HTTP forwarding, and is included for
compatibility reasons.

To cast an HTTPS port:

```bash
> apparea https 8000
>>> Listening on http://user.apparea.dev
```

The AppArea server decrypts the HTTPS on the server, and then serves it over
HTTP and HTTPS (depending on server configuration). It does not directly
forward it to the client for 2 reasons:

1. To provide logging data to the ssh session
2. The domain will probably be wrong, giving cert errors

Using HTTPS forwarding over HTTP forwarding does **not** provide any
additional security benefits. This is because, on connection to the server,
your HTTP is tunnelled through the SSH protocol, and so is already secure.
Then the server decrypts the session anyways.

If you want to forward the HTTPS directly to the client and entirely avoid
the server decrypting it at all, and preserving maximum security, you can
still forward the underlying TCP connection (as seen below).

## TCP

TCP forwarding is the lowest (and most primitive) layer of forwarding and can
(in theory) forward all protocols built on top of TCP.

To cast a TCP port:

```bash
> apparea tcp 8000
>>> Listening on user.apparea.dev:?????
```

Note that the port given does not just apply to `user.apparea.dev` but to all
subdomains of `apparea.dev` including `apparea.dev` itself. The subdomain
provided is simply a utility.
