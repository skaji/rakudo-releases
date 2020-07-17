#!/bin/bash

set -euxo pipefail

OLD_LINES=0
if [[ -f rakudo-releases.v1.csv ]]; then
  OLD_LINES=$(wc -l < rakudo-releases.v1.csv)
fi
NEW_LINES=$(wc -l < rakudo-releases.v1.csv.new)
if [[ $OLD_LINES -ge $NEW_LINES ]]; then
  echo "There is no diff, exit"
  exit
fi

mv rakudo-releases.v1.csv.new rakudo-releases.v1.csv
git config --global user.name 'github-actions[bot]'
git config --global user.email '41898282+github-actions[bot]@users.noreply.github.com'
git add rakudo-releases.v1.csv
MESSAGE='auto update'
git commit -m "$MESSAGE"
git push https://skaji:$GITHUB_TOKEN@github.com/skaji/rakudo-releases.git master
