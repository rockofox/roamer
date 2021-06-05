<p align="center">
  <a href="https://github.com/felkr/roamer/">
    <img src="logo.svg" alt="Logo" width="300" height="100">
  </a>
  <p align="center">
    Streamlined <a href="https://github.com/hashicorp/nomad">Nomad</a> deployment
  </p>
</p>

[![Codacy Badge](https://app.codacy.com/project/badge/Grade/c15075cc8342480abe6bf67cd64e06f8)](https://www.codacy.com?utm_source=github.com\&utm_medium=referral\&utm_content=felkr/roamer\&utm_campaign=Badge_Grade) [![Go](https://github.com/felkr/roamer/actions/workflows/go.yml/badge.svg)](https://github.com/felkr/roamer/actions/workflows/go.yml)

`roamer` is a tool which aims to simplify and streamline the deployment of jobs onto Nomad clusters.


**This software is still in an early stage and should be considered alpha. While it shouldn't break anything, I can't guarantee you that it doesn't. Also, some error handling and documentation may be missing.**

## Features

*   Automatic allocation of resources
*   Possibility to set weights to allocate more or less resources to a certain task group
*   Configuration via [HCL](https://github.com/hashicorp/hcl) files
*   Useful tools for observing deployments

## Usage

    roamer [global options] command [command options] [arguments...]

### Commands

| Command  | Functionality                                    |
| -------- | ------------------------------------------------ |
| overview | Show a simple overview                           |
| deploy   | Allocate resources and deploy to a nomad server  |
| help     | Shows a list of commands or help for one command |

### Global Options

| Flag            | Meaning                                                            |
| --------------- | ------------------------------------------------------------------ |
| --yes, -y       | Don't ask questions, answer yes (default: false)                   |
| --address value | The address of the Nomad server (default: "http://localhost:4646") |
| --help, -h      | show help (default: false)                                         |

### [Examples](https://github.com/felkr/roamer/wiki/Basic-Example)
