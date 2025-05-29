# FerretDB's DocumentDB Release Checklist

## Preparation

1. Create draft release on GitHub to see a list of merged PRs.
2. Update CHANGELOG.md manually. It will point to versions of DocumentDB and FerretDB that are not released yet.
3. Update `packaging/debian_files/changelog`.
4. Send PR with changes, merge it.

## Git tag

1. Make a signed tag with `git tag -s --cleanup=verbatim vX.Y.Z-ferretdb-A.B.C(-p)` (like `v0.103.0-ferretdb-2.2.0-beta.1`),
   where `X.Y.Z` is the SemVar formatted version of DocumentDB (like `0.103.0`),
   and `A.B.C(-p)` is the compatible FerretDB version (like `2.2.0-beta.1`).
2. Check `git status` output.
3. Push it!

## Release

1. Find [Packages CI build](https://github.com/FerretDB/documentdb/actions/workflows/ferretdb_packages.yml?query=event%3Apush)
   for the tag to release.
2. Check Docker images.
3. Upload `.deb` packages to the draft release.
4. Update release notes with the content of CHANGELOG.md.
5. Publish release on GitHub.
