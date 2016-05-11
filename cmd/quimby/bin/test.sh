#!/bin/bash

source test.env

rm $QUIMBY_DB
godep go install

quimby users add --username me --permission write --password hushhush
quimby users add --username him --permission read --password shhhhhhhh
quimby users add --username boss --permission admin --password sosecret

QUIMBY_TEST_SPRINKLERS_HOST="http://localhost:6111"
export QUIMBY_TEST_SPRINKLERS_HOST=$QUIMBY_TEST_SPRINKLERS_HOST

QUIMBY_TEST_SPRINKLERS_ID=$(quimby gadgets add --name sprinklers --host "http://localhost:6111")
export QUIMBY_TEST_SPRINKLERS_ID=$QUIMBY_TEST_SPRINKLERS_ID

quimby serve & SERVER=$!
gogadgets -c ./extras/sprinklers.conf & SPRINKLERS=$!

godep go test

kill $SERVER
kill $SPRINKLERS
