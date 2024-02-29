#!/bin/bash


kill -9 $(lsof -i :9988) &>/dev/null | exit 0 

go run .