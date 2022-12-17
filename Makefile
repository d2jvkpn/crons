build:
	bash scripts/go_build.sh

crons:
	bash scripts/go_build.sh
	target/crons -config configs/local.yaml

serve:
	bash scripts/go_build.sh
	target/crons -config configs/local.yaml -addr :8000
