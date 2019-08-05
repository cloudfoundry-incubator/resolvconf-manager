#!/bin/bash

set -eux
set -o pipefail

mkdir -p /dir && touch /dir/resolv.conf

mount -o bind /dir/resolv.conf /etc/resolv.conf
