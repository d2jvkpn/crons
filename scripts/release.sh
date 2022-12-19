#! /usr/bin/env bash
set -eu -o pipefail
_wd=$(pwd)
_path=$(dirname $0 | xargs -i readlink -f {})

bash scripts/go_build.sh
d=crons_$(date +%F)_$(./target/crons -h 2>&1 | awk '/GitCommit/{print $2}' | cut -c -16)
mkdir -p release/${d}

mv target/{crons,crons.exe} release/${d}/
cp README*.md release/${d}

# zip -jr ${d}.zip ${d}
cd release
zip -r ${d}.zip ${d}
rm -r ${d}/
