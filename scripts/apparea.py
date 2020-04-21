#!/usr/bin/env python3
# Wrapper for apparea.

import argparse
import getpass
import glob
import json
import subprocess
import sys
import time
import os

CONFIG_FILE = os.path.expanduser("~/.apparea.json")

SITE = None
PORT = None
USERNAME = None
KEY_FILE = None

def main():
    configure()

    parser = argparse.ArgumentParser(description="Client helper script to forward ports using AppArea")
    parser.add_argument("--verbose", "-v", action="store_true", help="enable verbose output")
    subparsers = parser.add_subparsers()

    http_parser = subparsers.add_parser("http", help="proxy a http port")
    http_parser.add_argument("port", type=int, help="target port to proxy")
    http_parser.add_argument("--subdomain", "-s", help="target domain to proxy to")
    http_parser.set_defaults(func=http)

    http_parser = subparsers.add_parser("https", help="proxy a https port")
    http_parser.add_argument("port", type=int, help="target port to proxy")
    http_parser.add_argument("--subdomain", "-s", help="target domain to proxy to")
    http_parser.set_defaults(func=https)
    
    tcp_parser = subparsers.add_parser("tcp", help="proxy a raw tcp port")
    tcp_parser.add_argument("ports", nargs="+", type=int, help="target ports to proxy")
    tcp_parser.set_defaults(func=tcp)

    args = parser.parse_args()

    if hasattr(args, "func"):
        args.func(args)
    else:
        parser.print_usage()

def exponential_backoff(f):
    def wrapper(*args, **kwargs):
        delay = 1
        while True:
            start = time.time()
            res = f(*args, **kwargs)
            end = time.time()

            if res:
                break

            if end - start > 2:
                # HACK: assume that if the process was running for longer than
                # 2 seconds, it successfully established a connection
                delay = 1

            time.sleep(delay)
            delay *= 2
            if delay > 60:
                delay = 60
    return wrapper

def http(args):
    username = craft_username(args.subdomain)
    forward(80, [args.port], username=username, verbose=args.verbose)

def https(args):
    username = craft_username(args.subdomain)
    forward(443, [args.port], username=username, verbose=args.verbose)

def tcp(args):
    forward(0, args.ports, verbose=args.verbose)

def craft_username(subdomain):
    if subdomain:
        username = subdomain.split('.') + [USERNAME]
        username.reverse()
        username = ".".join(username)
    else:
        username = USERNAME
    
    return username

def forward(dest, srcs, username=None, verbose=False):
    if username is None:
        username = USERNAME

    forwards = [("-R", f"0.0.0.0:{dest}:localhost:{src}") for src in srcs]
    forwards = [item for forward in forwards for item in forward]
    command = [*forwards, "-T", "-i", KEY_FILE, "-p", str(PORT), f"{username}@{SITE}"]
    if verbose:
        command.append("-v")

    run_ssh(command)

@exponential_backoff
def run_ssh(args):
    proc = subprocess.Popen(
            ["ssh", *args],
            stdout=subprocess.PIPE,
            stderr=subprocess.STDOUT,
            stdin=subprocess.PIPE)
    try:
        while True:
            line = proc.stdout.readline()
            if not line:
                break

            if b"want to continue connecting" in line:
                proc.communicate(input())

            sys.stdout.buffer.write(line)
            sys.stdout.buffer.flush()

        proc.wait()
        return False
    except KeyboardInterrupt:
        proc.terminate()
        return True

def configure():
    global SITE
    global PORT
    global USERNAME
    global KEY_FILE

    try:
        with open(CONFIG_FILE) as f:
            config = json.load(f)

            SITE = config["site"]
            PORT = config["port"]
            USERNAME = config["username"]
            KEY_FILE = config["keyfile"]
    except FileNotFoundError:
        print("Welcome to apparea!")
        print("Since this is your first time, this helper will get you setup.\n")

        site = "apparea.dev"
        new_site = input(f"Site [{site}]: ")
        if new_site:
            site = new_site

        port = 21
        new_port = input(f"Port [{port}]: ")
        if new_port:
            port = int(new_port)

        username = getpass.getuser()
        new_username = input(f"Username [{username}]: ")
        if new_username:
            username = new_username

        keyfiles = glob.iglob(os.path.abspath(os.path.expanduser("~/.ssh/id_*")))
        keyfiles = list(filter(lambda s: ".pub" not in s, keyfiles))
        keyfile = keyfiles[0]

        new_keyfile = input(f"SSH Key [{keyfile}]: ")
        if new_keyfile:
            keyfile = os.path.abspath(os.path.expanduser(new_keyfile))

        print()
        result = json.dumps({
            "site": site,
            "port": port,
            "username": username,
            "keyfile": keyfile,
        }, indent=4)
        print(result)
        ok = input("Is this ok? [yes]/no: ")
        if ok and ok[0].lower() == 'n':
            print("Alright, quitting.")
            sys.exit(1)

        with open(CONFIG_FILE, "w") as f:
            f.write(result)
        print(f"Written config to {CONFIG_FILE}")

        print()
        install = input("Do you want to install this script to /usr/local/bin? [yes]/no: ")
        if not install or install[0].lower() != 'n':
            command = f"sudo cp {os.path.realpath(__file__)} /usr/local/bin/apparea"
            print("$ " + command)
            subprocess.run(command, shell=True)

        print()
        print("All done!")
        print()

        SITE = site
        PORT = port
        KEY_FILE = keyfile
        USERNAME = username

if __name__ == "__main__":
    main()