---
title: 'Advanced usage'
weight: 4
summary: |
    Get more out of your AppArea usage with a number of advanced features!
---

## Using subdomains

As an AppArea user, you also can create arbitrary subdomains (as long as they
end in your username)!

Using the helper script, you can specify the `--subdomain` flag to specify a
subdomain of your username.

```bash
$ apparea http 8000 --subdomain foo
>>> Listening on http://foo-user.apparea.dev
```

## Connecting without helper script

The helper script, while useful, may not always be available, and you may
want to connect without it sometimes.

Since the script only wraps existing SSH functionality, and all the outputs
are handled by the server, you can connect without the script without any
problems.

To cast HTTP from port 8000:

```bash
$ ssh -R 0.0.0.0:80:localhost:8000 -p 21 user@apparea.dev
```

To cast HTTPS from port 8000:

```bash
$ ssh -R 0.0.0.0:443:localhost:8000 -p 21 user@apparea.dev
```

To cast TCP from port 4000:

```bash
$ ssh -R 0.0.0.0:0:localhost:4000 -p 21 user@apparea.dev
```

## Forwarding a remote host

Since the tunnel is being created by SSH remote forwarding you can also point
the tunnel at any remote service you can access.

To cast HTTP from port 8000 on the hostname `server.lan`:

```bash
$ ssh -R 0.0.0.0:80:server.lan:8000 -p 21 user@apparea.dev
```

### Caveats

Note that all requests made to `server.lan` over the tunnel will have a
Host header of `user.apparea.dev` which may cause issues if the remote
service is doing any form of hostname based routing. If this is a service you
control then you simply need to add `user.apparea.dev` as an expected
hostname.
