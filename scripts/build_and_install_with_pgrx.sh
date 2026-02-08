#!/bin/bash

# fail if trying to reference a variable that is not set.
set -u

# exit immediately if a command exits with a non-zero status
set -e

PGVERSION=""
SOURCEDIR=""
INSTALL="False"
help="false"
PACKAGEDIR=""
profile=""

while getopts "d:v:ihp:r:" opt; do
  case $opt in
    d) SOURCEDIR="$OPTARG"
    ;;
    v) PGVERSION="$OPTARG"
    ;;
    i) INSTALL="True"
    ;;
    h) help="true"
    ;;
    p) PACKAGEDIR="$OPTARG"
    ;;
    r) profile="$OPTARG"
    ;;
  esac

  # Assume empty string if it's unset since we cannot reference to
  # an unset variabled due to "set -u".
  case ${OPTARG:-""} in
    -*) echo "Option $opt needs a valid argument. use -h to get help."
    exit 1
    ;;
  esac
done

if [ "$help" == "true" ]; then
    echo "Usage: $0 -d <source_directory> -v <postgres_version> [-i] [-h]"
    echo "  -d <source_directory>   : Directory containing the source code to build and install (defaults to current dir)."
    echo "  -v <postgres_version>   : Version of PostgreSQL to use (e.g., 12, 13, 14, 15)."
    echo "  -i                      : Install the built extension into PostgreSQL."
    echo "  -h                      : Display this help message."
    echo "  -p <package_directory>  : Directory to store the built package (optional)."
    echo "  -r <profile>            : Build profile to use (optional, e.g., release, debug)."
    exit 0
fi

if [ "$SOURCEDIR" == "" ]; then
    SOURCEDIR=$(pwd)
fi

if [ "$PACKAGEDIR" != "" ] && [ "$INSTALL" == "True" ]; then
    echo "Cannot specify both package directory and install option."
    exit 1
fi

if [ "$PGVERSION" == "" ]; then
    echo "Postgres version not provided. Use -v to provide Postgres version."
    exit 1
fi

source="${BASH_SOURCE[0]}"
while [[ -h $source ]]; do
   scriptroot="$( cd -P "$( dirname "$source" )" && pwd )"
   source="$(readlink "$source")"

   # if $source was a relative symlink, we need to resolve it relative to the path where the
   # symlink file was located
   [[ $source != /* ]] && source="$scriptroot/$source"
done

scriptDir="$( cd -P "$( dirname "$source" )" && pwd )"

. $scriptDir/utils.sh

pgBinDir=$(GetPostgresPath $PGVERSION)
PATH=$pgBinDir:$PATH;
pg_config_path=$pgBinDir/pg_config

# Install cargo-pgrx
pgrxInstallRequired="false"
pgrxVersionRequired=$(grep -o 'pgrx = "[^"]*"' $SOURCEDIR/Cargo.toml | cut -d '"' -f 2 | sed 's/=//')
if command -v cargo-pgrx > /dev/null; then
    pgrxVersionInstalled=$(cargo pgrx --version | awk '{print $2}')
    if [ "$pgrxVersionInstalled" != "$pgrxVersionRequired" ]; then
      pgrxInstallRequired="true"
    else
      echo "cargo-pgrx version $pgrxVersionInstalled is already installed."
    fi
else
  pgrxInstallRequired="true"
fi

if [ "$pgrxInstallRequired" == "true" ]; then
    echo "Installing cargo-pgrx..."
    cargo install --locked cargo-pgrx@${pgrxVersionRequired}
fi

cargo pgrx init --pg$PGVERSION $pg_config_path

packageProfileArg=""
installProfileArg=""
if [ "$profile" != "" ]; then
    packageProfileArg="--profile $profile"
    installProfileArg=$packageProfileArg
else
    installProfileArg="--release"
    packageProfileArg=""
fi

if [ "$INSTALL" == "True" ]; then
    pushd $SOURCEDIR
    cargo pgrx install --sudo --pg-config $pg_config_path $installProfileArg
    popd
elif [ "$PACKAGEDIR" != "" ]; then
    pushd $SOURCEDIR
    cargo pgrx package --pg-config $pg_config_path --out-dir $PACKAGEDIR $packageProfileArg --no-default-features
    popd
fi