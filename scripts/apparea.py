#!/usr/bin/env python3
# Wrapper for apparea.

import argparse
import json
import getpass
import glob
import os
import subprocess
import sys

CONFIG_FILE = os.path.expanduser("~/.apparea.json")

SITE = None
PORT = None
USERNAME = None
KEY_FILE = None

def main():
    configure()

    parser = argparse.ArgumentParser()
    parser.add_argument("--verbose", "-v", action="store_true")
    subparsers = parser.add_subparsers()

    http_parser = subparsers.add_parser("http")
    http_parser.add_argument("port", type=int)
    http_parser.set_defaults(func=http)
    
    tcp_parser = subparsers.add_parser("tcp")
    tcp_parser.add_argument("ports", nargs="+", type=int)
    tcp_parser.set_defaults(func=tcp)

    args = parser.parse_args()

    if hasattr(args, "func"):
        args.func(args)
    else:
        parser.print_usage()

def http(args):
    forwards = ["-R", f"0.0.0.0:80:localhost:{args.port}"]
    command = [*forwards, "-p", str(PORT), f"{USERNAME}@{SITE}"]
    if args.verbose:
        command.append("-v")

    run_ssh(command)

def tcp(args):
    forwards = [("-R", f"0.0.0.0:0:localhost:{port}") for port in args.ports]
    forwards = [item for forward in forwards for item in forward]
    command = [*forwards, "-p", str(PORT), f"{USERNAME}@{SITE}"]
    if args.verbose:
        command.append("-v")

    run_ssh(command)

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
    except KeyboardInterrupt:
        proc.terminate()

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