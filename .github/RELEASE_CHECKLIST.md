# FerretDB's DocumentDB Release Checklist

## Preparation

1. Create draft release on GitHub to see a list of merged PRs.
2. Update CHANGELOG.md manually.
3. Push changes.

## Git tag

1. Make a signed tag `vX.Y.Z-ferretdb-A.B.C(-p)` (like `v0.102.0-ferretdb-2.0.0-rc.2`),
   where `X.Y.Z` is the SemVar formatted version of DocumentDB (like `0.102.0`),
   and `A.B.C(-p)` is the compatible FerretDB version (like `2.0.0-rc.2`).
2. Check `git status` output.
3. Push it!

## Release

1. Find [Packages CI build](https://github.com/FerretDB/documentdb/actions/workflows/ferretdb_packages.yml?query=event%3Apush)
   for the tag to release.
2. Upload `.deb` packages to the draft release.
3. Update release notes with the list of changes from CHANGELOG.md.
4. Publish release on GitHub.
