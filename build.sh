#!/bin/bash 
function build_server {
    echo "building server app"
    cd cli_server
    go generate && go build -o ../server
    cd ../
    echo "done"
}

function build_client {
    echo "building client app"
    cd gui_client
    go generate && go build -o ../client
    cd ../
    echo "done"
}  

if [ "$1" = "server" ]; then
    build_server
elif [ "$1" = "client" ]; then
    build_client
else
    build_server
    build_client
fi
