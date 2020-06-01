#!/bin/bash

# The Common Name (AKA CN) represents the server name protected by the SSL certificate. The certificate is valid only if the request hostname matches the certificate common name.
openssl genrsa -out key.pem 2048
openssl req -new -x509 -key key.pem -out cert.pem -days 365
