git tag -d $1
gh release delete -y --cleanup-tag $1
git tag $1
git push origin $1
