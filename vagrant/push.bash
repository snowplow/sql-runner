#!/bin/bash
set -e

bintray_package=sql-runner
bintray_user=snowplowbot
bintray_repository=snowplow/snowplow-generic

# Can't pass args thru vagrant push so have to prompt
read -e -p "Please enter API key for Bintray user ${bintray_user}: " bintray_api_key

# Get the parent directory of where this script is
source="${BASH_SOURCE[0]}"
while [ -h "${source}" ] ; do source="$(readlink "${source}")"; done
dir="$( cd -P "$( dirname "${source}" )/.." && pwd )"
cd ${dir}

# Version is stored in a file
version=`cat VERSION`

# Build Windows, OSX and Linux variations
build_dir="/opt/gopath/src/github.com/snowplow/sql-runner"
dist_dir="/vagrant/dist"
build_cmd="godep go build"
goarch_64="amd64"

# 1: Windows
echo "============================="
echo "  Building Windows artifact  "
echo "-----------------------------"
goos_win="windows"
vagrant ssh -c "cd ${build_dir} && export GOOS=${goos_win} && export GOARCH=${goarch_64} && ${build_cmd} -o ${dist_dir}/${bintray_package}.exe"
artifact_win="sql_runner_${version}_${goos_win}_${goarch_64}.zip"
zip dist/${artifact_win} dist/${bintray_package}.exe

# 2: Mac OSX
echo "========================="
echo "  Building OSX artifact  "
echo "-------------------------"
goos_osx="darwin"
vagrant ssh -c "cd ${build_dir} export GOOS=${goos_osx} && export GOARCH=${goarch_64} && ${build_cmd} -o ${dist_dir}/${bintray_package}"
artifact_osx="sql_runner_${version}_${goos_osx}_${goarch_64}.zip"
zip dist/${artifact_osx} dist/${bintray_package}

# 3: Linux
echo "==========================="
echo "  Building Linux artifact  "
echo "---------------------------"
goos_linux="linux"
vagrant ssh -c "cd ${build_dir} export GOOS=${goos_linux} && export GOARCH=${goarch_64} && ${build_cmd} -o ${dist_dir}/${bintray_package}"
artifact_linux="sql_runner_${version}_${goos_linux}_${goarch_64}.zip"
zip dist/${artifact_linux} dist/${bintray_package}

echo "=================================="
echo "  Uploading artifacts to bintray  "
echo "----------------------------------"

# Create the version (does nothing if already exists)
echo '{"name":"'${version}'","desc":"Release of '${bintray_package}'"}' | curl -d @- \
"https://api.bintray.com/packages/${bintray_repository}/${bintray_package}/versions" \
--header "Content-Type:application/json" \
-u${bintray_user}:${bintray_api_key} \

# Upload the artifact (overwrites if already exists)
curl -T dist/${artifact_win} \
"https://api.bintray.com/content/${bintray_repository}/${bintray_package}/${version}/${artifact_win}?publish=1&override=1" \
-u${bintray_user}:${bintray_api_key} \

curl -T dist/${artifact_osx} \
"https://api.bintray.com/content/${bintray_repository}/${bintray_package}/${version}/${artifact_osx}?publish=1&override=1" \
-u${bintray_user}:${bintray_api_key} \

curl -T dist/${artifact_linux} \
"https://api.bintray.com/content/${bintray_repository}/${bintray_package}/${version}/${artifact_linux}?publish=1&override=1" \
-u${bintray_user}:${bintray_api_key} \
