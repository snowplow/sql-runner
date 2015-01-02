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

# Zip the artifact (assumes godep go build already run)
hyphenated_package=`echo ${bintray_package}|tr '-' '_'`
artifact="${hyphenated_package}_${version}_linux_amd64.zip"
rm -f ${artifact}
zip ${artifact} ${bintray_package}

# Create the version (does nothing if already exists)
echo '{"name":"'${version}'","desc":"Release of '${bintray_package}'"}' | curl -d @- \
"https://api.bintray.com/packages/${bintray_repository}/${bintray_package}/versions" \
--header "Content-Type:application/json" \
-u${bintray_user}:${bintray_api_key} \

# Upload the artifact (overwrites if already exists)
curl -T ${artifact} \
"https://api.bintray.com/content/${bintray_repository}/${bintray_package}/${version}/${artifact}?publish=1&override=1" \
-u${bintray_user}:${bintray_api_key} \
