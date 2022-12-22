at = $(shell date +'%FT%T%:z')

build:
	echo ">>> ${at}"
	bash scripts/go_build.sh
	ls -al target/crons

crons:
	echo ">>> ${at}"
	bash scripts/go_build.sh
	target/crons -config configs/local.yaml

serve:
	echo ">>> ${at}"
	bash scripts/go_build.sh
	target/crons -config configs/local.yaml -addr :8000

release:
	echo ">>> ${at}"
	bash scripts/release.sh
