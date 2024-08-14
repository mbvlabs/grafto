#!/bin/bash

echo "this will replace mbv-labs with your gh_profile and project name in all files"

echo "Enter your gh_profile name: "

read gh_profile_name

echo "replacing mbv-labs with $gh_profile_name"

find . -type f -name "*.go" -exec sed -i'' -e "s/mbvlabs/$gh_profile_name/g" {} +
find . -type f -name "go.mod" -exec sed -i'' -e "s/mbvlabs/$gh_profile_name/g" {} +
find . -type f -name "*.templ" -exec sed -i'' -e "s/mbvlabs/$gh_profile_name/g" {} +

echo "Enter your project name: "

read project_name

echo "replacing grafto with $project_name"

find . -type f -name "*.go" -exec sed -i'' -e "s/grafto/$project_name/g" {} +
find . -type f -name "go.mod" -exec sed -i'' -e "s/grafto/$project_name/g" {} +
find . -type f -name "*.templ" -exec sed -i'' -e "s/grafto/$project_name/g" {} +

echo "deleting rename.sh"

rm rename.sh
