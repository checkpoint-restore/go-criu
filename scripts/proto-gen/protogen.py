import os
import shutil
import argparse
import subprocess

parser = argparse.ArgumentParser(
    description='A script to generate Go bindings for the protobuf definitions provided by CRIU')
parser.add_argument(
    'src', help='Path to the definitions directory', type=str)
parser.add_argument(
    'dest', help='Path to the destination directory', type=str)

args = parser.parse_args()

# The import paths for each package passed to --go_opt
pkg_opts = ''
# The names of the .proto files without the extension
names = []

# Loop over the files in the src dir
for file in os.listdir(args.src):
    if file.endswith('.proto'):
        # Strip the .proto extension
        name = os.path.splitext(file)[0]
        names.append(name)
        # Add the import path for the protoc file
        pkg_opts += ',M{0}.proto=github.com/checkpoint-restore/go-criu/v7/crit/images/{0}'.format(
            name)

# Create the dest dir
if not os.path.exists(args.dest):
    os.makedirs(args.dest)
# Generate the .pb.go files
command = 'protoc -I {} --go_opt=paths=source_relative{} --go_out={} {}'.format(
    args.src, pkg_opts, args.dest, ' '.join(map(lambda s: os.path.join(args.src, s + '.proto'), names)))
result = subprocess.run(command, shell=True)

# Move the files to the respective dirs
for name in names:
    # Create dir with the same name as the file
    dir_name = os.path.join(args.dest, name)
    os.makedirs(dir_name, exist_ok=True)
    # Move the generated .pb.go file from the dest dir
    shutil.move(os.path.join(args.dest, name + '.pb.go'),
                os.path.join(dir_name, name + '.pb.go'))
