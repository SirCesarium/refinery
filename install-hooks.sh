#!/bin/bash

if ! command -v go >/dev/null 2>&1; then
    echo "Warning: 'go' is not installed. Unit tests and formatting checks will fail."
fi

if ! command -v docker >/dev/null 2>&1; then
    echo "Warning: 'docker' is not installed. Smoke tests will fail."
fi

if ! command -v make >/dev/null 2>&1; then
    echo "Warning: 'make' is not installed. Hooks depend on Makefile targets."
fi

mkdir -p .git/hooks

cp hooks/pre-commit .git/hooks/pre-commit
chmod +x .git/hooks/pre-commit

cp hooks/pre-push .git/hooks/pre-push
chmod +x .git/hooks/pre-push

echo "Hooks installed successfully."
