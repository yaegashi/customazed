#!/bin/bash

set -ex

apt-get update
apt-get install -y build-essential
apt-get dist-upgrade -y
apt-get autoremove -y --purge
apt-get clean
