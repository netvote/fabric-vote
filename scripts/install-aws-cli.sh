#!/bin/bash


yum -y install unzip

curl -O https://bootstrap.pypa.io/get-pip.py

python27 get-pip.py

mkdir ~/.aws/

echo "need to populate ~/.aws/config and ~/.aws/credentials"