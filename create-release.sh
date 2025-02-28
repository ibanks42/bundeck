git tag -d $1
git push --delete origin $1
gh release delete $1
git tag $1
git push origin $1
