#!/usr/bin/env bats

# shellcheck disable=SC2030,SC2031
# SC2030/SC2031: PATH and test paths are intentionally local to each test file.

CRIT=""
FILES_IMG=""
PRETTY_JSON=""
RAW_JSON=""

setup() {
	# BATS_TEST_DIRNAME is set by bats to the directory containing this file.
	CRIT="${BATS_TEST_DIRNAME}/../../crit/bin/crit"
	FILES_IMG="${BATS_TEST_DIRNAME}/test-imgs/inetsk/files.img"

	if [[ ! -x "$CRIT" ]]; then
		skip "crit binary not found; run 'make -C crit bin/crit'"
	fi
	if [[ ! -f "$FILES_IMG" ]]; then
		skip "checkpoint images not found; run 'make -C test/crit test-imgs'"
	fi

	PRETTY_JSON="$(mktemp)"
	RAW_JSON="$(mktemp)"
}

teardown() {
	[[ -n "$PRETTY_JSON" && -f "$PRETTY_JSON" ]] && rm -f "$PRETTY_JSON"
	[[ -n "$RAW_JSON" && -f "$RAW_JSON" ]] && rm -f "$RAW_JSON"
}

@test "crit show humanizes INETSK fields in files.img" {
	run "$CRIT" show "$FILES_IMG"
	[ "$status" -eq 0 ]
	[[ "$output" == *'"type": "INETSK"'* ]]
	[[ "$output" == *'"family": "INET"'* ]]
	[[ "$output" == *'"type": "STREAM"'* ]]
	[[ "$output" == *'"proto": "TCP"'* ]]
	[[ "$output" == *'"src_addr": ['*'"0.0.0.0"'* ]]
}

@test "crit decode --pretty humanizes INETSK fields in files.img" {
	run "$CRIT" decode -i "$FILES_IMG" --pretty
	[ "$status" -eq 0 ]
	[[ "$output" == *'"proto": "TCP"'* ]]
	[[ "$output" == *'"family": "INET"'* ]]
	[[ "$output" != *'"proto": 6'* ]]
}

@test "crit decode without --pretty keeps numeric INETSK fields" {
	run "$CRIT" decode -i "$FILES_IMG"
	[ "$status" -eq 0 ]
	[[ "$output" == *'"proto":6'* ]]
	[[ "$output" != *'"proto":"TCP"'* ]]
}

@test "humanized show proto matches crit x sk protocol field" {
	run "$CRIT" show "$FILES_IMG"
	[ "$status" -eq 0 ]
	pretty_show="$output"

	run "$CRIT" x "${BATS_TEST_DIRNAME}/test-imgs/inetsk" sk
	[ "$status" -eq 0 ]

	# files.img uses "proto"; crit x sk uses "protocol" (same helper underneath).
	[[ "$pretty_show" == *'"proto": "TCP"'* ]]
	[[ "$output" == *'"protocol": "TCP"'* ]]
}
