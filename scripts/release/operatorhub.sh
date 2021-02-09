#!/bin/bash

HUB_REPO_URL=git@github.com:operator-framework/community-operators.git
CURR_FOLDER=$(pwd)
TEMP_FOLDER?=$(mktemp)
OPERATOR_NAME?=istio-workspace
VERSION?=0.0.5

# clone

git clone $HUB_REPO_URL $TEMP_FOLDER

# make branch

cd $TEMP_FOLDER
git branch -c $OPERATOR_NAME-release

# copy files

cp -R $CURR_FOLDER/bundle/ community-operators/$OPERATOR_NAME/$VERSION/

# commit - signed



# push
# open pr with template body


