# Creating a new release

1. Update the version string in `version.txt`.
2. Update the version in `quickstart.yaml`.
3. Add new release notes to `CHANGELOG.md`.
4. Commit changes, push, and open a release preparation pull request for review
5. Once the pull request is approved and merged, fetch the updated `main` branch
6. Add a tag to the `main` branch with the new version in the following format: `v1.2.3`.
7. Push the new version tag up to the project repository to kick off the CI workflow, which will package and publish the image to Docker Hub.
8. Update the Draft Release with proper release notes (copied from CHANGELOG or auto-generated).
