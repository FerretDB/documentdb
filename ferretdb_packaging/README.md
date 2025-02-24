# To Build Your Own Debian Packages With Docker

Run `./ferretdb_packaging/build_packages.sh -h` and follow the instructions.
E.g. to build for Debian 12 and PostgreSQL 16 for `0.102.0~ferretdb~2.0.0~rc.2` version, run:

```sh
./ferretdb_packaging/build_packages.sh --os deb12 --pg 16 --version 0.102.0~ferretdb~2.0.0~rc.2
```

Packages can be found at the `packages` directory by default,
but that can be configured with the `--output-dir` option.

**Note:** The packages do not include pg_documentdb_distributed in the `internal` directory.
