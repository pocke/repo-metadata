repo-metadata
===

Put `.repo-metadata.yaml`, it updates metadata of the repository.



Usage
---

Put `.repo-metadata.yaml`

```yaml
description: Specify GitHub repository metadata from a file
homepage: https://example.com
topics:
  - golang
  - github
  - ci
```

Note: topics must not include upper case characters.


Add the following code to your .travis.yml

```yaml
after_success:
  - go get github.com/pocke/repo-metadata
  - repo-metadata
```

Get a personal access token from https://github.com/settings/tokens with `public_repo` scope.

Set `GITHUB_ACCESS_TOKEN` environment variable into Travis CI.
