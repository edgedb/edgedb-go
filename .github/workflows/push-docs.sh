git config user.name edgedb-ci
git config user.email releases@edgedb.com
git add .

if ! git diff --cached --exit-code > /dev/null;
then
  git commit -m "Build rst docs from $GITHUB_SHA"
  git push
else
  echo 'No changes to built docs'
fi
