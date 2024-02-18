#!/bin/bash

echo "this will replace mbv-labs with your project name in all files"

echo "Enter your project name: "

read project_name

echo "replacing mbv-labs with $project_name"

find . -type f -name "*.go" -exec sed -i'' -e "s/MBvisti/$project_name/g" {} +
find . -type f -name "*.templ" -exec sed -i'' -e 's/mbv-labs/$project_name/g' {} +
