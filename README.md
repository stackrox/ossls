[![CircleCI][circleci-badge]][circleci-link]

# OSSLS

📜 Automated dependency license scanning and auditing

## Installing

### From source

You can use `go get` to install a development version of tool by running:

```bash
$ go get -u github.com/stackrox/ossls
```

### Precompiled binary

Alternatively, you can download a static [release][github-release-link] binary using [fetch](https://github.com/gruntwork-io/fetch):

```bash
$ fetch --repo="https://github.com/stackrox/ossls" \
  --tag="0.2.0" --release-asset="ossls_linux_amd64" .
$ sudo install ossls_linux_amd64 /usr/bin/ossls
```

## Configuration

By default, ossls refers to a file named `.ossls.yml` for all configuration. The fine contains two top-level properties, `resolvers` and `dependencies`.

```yaml
resolvers:
  dep:
    manifest: Gopkg.toml
    vendor-dir: vendor

  js:
    manifest: ui/package.json
    module-dir: ui/node_modules

dependencies:
  ui/node_modules/react:
    url: https://github.com/facebook/react
    license: MIT
    files:
      LICENSE: 52412d7bc7ce4157ea628bbaacb8829e0a9cb3c58f57f99176126bc8cf2bfc85
      package.json:
        license: 809c46917bff0e06079ac81a33b2ee85061ce18988dc1ae584240fc6408328b1
    attribution:
    - Copyright (c) Facebook, Inc. and its affiliates.

  ui/node_modules/redux:
    url: https://github.com/reduxjs/redux
    ...
```

### Resolvers Configuration

This property provides sources for different dependency tracking manifests. Specifically, `Gopkg.toml` files used by [`dep`](https://github.com/golang/dep) and `package.json` files used by [`npm`](https://www.npmjs.com), [`yarn`](https://yarnpkg.com), and the like.

### Dependencies Configuration

This property provides a manifest for the current known set of project dependencies. Each sub-property is the relative name of a directory, containing a single installed dependency.

It has additional properties including a project url, the specific type of license it uses, and copyright attribution information. There is also a list of files with corresponding SHA256 hashes, for use during auditing.

## Usage

You can always view help information on the various actions like so:

```
$ ossls -help
Usage of ./ossls:
  -audit
        Audit all dependencies.
  -checksum
        Calculate checksum for a file.
  -config string
        Path to configuration file. (default ".ossls.yml")
  -list
        List all dependencies.
  -scan
        Scan single dependency.
  -version
        Displays the version and exits.
```

### Auditing Dependencies

Auditing, is the action of comparing the set of known dependencies to the set of currently installed dependencies, and detecting violations in our expectations.

You can run an audit like so:

```
$ ossls -audit
✓ ui/node_modules/react
✓ ui/node_modules/redux
...
```

#### Auditing Failures

Occasionally, typically after updating a dependency, an audit may fail. This section outlines the different failure types and their meaning.

##### Dependency Added

Indicates that a new dependency was added to a package manager manifest (like `Gopkg.toml` or `package.json`) but does not exist in the ossls dependency list.

```
$ ossls -audit
...
✗ ui/node_modules/example
  ↳ dependency added
ossls: violations found
```

##### Dependency Deleted

Indicates that a dependency was removed from a package manager manifest (like `Gopkg.toml` or `package.json`) but still exists in the ossls dependency list.

```
$ ossls -audit
...
✗ ui/node_modules/example
  ↳ dependency deleted
ossls: violations found
```

##### Checksum mismatch

Indicates that a pinned file for this dependency has been modified. Re-examine the file to determine if licensing or copyright holders have changed. You can re-generate the SHA256 checksum with `ossls -checksum <file>` or `shasum -a 256 <file>`.

```
$ ossls -audit
...
✗ ui/node_modules/example
  ↳ checksum mismatch for ui/node_modules/example/LICENSE. expected <some SHA256> but got <some other SHA256>
ossls: violations found
```

##### File does not exist

Indicates that a pinned file for this dependency has been renamed or deleted. Re-examine the dependency and update the list of pinned files.

```
$ ossls -audit
...
✗ ui/node_modules/example
  ↳ file ui/node_modules/example/LICENSE does not exist.
ossls: violations found
```

##### Invalid url / no license/attribution/files

Indicates that a property for this dependency is improperly specified, or left blank.

```
$ ossls -audit
...
✗ ui/node_modules/example
  ↳ no license
ossls: violations found
```

[circleci-badge]:      https://circleci.com/gh/stackrox/ossls.svg?&style=shield&circle-token=5ac8a87fbadae84c41f8c1fc868ad5d8ba85c90e
[circleci-link]:       https://circleci.com/gh/stackrox/ossls/tree/master
[github-release-link]: https://github.com/stackrox/ossls/releases/latest
