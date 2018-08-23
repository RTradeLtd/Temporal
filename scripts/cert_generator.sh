#! /bin/bash

# used to generate certificates for testing

sudo openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout "$HOME/api.key" -out "$HOME/api.crt"