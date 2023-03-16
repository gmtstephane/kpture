# Kpture

[![Maintainability](https://api.codeclimate.com/v1/badges/cc2f36936ff42aa2376d/maintainability)](https://codeclimate.com/github/gmtstephane/kpture/maintainability)
[![Test Coverage](https://api.codeclimate.com/v1/badges/cc2f36936ff42aa2376d/test_coverage)](https://codeclimate.com/github/gmtstephane/kpture/test_coverage) 
![GitHub Workflow Status](https://img.shields.io/github/actions/workflow/status/gmtstephane/kpture/codeql.yml?label=codeQL&logo=GitHub&logoColor=white )
![GitHub Workflow Status](https://img.shields.io/github/actions/workflow/status/gmtstephane/kpture/unit-tests.yaml?label=unit%20test&logo=GitHub%20Actions&logoColor=white )
![GitHub Workflow Status](https://img.shields.io/github/actions/workflow/status/gmtstephane/kpture/e2e.yaml?label=e2e%20test&logo=GitHub%20Actions&logoColor=white )
[![Go Report Card](https://goreportcard.com/badge/github.com/gmtstephane/kpture)](https://goreportcard.com/report/github.com/gmtstephane/kpture)

## Description
Kpture is a simple tool that allows you to capture packets from remote pods in your cluster using ephemeral debug containers injection.


<img src="assets/kptureTerm.svg" width="100%" >

## Installation

### With **golang** :
  
```bash
go install --tags=cli github.com/gmtstephane/kpture@latest
```

### With **homebrew** :
  
```bash
brew install gmtstephane/kpture/kpture
```
### With **prebuilt binaries**
You can find the latest binaries for your platform on the [releases page](http://www.github.com/gmtstephane/kpture/releases).

### Prerequisites
- A kubernetes cluster version 1.23 or higher

## Capturing packets
#### Start kpture in separated pcap files
```bash
kpture packets nginx-679f748897-vmc5r nginx-6fdt248897-380f4  -o output
```
This will create the following output directory: 
```bash 
output
├── nginx-679f748897-vmc5r.pcap
└── nginx-6fdt248897-380f4.pcap
```
#### Start kpture and pipe the output to **tshark**
```bash
kpture packets nginx-679f748897-vmc5r nginx-6fdt248897-380f4  --raw | tshark -r -
```
#### Start kpture all pods in current namespace to ./output and pipe the output to **wireshark** at the same time
```bash
kpture packets --all -o output --raw | wireshark -k -i -
```

### Options

```
  -a, --all             Capture from all pods in the selected namespace
  -h, --help            help for packets
  -o, --output string   output folder
  -r, --raw             Print raw packet to stdout (for tshark/wireshark)
  -s, --split           split pcap files per pod (default true)
```

## Auto completion

You can enable auto completion (pods and options) for bash and zsh by running the following command:

```bash
kpture completion bash > /etc/bash_completion.d/kpture
```

```bash
kpture completion zsh > /usr/local/share/zsh/site-functions/_kpture
```