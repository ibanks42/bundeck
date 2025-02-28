git tag -d $1 > /dev/null 2>&1
git push --delete origin $1 > /dev/null 2>&1
git tag $1
git push origin $1
