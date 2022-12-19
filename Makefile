build:
	bash scripts/go_build.sh
	ls -al target/crons

crons:
	bash scripts/go_build.sh
	target/crons -config configs/local.yaml

serve:
	bash scripts/go_build.sh
	target/crons -config configs/local.yaml -addr :8000

pack:
	bash scripts/release.sh
