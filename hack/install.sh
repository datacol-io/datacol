#!/bin/sh

set -e
PLATFORM='unknown'
unamestr=`uname`

case "$unamestr" in
  'Linux')      PLATFORM='linux';;
  'Darwin')     PLATFORM='osx';;
  *);;
esac

if [[ "$PLATFORM" == 'unknown' ]]; then
  echo "We don't provide datacol CLI for $unamestr yet. Please contact at http://datacol.io." && exit 1
fi

version=$(curl -s https://storage.googleapis.com/datacol-distros/binaries/latest.txt)
name="$PLATFORM.zip"
url="https://storage.googleapis.com/datacol-distros/binaries/$version/$name"

echo "==================================="
echo "Downloading DATACOL CLI version: $version"

tmp_path=$(mktemp -d)
response=$(curl --write-out %{http_code} --silent --output /dev/null $url)

if [[ "$response" != '200' ]]; then
  echo "Got $response with $url" && exit 1
fi

bin_path=$tmp_path/$name
curl -Ls $url > $bin_path && unzip -q $bin_path -d $tmp_path
chmod +x $tmp_path/datacol && mv $tmp_path/datacol /usr/local/bin

echo '
Datacol installed successfully!
Run with: datacol
'
