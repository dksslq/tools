#!/bin/bash

# note: The Common Name (AKA CN) represents the hostname protected by the SSL certificate.
openssl genrsa -out privkey.pem 2048
openssl req -new -x509 -key privkey.pem -out cert.pem -days 365
