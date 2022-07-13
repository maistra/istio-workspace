#!/bin/bash

echo "Setting up wg0"
wireguard-go wg0

while [ 1 ]
do
   sleep 5
done