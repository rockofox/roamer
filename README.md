
<p align="center">
  <a href="https://github.com/felkr/roamer/">
    <img src="logo.svg" alt="Logo" width="300" height="100">
  </a>
  <p align="center">
    Streamlined <a href="https://github.com/hashicorp/nomad">Nomad</a> deployment
  </p>
</p>

[![Codacy Badge](https://app.codacy.com/project/badge/Grade/c15075cc8342480abe6bf67cd64e06f8)](https://www.codacy.com?utm_source=github.com&amp;utm_medium=referral&amp;utm_content=felkr/roamer&amp;utm_campaign=Badge_Grade) [![Go](https://github.com/felkr/roamer/actions/workflows/go.yml/badge.svg)](https://github.com/felkr/roamer/actions/workflows/go.yml)

`roamer` is a tool which aims to simplify and streamline the deployment of jobs onto Nomad clusters.



## Features
* Automatic allocation of resources
* Possibility to set weights to allocate more or less resources to a certain task group
* Configuration via [HCL](https://github.com/hashicorp/hcl) files
* Useful tools for observing deployments
## Installation
    go install github.com/felkr/roamer
## Usage
      roamer [global options] command [command options] [arguments...]

    COMMANDS:
      overview   
      deploy, d  Allocate resources and deploy to a nomad server
      help, h    Shows a list of commands or help for one command

    GLOBAL OPTIONS:
      --yes, -y        Don't ask questions, answer yes (default: false)
      --address value  The address of the Nomad server (default: "http://localhost:4646")
      --help, -h       show help (default: false)

    EXAMPLES:
      roamer deploy --config deployment.hcl my_project.nomad