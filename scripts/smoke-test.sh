#!/bin/bash
set -e

REFINERY_BIN=$(realpath "$1")
FILTER_OS="$2"
[ -f "$REFINERY_BIN" ] || (echo "Refinery bin not found" && exit 1)

cd tests/smoke/rust-project
rm -rf dist logs
mkdir -p logs

declare -A TARGETS
TARGETS["linux"]="x86_64:gnu x86_64:musl i686:gnu aarch64:gnu"
if [[ "$(uname -s)" == "MINGW"* || "$(uname -s)" == "CYGWIN"* || "$(uname -s)" == "MSYS"* ]]; then
    TARGETS["windows"]="x86_64:msvc i686:msvc"
else
    TARGETS["windows"]="x86_64:gnu i686:gnu"
fi
TARGETS["wasi"]="wasm32"

PIDS=()
LOGS=()
NAMES=()

for os in "${!TARGETS[@]}"; do
    if [ -n "$FILTER_OS" ] && [[ "$FILTER_OS" != *"$os"* ]]; then
        continue
    fi
    for arch_abi in ${TARGETS[$os]}; do
        arch=$(echo $arch_abi | cut -d: -f1)
        abi=$(echo $arch_abi | cut -d: -f2 -s)
        
        EXTRA_RUSTFLAGS=""
        [ "$os" = "windows" ] && EXTRA_RUSTFLAGS="-C link-args=-static"
        
        # Build bin
        name_bin="bin-$os-$arch-$abi"
        args="--artifact smoke-bin --os $os --arch $arch"
        [ -n "$abi" ] && args="$args --abi $abi"
        RUSTFLAGS="$EXTRA_RUSTFLAGS" "$REFINERY_BIN" build $args > "logs/$name_bin.log" 2>&1 &
        PIDS+=($!)
        LOGS+=("logs/$name_bin.log")
        NAMES+=("$name_bin")
        
        # Build lib
        name_lib="lib-$os-$arch-$abi"
        args_lib="--artifact smoke-lib --os $os --arch $arch"
        [ -n "$abi" ] && args_lib="$args_lib --abi $abi"
        RUSTFLAGS="$EXTRA_RUSTFLAGS" "$REFINERY_BIN" build $args_lib > "logs/$name_lib.log" 2>&1 &
        PIDS+=($!)
        LOGS+=("logs/$name_lib.log")
        NAMES+=("$name_lib")
    done
done

echo "Waiting for builds (${#PIDS[@]}) to complete..."
FAILED=0
for i in "${!PIDS[@]}"; do
    if ! wait ${PIDS[$i]}; then
        echo "FAILED: ${NAMES[$i]}"
        cat "${LOGS[$i]}"
        FAILED=1
    fi
done

[ $FAILED -eq 1 ] && exit 1

# 1. Verify Header Content
HEADER="dist/smoke-test.h"
grep -q "smoke_test_fn" "$HEADER" && echo "Header content: VALID"

# 2. Runtime Verification
check_exec() {
    local bin=$1
    local cmd=$2
    local full_path=$(realpath "dist/$bin")
    echo -n "Testing execution of $bin... "
    if [ -f "$full_path" ]; then
        local output
        output=$(LC_ALL=C $cmd "$full_path" 2>&1) || {
            echo "FAILED (Runtime error: exit code $?)"
            echo "$output"
            exit 1
        }
        echo "$output" | grep -q "Refinery Smoke Test: Success" && echo "PASSED" || (echo "FAILED (Output mismatch)" && exit 1)
    else
        echo "FAILED (Missing binary)" && exit 1
    fi
}

# Only run tests for OSes that were built (respect FILTER_OS)
if [ -z "$FILTER_OS" ] || [[ "$FILTER_OS" == *"linux"* ]]; then
    check_exec "smoke-bin-linux-x86_64-gnu" ""
    [ -x "$(command -v qemu-i386)" ] && check_exec "smoke-bin-linux-i686-gnu" "qemu-i386"
    [ -x "$(command -v qemu-aarch64)" ] && check_exec "smoke-bin-linux-aarch64-gnu" "qemu-aarch64 -L /usr/aarch64-linux-gnu"
fi

if [ -z "$FILTER_OS" ] || [[ "$FILTER_OS" == *"wasi"* ]]; then
    [ -x "$(command -v wasmtime)" ] && check_exec "smoke-bin-wasi-wasm32.wasm" "wasmtime"
fi

if [ -z "$FILTER_OS" ] || [[ "$FILTER_OS" == *"windows"* ]]; then
    [ -f "dist/smoke-bin-windows-x86_64-msvc.exe" ] && echo "smoke-bin-windows-x86_64-msvc.exe: OK"
    [ -f "dist/smoke-bin-windows-i686-msvc.exe" ] && echo "smoke-bin-windows-i686-msvc.exe: OK"
fi

# 3. Verify existence
if [ -z "$FILTER_OS" ] || [[ "$FILTER_OS" == *"linux"* ]]; then
    [ -f "dist/smoke-bin-0.1.0-linux-x86_64-gnu.deb" ] && echo ".deb: OK"
    [ -f "dist/smoke-bin-0.1.0-linux-x86_64-gnu.rpm" ] && echo ".rpm: OK"
fi

echo "All smoke tests PASSED!"
