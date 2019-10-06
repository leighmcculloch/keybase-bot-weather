#! /usr/bin/env sh

set -e

echo 'Oneshot logging into keybase'
keybase oneshot
echo 'Oneshot logged into keybase'

./bot
