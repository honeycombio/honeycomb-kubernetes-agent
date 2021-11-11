# Creating a new release

1. Update the version string in `version.txt`. Add new release notes to `CHANGELOG.md`. 
2. Open a PR with those changes, and await for it to be approved, then merge it.
3. Create a new Release with proper release notes. Add a tag to the release in the following format: `v1.2.3`.
4. This will kick off a CI workflow, which will package and publish the image to Docker Hub.
