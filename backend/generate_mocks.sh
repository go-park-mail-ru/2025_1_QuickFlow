#!/bin/bash

BASE_DIR="./"

generate_mock() {
    local source_file="$1"
    local mock_file="$2"
    echo "Generating mock for: $source_file"
    go run github.com/golang/mock/mockgen -source="$source_file" -destination="$mock_file" -package=mocks
}

find "$BASE_DIR" -type f -name "*.go" | while read -r file; do
    if grep -E "type[[:space:]]+[[:alnum:]_]+[[:space:]]+interface[[:space:]]*{" "$file" >/dev/null; then
        mock_dir="$(dirname "$file")/mocks"
        mkdir -p "$mock_dir"

        mock_file="$mock_dir/$(basename "$file" .go)_mock.go"
        generate_mock "$file" "$mock_file"
    fi
done