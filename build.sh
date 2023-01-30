#!/usr/bin/bash

archs=(amd64 arm64 386 arm)
os=(darwin linux windows)

for os in ${os[@]}
do
    for arch in ${archs[@]}
    do
        echo "Building $os-$arch"
        if [ $os = "windows" ]; then
            if [ $arch = "386" ] || [ $arch = "amd64" ]; then
                env GOOS=$os GOARCH=$arch go build -o telegraf_build/telegraf_"$os"_"$arch"_v2.exe -tags "goplugin" -ldflags " -s -w -X main.commit=2101b2ba -X main.branch=dotwo_battery -X main.goos=$os -X main.goarch=$arch -X main.version=1.23.4-2101b2ba" ./cmd/telegraf
                zip telegraf_build/telegraf_"$os"_"$arch"_v2.zip telegraf_build/telegraf_"$os"_"$arch"_v2.exe
                rm telegraf_build/telegraf_"$os"_"$arch"_v2.exe
            fi
        elif  [ $os = "darwin" ]; then
            if [ $arch = "amd64" ] || [ $arch = "arm64" ]; then
                env GOOS=$os GOARCH=$arch go build -o telegraf_build/telegraf_"$os"_"$arch"_v2 -tags "goplugin" -ldflags " -s -w -X main.commit=2101b2ba -X main.branch=dotwo_battery -X main.goos=$os -X main.goarch=$arch -X main.version=1.23.4-2101b2ba" ./cmd/telegraf
                tar -czvf telegraf_build/telegraf_"$os"_"$arch"_v2.tar.gz telegraf_build/telegraf_"$os"_"$arch"_v2
                rm telegraf_build/telegraf_"$os"_"$arch"_v2
            fi
        else
            env GOOS=$os GOARCH=$arch go build -o telegraf_build/telegraf_"$os"_"$arch"_v2 -tags "goplugin" -ldflags " -s -w -X main.commit=2101b2ba -X main.branch=dotwo_battery -X main.goos=$os -X main.goarch=$arch -X main.version=1.23.4-2101b2ba" ./cmd/telegraf
            tar -czvf telegraf_build/telegraf_"$os"_"$arch"_v2.tar.gz telegraf_build/telegraf_"$os"_"$arch"_v2
            rm telegraf_build/telegraf_"$os"_"$arch"_v2
        fi
    done
done
