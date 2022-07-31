#!/bin/bash

set -x

CRIT=../../crit/bin/crit

function gen_img_list {
	images_list=$(ls -1 ./*.img)
	if [ -z "$images_list" ]; then
		echo "Failed to generate images"
		exit 1
	fi
}

function run_test1 {
	for x in $images_list
	do
		echo "=== $x"
		if [[ $x == *pages* ]]; then
			echo "skip"
			continue
		fi

		echo "  -- to json"
		$CRIT decode -i "$x" -o "$x"".json" --pretty || exit $?
		echo "  -- to img"
		$CRIT encode -i "$x"".json" -o "$x"".json.img" || exit $?
		echo "  -- cmp"
		cmp "$x" "$x"".json.img" || exit $?

		echo "=== done"
	done
}


function run_test2 {
	PROTO_IN=./inventory.img
	JSON_IN=$(mktemp -p ./ tmp.XXXXXXXXXX.json)
	OUT=$(mktemp -p ./ tmp.XXXXXXXXXX.img)

	# prepare
	${CRIT} decode -i "${PROTO_IN}" -o "${JSON_IN}"

	# proto in - json out decode
	cat "${PROTO_IN}" | ${CRIT} decode || exit 1
	cat "${PROTO_IN}" | ${CRIT} decode -o "${OUT}" || exit 1
	cat "${PROTO_IN}" | ${CRIT} decode > "${OUT}" || exit 1
	${CRIT} decode -i "${PROTO_IN}" || exit 1
	${CRIT} decode -i "${PROTO_IN}" -o "${OUT}" || exit 1
	${CRIT} decode -i "${PROTO_IN}" > "${OUT}" || exit 1
	${CRIT} decode < "${PROTO_IN}" || exit 1
	${CRIT} decode -o "${OUT}" < "${PROTO_IN}" || exit 1
	${CRIT} decode < "${PROTO_IN}" > "${OUT}" || exit 1

	# proto in - json out encode -> should fail
	cat "${PROTO_IN}" | ${CRIT} encode || true
	cat "${PROTO_IN}" | ${CRIT} encode -o "${OUT}" || true
	cat "${PROTO_IN}" | ${CRIT} encode > "${OUT}" || true
	${CRIT} encode -i "${PROTO_IN}" || true
	${CRIT} encode -i "${PROTO_IN}" -o "${OUT}" || true
	${CRIT} encode -i "${PROTO_IN}" > "${OUT}" || true

	# json in - proto out encode
	cat "${JSON_IN}" | ${CRIT} encode || exit 1
	cat "${JSON_IN}" | ${CRIT} encode -o "${OUT}" || exit 1
	cat "${JSON_IN}" | ${CRIT} encode > "${OUT}" || exit 1
	${CRIT} encode -i "${JSON_IN}" || exit 1
	${CRIT} encode -i "${JSON_IN}" -o "${OUT}" || exit 1
	${CRIT} encode -i "${JSON_IN}" > "${OUT}" || exit 1
	${CRIT} encode < "${JSON_IN}" || exit 1
	${CRIT} encode -o "${OUT}" < "${JSON_IN}" || exit 1
	${CRIT} encode < "${JSON_IN}" > "${OUT}" || exit 1

	# json in - proto out decode -> should fail
	cat "${JSON_IN}" | ${CRIT} decode || true
	cat "${JSON_IN}" | ${CRIT} decode -o "${OUT}" || true
	cat "${JSON_IN}" | ${CRIT} decode > "${OUT}" || true
	${CRIT} decode -i "${JSON_IN}" || true
	${CRIT} decode -i "${JSON_IN}" -o "${OUT}" || true
	${CRIT} decode -i "${JSON_IN}" > "${OUT}" || true

	# explore image directory
	${CRIT} x ./ ps || exit 1
	${CRIT} x ./ fds || exit 1
	${CRIT} x ./ mems || exit 1
	${CRIT} x ./ rss || exit 1
}

echo "Generating image list..."
gen_img_list
echo "Testing recode..."
run_test1
echo "Testing commands..."
run_test2
