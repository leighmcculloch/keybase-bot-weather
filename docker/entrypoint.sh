#! /usr/bin/env sh

set -e

keybase oneshot --username "$KEYBASE_USERNAME" --paperkey "$KEYBASE_PAPERKEY"

./bot
