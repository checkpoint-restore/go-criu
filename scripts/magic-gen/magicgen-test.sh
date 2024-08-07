#!/bin/bash
# print all commands to stdout for debugging
set -x

# generate sample input file
cat > input.h <<EOF
#define TEST 0xbc614e
EOF

# generate expected output file
cat > expected.go <<EOF
// Code generated by magicgen. DO NOT EDIT.

package magic

type MagicMap struct {
	ByName  map[string]uint64
	ByValue map[uint64]string
}

func LoadMagic() MagicMap {
	magicMap := MagicMap{
		ByName:  make(map[string]uint64),
		ByValue: make(map[uint64]string),
	}
	magicMap.ByName["TEST"] = 12345678
	magicMap.ByValue[12345678] = "TEST"
	return magicMap
}
EOF

python3 magicgen.py input.h output.go
cmp output.go expected.go
if [[ $? -eq 0 ]]
then
	echo "---PASS---"
	exit 0
else
	echo "---FAIL---"
	diff output.go expected.go
	exit 1
fi
