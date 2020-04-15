# AppArea

AppArea is an open-source reverse proxying tool powered by SSH.

Unlike other tools out there, AppArea uses a lightweight abstraction layer
over SSH, moving as much logic as possible to the server, while keeping the
client as minimal as possible; in fact it's completely possible to just use
the default SSH command built in to most linux distributions!

## Usage

To get started, install and run the client helper script:

    $ wget https://raw.githubusercontent.com/jedevc/apparea/master/scripts/apparea.py
    $ chmod +x apparea.py
    $ ./apparea.py
    ...

Make sure that the server owner has copied your key and username over to the
server's config file!

Now you can expose ports on your local machine to the server!

    $ apparea http 8080
    >>> Listening on http://jedevc.apparea.dev
    
### Manual

Since all of the logic is handled by the server, you don't need to use the
helper script!

Once you've got your key and username in the server's config file, you can
just execute the following to expose local port 8080:

    $ ssh -R 0.0.0.0:80:localhost:8080 -p 21 jedevc@apparea.dev

## Server

Setting up the server is a bit more involved, but not that tricky.

To install it:

    $ go get github.com/jedevc/apparea
    $ go install github.com/jedevc/apparea

Then, to generate server keys and create the config file:

    $ apparea setup

To add users to the server, modify the `.apparea/authorized_keys` file as
follows:

```
ssh-<algorithm> <key> <username>
```

The format is identical to the authorized keys format followed by normal ssh
servers, however, the comment field is used to container the username.

Once you've configured your allowed users:

    $ apparea serve --bind-ssh 0.0.0.0:21 --bind-http 0.0.0.0:80

However, you probably don't want to do this, as this requires running the
program as root. For better ways of running the server see the deployment
section.

## Deployment

Deployment of the server can be tricky.

Setup on <apparea.dev> uses an nginx instance in front of a version of
apparea which is bound to localhost:8080 for http, and uses
`CAP_NET_BIND_SERVICE+ep` for listening on 0.0.0.0:21 for ssh.

You can see the configs for how the setup is done exactly in `deploy/`.
