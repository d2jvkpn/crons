#! /usr/bin/env bash
set -eu -o pipefail
_wd=$(pwd)
_path=$(dirname $0 | xargs -i readlink -f {})

bash scripts/go_build.sh
tag=$(./target/crons -h 2>&1 | awk '/GitCommit/{print $2; exit}' | cut -c -12)
# d=crons_$(date +%F)_${tag}
d=$(printf "crons_%s_%s" $(date +%F) $tag)
mkdir -p target/${d}

mv target/{crons,crons.exe} target/${d}/
cp README*.md target/${d}/

# zip -jr ${d}.zip ${d}
cd target
zip -r ${d}.zip ${d}
rm -r ${d}/
