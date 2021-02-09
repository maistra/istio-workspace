#!/bin/bash

HUB_REPO_URL=git@github.com:operator-framework/community-operators.git
CURR_FOLDER=$(pwd)
TEMP_FOLDER=$(mktemp -d)
OPERATOR_NAME=istio-workspace-operator
VERSION=0.0.5

# clone
git clone $HUB_REPO_URL $TEMP_FOLDER

# add "maistra" fork

# make branch
cd $TEMP_FOLDER
git checkout -b $OPERATOR_NAME-release-$VERSION

# copy files
mkdir -p community-operators/$OPERATOR_NAME/$VERSION/
cp -R $CURR_FOLDER/bundle/ community-operators/$OPERATOR_NAME/$VERSION/

# commit - signed
git add .
git commit -s -m"Add $OPERATOR_NAME release $VERSION"


# push to fork
# open pr with template body


