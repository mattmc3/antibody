# Contributing

By participating to this project, you agree to abide our [code of
conduct](/CODE_OF_CONDUCT.md).

## Setup your machine

`antibody` is written in [Go](https://golang.org/).

Prerequisites are:

- Build:
  - `just`
  - [Go 1.26+](http://golang.org/doc/install)

Clone `antibody` from source into `$GOPATH`:

```sh
$ mkdir -p $GOPATH/src/github.com/mattmc3
$ cd $_
$ git clone git@github.com:mattmc3/antibody.git
$ cd antibody
```

Install the build and lint dependencies:

```sh
$ just setup
```

A good way of making sure everything is all right is running the test suite:

```sh
$ just test
```

## Test your change

You can create a branch for your changes and try to build from the source as you go:

```sh
$ just build
```

When you are satisfied with the changes, we suggest you run:

```sh
$ just ci
```

Which runs all the linters and tests.

## Submit a pull request

Push your branch to your `antibody` fork and open a pull request against the
main branch.
