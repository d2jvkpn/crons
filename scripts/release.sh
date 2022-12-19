#! /usr/bin/env bash
set -eu -o pipefail
_wd=$(pwd)
_path=$(dirname $0 | xargs -i readlink -f {})

bash scripts/go_build.sh
d=crons_$(./target/crons -h 2>&1 | awk '/GitCommit/{print $2}' | cut -c -16)_$(date +%F)
mkdir -p target/${d}

mv target/{crons,crons.exe} target/${d}/
cp README*.md target/${d}/

# zip -jr ${d}.zip ${d}
cd target
zip -r ${d}.zip ${d}
rm -r ${d}/
