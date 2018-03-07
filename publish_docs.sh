#!/bin/sh

if [[ $(git status -s) ]]
then
    echo "The working directory is dirty. Please commit any pending changes."
    exit 1;
fi

echo "Deleting old publication"
rm -rf public
mkdir public
git worktree prune
rm -rf .git/worktrees/public/

echo "Checking out gh-pages branch into public"
git worktree add -B gh-pages public origin/gh-pages

echo "Removing existing files"
rm -rf public/*

# This is a fix for git bash not being able to execute "hugo" sometimes
echo "Generating site"
if [[ $(hugo version > /dev/null 2>&1) ]]; then
  cd docs && hugo && cd ..
else
  cd docs && hugo.exe && cd ..
fi

echo "Updating gh-pages branch"
cd public && git add --all && git commit -m "[docs] publishing to gh-pages (publish_docs.sh)"
