# AppArea

AppArea is an open-source reverse proxying tool powered by SSH.

Unlike other tools out there, AppArea uses a lightweight abstraction layer
over SSH, moving as much logic as possible to the server, while keeping the
client as minimal as possible; in fact it's completely possible to just use
the default SSH command built in to most linux distributions!

## Installation

Clone the repository:

    $ git clone --recursive https://github.com/jedevc/apparea.git
    $ cd apparea

Then you need to configure everything:

1. Create `.env` using `.example.env` as a template.
2. Create `docker-compose.override.yml` to configure ACME DNS challenge
   environment variables for the traefik service.
3. Then, create the required server config files:

    ```
    $ mkdir -p config
    $ ssh-keygen -N "" -f ./config/id_rsa -t rsa -b 4096
    $ touch config/authorized_keys
    ```

    You should have the following files:

    ```
    $ tree config
    config
    ├── authorized_keys
    ├── id_rsa
    └── id_rsa.pub

    0 directories, 3 files
    ```

4. Then you can bring the containers up:

    ```
    $ docker-compose up
    ```

## Configuration format

The format is identical to the authorized keys format followed by normal SSH
servers, however, the comment field is used to contain the username.

```
ssh-<algorithm> <key> <username>
```

## Usage

To get started, install and run the client helper script:

    $ wget https://raw.githubusercontent.com/jedevc/apparea/master/scripts/apparea.py
    $ chmod +x apparea.py
    $ ./apparea.py
    ...

Make sure that you copy your key and username over to config/authorized_keys
and restart the server.

Now you can expose ports on your local machine to the server!

    $ apparea http 8080
    >>> Listening on http://jedevc.apparea.dev
    
### Manual

Since all of the logic is handled by the server, you don't need to use the
helper script!

Once you've got your key and username in the server's config file, you can
just execute the following to expose local port 8080:

    $ ssh -R 0.0.0.0:80:localhost:8080 -p 21 jedevc@apparea.dev
    >>> Listening on http://jedevc.apparea.dev
