#!/bin/bash
set -e

REFINERY_BIN=$1
if [ -z "$REFINERY_BIN" ]; then
    echo "Usage: $0 <path-to-refinery-bin>"
    exit 1
fi

chmod +x "$REFINERY_BIN"

OS_NAME=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH_NAME=$(uname -m)

if [ "$ARCH_NAME" = "x86_64" ] || [ "$ARCH_NAME" = "amd64" ]; then
    R_ARCH="x86_64"
elif [ "$ARCH_NAME" = "aarch64" ] || [ "$ARCH_NAME" = "arm64" ]; then
    R_ARCH="aarch64"
else
    R_ARCH=$ARCH_NAME
fi

if [[ "$OS_NAME" == *"mingw"* ]] || [[ "$OS_NAME" == *"msys"* ]]; then
    OS_NAME="windows"
fi

cd tests/smoke/rust-project
rm -rf dist

"$REFINERY_BIN" build --artifact smoke-bin --os "$OS_NAME" --arch "$R_ARCH"
"$REFINERY_BIN" build --artifact smoke-lib --os "$OS_NAME" --arch "$R_ARCH"

echo "Verifying outputs..."
ls -R dist/

BIN_EXT=""
[ "$OS_NAME" = "windows" ] && BIN_EXT=".exe"

if [ -f "dist/smoke-bin$BIN_EXT" ]; then
    if [ "$OS_NAME" != "windows" ]; then
        chmod +x "dist/smoke-bin$BIN_EXT"
        ./dist/smoke-bin$BIN_EXT | grep "Refinery Smoke Test: Success"
    fi
else
    echo "Binary NOT found!" && exit 1
fi

FOUND_LIB=false
for ext in "so" "dylib" "dll" "a" "lib"; do
    if ls dist/libsmoke_lib.$ext >/dev/null 2>&1 || ls dist/smoke_lib.$ext >/dev/null 2>&1; then
        FOUND_LIB=true && break
    fi
done
[ "$FOUND_LIB" = false ] && echo "Library NOT found!" && exit 1

if ls dist/*.h >/dev/null 2>&1; then
    echo "Headers found."
else
    echo "Headers NOT found!" && exit 1
fi

ls dist/*.tar.gz dist/*.zip >/dev/null 2>&1 || (echo "Packages NOT found!" && exit 1)

echo "Smoke test passed!"
