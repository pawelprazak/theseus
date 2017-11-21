# Theseus

## Overview
Theseus is a command-line tool for comparing sets of Kubernetes objects. It reads object definitions from a running cluster, an Ark backup file, or a local directory, and performs a comparison to another source of object definitions. Theseus provides reports of which objects are in the first source only, the second source only, and both. Additionally, for objects that are in both, it performs a line-by-line comparison of the object definitions, which it outputs to stdout with highlighting, and to a text file.

Theseus is intended to be used either via CLI by end-users, or as an imported package within other applications.

## Getting Started
To chekout Theseus locally:
```
git clone https://github.com/heptiolabs/theseus $GOPATH/src/github.com/heptio/theseus
```

To build Theseus locally:
```
go build ./cmd/theseus
```

To run Theseus:
```
./theseus diff <source-1> <source-2>
```
where each source is one of `cluster=<kubeconfig-path>`, `backup=<backup-path>`, or `directory=<directory-path>`.

Additionally, Theseus supports filtering the objects to be compared by *scope*, i.e. a list of specific namespaces and/or "cluster" for cluster-scoped objects, and also by *label selector*. For example:
```
./theseus diff cluster=./kubeconfig backup=./my-backup.tar.gz --included-scopes cluster,my-ns --selector app=my-app
``` 

## Importing
The logical entrypoint for using Theseus as an imported package is the `pkg/diff` package, specifically the `Generate` function. 
