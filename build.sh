#!/usr/bin/bash
archs=(amd64 arm64 386 arm)
os=(linux windows)

for os in ${os[@]}
do
    for arch in ${archs[@]}
    do
        echo "Building $os-$arch"
        if [ $os = "windows" ]; then
            env GOOS=$os GOARCH=$arch go build -o finalbuild/telegraf_$os_$arch.exe -tags "goplugin" -ldflags " -X main.commit=2101b2ba -X main.branch=dotwo_battery -X main.goos=$os -X main.goarch=$arch -X main.version=1.23.4-2101b2ba" ./cmd/telegraf
        else
            env GOOS=$os GOARCH=$arch go build -o finalbuild/telegraf_$os_$arch -tags "goplugin" -ldflags " -X main.commit=2101b2ba -X main.branch=dotwo_battery -X main.goos=$os -X main.goarch=$arch -X main.version=1.23.4-2101b2ba" ./cmd/telegraf
        fi
    done
done