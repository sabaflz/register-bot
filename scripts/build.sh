#!/bin/bash

echo "Building register-bot..."
go build -o bin/register-bot main.go

if [ $? -ne 0 ]; then
    echo "Failed to build."
    exit 1
fi

echo "Built successfully: bin/register-bot"
exit 0