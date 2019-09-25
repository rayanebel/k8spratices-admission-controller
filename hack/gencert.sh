#!/usr/bin/env bash

cfssl gencert -initca ca-csr.json | cfssljson -bare ../certs/ca
cfssl gencert -ca ../certs/ca.pem -ca-key ../certs/ca-key.pem -config ca-config.json --profile=server webhook-csr.json | cfssljson -bare ../certs/webhook