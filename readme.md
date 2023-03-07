<img src="assets/kpture.svg" width="100">

# Kpture

[![Maintainability](https://api.codeclimate.com/v1/badges/cc2f36936ff42aa2376d/maintainability)](https://codeclimate.com/github/gmtstephane/kpture/maintainability)
[![Test Coverage](https://api.codeclimate.com/v1/badges/cc2f36936ff42aa2376d/test_coverage)](https://codeclimate.com/github/gmtstephane/kpture/test_coverage) 
![GitHub Workflow Status](https://img.shields.io/github/actions/workflow/status/gmtstephane/kpture/codeql.yml?label=codeQL&logo=GitHub&logoColor=white )
![GitHub Workflow Status](https://img.shields.io/github/actions/workflow/status/gmtstephane/kpture/unit-tests.yaml?label=unit%20test&logo=GitHub%20Actions&logoColor=white )
![GitHub Workflow Status](https://img.shields.io/github/actions/workflow/status/gmtstephane/kpture/e2e.yaml?label=e2e%20test&logo=GitHub%20Actions&logoColor=white )
[![Go Report Card](https://goreportcard.com/badge/github.com/gmtstephane/kpture)](https://goreportcard.com/report/github.com/gmtstephane/kpture)

## Description
Kpture is a simple tool that allows you to capture packets and logs from remote pods in your cluster using ephemeral debug containers injection.


<img src="assets/architecture.svg" width="100%" height="200">

## Installation
### Using Go

```bash
go install --tags=cli  github.com/gmtstephane/kpture
```

### Using pre-built binaries 
Download the latest release from the [release page](https://github.com/gmtstephane/kpture/releases/)


## Getting started
### Capture packets from a pod
kpture uses you kubectl context to connect to your cluster. You can specify the namespace and pod name to capture packets from.

```bash
kpture packets -p my-pod
```


##
<img src="assets/kpture.gif" width="100%">
