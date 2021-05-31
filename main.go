package main

import (
	"fmt"
	"log"
	"os"

	"github.com/hashicorp/hcl/v2/hclsimple"
	"github.com/hashicorp/nomad/api"
	"github.com/hashicorp/nomad/jobspec"
	"github.com/hashicorp/nomad/jobspec2"
	"github.com/urfave/cli"
)

type Config struct {
	Infrastructure InfrastructureConfig `hcl:"infrastructure,block"`
}

type InfrastructureConfig struct {
	Memory int `hcl:"memory"`
	CPU    int `hcl:"cpu"`
}

func main() {
	jobspec.ParseFile("config.hcl")

	app := &cli.App{
		Name:      "roamer",
		Usage:     "streamlined nomad deployment",
		UsageText: "roamer <config hcl file> <job hcl file>",
	}

	app.Action = func(c *cli.Context) error {
		if c.NArg() < 2 {
			cli.ShowAppHelpAndExit(c, 1)
		}
		job_path := c.Args().Get(1)
		config_path := c.Args().Get(0)

		job_file, err := os.Open(job_path)

		if os.IsNotExist(err) {
			return cli.NewExitError(job_path+": No such file or directory", 1)
		}

		job, err := jobspec2.Parse(job_path, job_file)
		if err != nil {
			log.Fatal(err)
		}
		var config Config
		err = hclsimple.DecodeFile(config_path, nil, &config)
		if err != nil {
			log.Fatalf("Failed to load configuration: %s", err)
		}
		fmt.Printf("CPU per group %d\n", config.Infrastructure.CPU/len(job.TaskGroups))
		fmt.Printf("Memory per group %d\n", config.Infrastructure.Memory/len(job.TaskGroups))
		fmt.Printf("%s\n", job)

		client_config := api.DefaultConfig()
		client_config.Address = "http://localhost:4646"
		client, err := api.NewClient(client_config)

		if err != nil {
			log.Fatalf("%s", err)
		}
		fmt.Println(client.Agent().Region())
		return nil
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
