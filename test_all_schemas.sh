#!/bin/bash
set -e

echo "Building lens-gen..."
go build -o lens-gen ./cmd/lens-gen

mkdir -p scratch/out

echo "Testing schemas..."
schemas=(
    "address.schema.json"
    "blog-post.schema.json"
    "calendar.schema.json"
    "device.schema.json"
    "geographical-location.schema.json"
    "health-record.schema.json"
    "job-posting.schema.json"
    "movie.schema.json"
    "user-profile.schema.json"
)

for schema in "${schemas[@]}"; do
    pkgName="${schema%%.*}"
    structName="${pkgName}Lens"
    
    echo "Processing $schema..."
    ./lens-gen --pkg "$pkgName" --struct "$structName" --schema "testdata/schemas/$schema" --out "scratch/out/$pkgName.go" --auto-sanitize
    
    echo "Formatting and building $pkgName..."
    gofmt -s -w "scratch/out/$pkgName.go"
    go build -o /dev/null "scratch/out/$pkgName.go"
done

# Note: ecommerce.schema.json is a pure definition library, we test a fragment
echo "Processing ecommerce.schema.json#OrderSchema..."
./lens-gen --pkg "ecommerce" --struct "EcommerceLens" --schema "testdata/schemas/ecommerce.schema.json#OrderSchema" --out "scratch/out/ecommerce.go" --auto-sanitize
gofmt -s -w "scratch/out/ecommerce.go"
go build -o /dev/null "scratch/out/ecommerce.go"

echo "All tests passed!"
