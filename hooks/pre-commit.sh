#!/bin/bash

echo "Running pre-commit checks..."

echo "Running cargo fmt --check..."
if ! cargo fmt --check; then
    echo "Formatting check failed. Please run 'cargo fmt' to fix."
    exit 1
fi
echo "Formatting check passed."

echo "Running cargo clippy --all-targets --all-features -- -D warnings..."
if ! cargo clippy --all-targets --all-features -- -D warnings; then
    echo "Clippy check failed. Please fix the reported errors."
    exit 1
fi
echo "Clippy check passed."

echo "Running cargo build..."
if ! cargo build; then
    echo "Build failed. Please fix compilation errors."
    exit 1
fi
echo "Build passed."

echo "Running sweet..."
if ! swt --quiet; then
    echo "'swt' command failed. Please fix the issues."
    exit 1
fi
echo "'swt' command passed."

echo "All pre-commit checks passed. Proceeding with commit."
exit 0
