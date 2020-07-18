#!/bin/bash

set -euxo pipefail

FILE=$1

OLD_LINES=0
if [[ -f $FILE ]]; then
  OLD_LINES=$(wc -l < $FILE)
fi
NEW_LINES=$(wc -l < $FILE.new)
if [[ $OLD_LINES -ge $NEW_LINES ]]; then
  echo "There is no diff, exit"
  exit
fi

mv $FILE.new $FILE
git config --global user.name 'github-actions[bot]'
git config --global user.email '41898282+github-actions[bot]@users.noreply.github.com'
git add $FILE
MESSAGE='auto update'
git commit -m "$MESSAGE"
git push https://skaji:$GITHUB_TOKEN@github.com/skaji/rakudo-releases.git master
