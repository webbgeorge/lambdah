# Contributing to lambdah

Thank you for taking the time to contribute.

## Setting up the project

### Forking the repository

Fork this repository and clone your fork to your workspace. Note that this project
uses go modules for dependencies, so does not need to be cloned into your GOPATH.

If you haven't worked with forked Go repos before, take a look at this blog post
for some excellent advice about
[contributing to go open source git repositories](https://splice.com/blog/contributing-open-source-git-repositories-go/).

### Installing dependencies

`go mod download`

### Running tests

The tests for this repository run `go fmt`, `go vet` and `go test`, all included in go.
However, it also uses 2 tools for security scanning: `gosec` and `Nancy` - to run the
tests you will need these installed, which you can do by running:

```
make install-dev-dependencies-mac
OR
make install-dev-dependencies-linux
```

Once those are installed you can run full tests for repository:

```
make test
```

## Pull requests

- Please open PRs against master.
- We prefer single commit PRs
- Please add/modify tests to cover the proposed code changes.
- If the PR contains a new feature, please document it in the README.

## Documentation

For simple typo fixes and documentation improvements feel free to raise
a PR without raising an issue in github. For anything more complicated
please file an issue.
