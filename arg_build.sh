#!/usr/bin/bash

arch=$2
os=$1

echo "Building $os-$arch"
env GOOS=$os GOARCH=$arch go build -o telegraf_build/telegraf_"$os"_"$arch"_v2 -tags "goplugin" -ldflags " -s -w -X main.commit=2101b2ba -X main.branch=dotwo_battery -X main.goos=$os -X main.goarch=$arch -X main.version=1.23.4-2101b2ba" ./cmd/telegraf
# zip telegraf_build/telegraf_"$os"_"$arch"_v2.zip telegraf_build/telegraf_"$os"_"$arch"_v2
# rm telegraf_build/telegraf_"$os"_"$arch"_v2