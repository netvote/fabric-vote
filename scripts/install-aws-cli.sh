#!/bin/bash

curl -O https://bootstrap.pypa.io/get-pip.py

python27 get-pip.py

pip install awscli

mkdir ~/.aws/

echo "need to populate ~/.aws/config and ~/.aws/credentials"