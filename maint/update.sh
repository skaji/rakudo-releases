#!/bin/bash

set -euxo pipefail

FILE=$1

OLD_MD5=none
if [[ -f $FILE ]]; then
  OLD_MD5=$(md5sum < $FILE)
fi
NEW_MD5=$(md5sum < $FILE.new)
if [[ $OLD_MD5 = $NEW_MD5 ]]; then
  echo "There is no diff, exit"
  exit
fi

mv $FILE.new $FILE
git config --global user.name 'github-actions[bot]'
git config --global user.email '41898282+github-actions[bot]@users.noreply.github.com'
git add $FILE
MESSAGE='auto update'
git commit -m "$MESSAGE"
git push https://skaji:$GITHUB_TOKEN@github.com/skaji/rakudo-releases.git main
