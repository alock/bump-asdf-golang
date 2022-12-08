#!/bin/bash

if [ -z "$1" ]
  then
      echo "Please pass golang semver (ex: ./golang.sh 1.18.5)"
    exit
fi

version="$1"
if [[ ! ${version} =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    printf >&2 'Error: %s is not a valid semver.\n' $version
    exit 1
fi

printf "semver passed: $version\n"
IFS=. read -r major minor patch <<EOF
$version
EOF

previous_version="$major.$minor.$(( $patch -1))"

workspaces_found=$(rg --fixed-strings "$major.$minor." $(fd -IH workspace.xml) -l | wc -l | xargs)
printf "found $workspaces_found jetbrains workspace files with $major.$minor in it\n"
printf "manipulate the workspace files\n"
for file in $(rg --fixed-strings "$previous_version" $(fd -IH workspace.xml) -l); do sed -i '' "s/$previous_version/$version/g" $file; done
workspaces_fixed=$(rg --fixed-strings "$version" $(fd -IH workspace.xml) -l | wc -l | xargs)
printf "after changes $workspaces_fixed workspace files with $version in it\n"


tool_versions_found=$(rg --fixed-strings "$major.$minor." $(fd -IH tool-versions -E vendor -E node_modules) -l | wc -l | xargs)
printf "found $tool_versions_found .tool-versions files with $major.$minor in it\n"
printf "manipulate the .tool-versions files\n"
for file in $(rg --fixed-strings "$previous_version" $(fd -IH tool-versions -E vendor -E node_modules) -l); do sed -i '' "s/$previous_version/$version/g" $file; done
tool_versions_fixed=$(rg --fixed-strings "$version" $(fd -IH tool-versions -E vendor -E node_modules) -l | wc -l | xargs)
printf "after changes $tool_versions_fixed .tool-versions files with $version in it\n"
