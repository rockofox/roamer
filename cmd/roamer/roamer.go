package main

import (
	"fmt"
	"log"
	"os"

	"github.com/fatih/color"
	"github.com/felkr/roamer/internal/allocation"
	"github.com/felkr/roamer/internal/configuration"
	"github.com/hashicorp/hcl/v2/hclsimple"
	"github.com/hashicorp/nomad/api"
	"github.com/hashicorp/nomad/jobspec"
	"github.com/hashicorp/nomad/jobspec2"
	"github.com/manifoldco/promptui"
	"github.com/urfave/cli/v2"
)

func drawUnitBar(part int, full int, unit string, label string) {
	const width = 25
	bar := make([]rune, width)
	fmt.Printf("%s\t%d/%d %s\t", label, part, full, unit)
	for i := 0; i < width; i++ {
		bar[i] = '\u25a1'
	}
	for i := 0; i < width*part/full; i++ {
		bar[i] = '\u25a0'
	}
	fmt.Printf("%s  %.2f%%\n", string(bar), float32(part)/float32(full)*100.0)
}
func askForDeployment() bool {
	prompt := promptui.Select{
		Label: "Deploy? [Yes/No]",
		Items: []string{"Yes", "No"},
	}
	_, result, err := prompt.Run()
	if err != nil {
		log.Fatalf("Prompt failed %v\n", err)
	}
	return result == "Yes"
}
func main() {
	jobspec.ParseFile("config.hcl")
	address := new(string)
	app := &cli.App{
		Name:      "roamer",
		Usage:     "streamlined nomad deployment",
		UsageText: "roamer [flags] <config hcl file> <job hcl file>",
	}
	app.Flags = append(app.Flags, &cli.StringFlag{
		Name:        "address",
		Usage:       "The address of the Nomad server",
		Value:       "http://localhost:4646",
		Destination: address,
	})
	app.Flags = append(app.Flags, &cli.BoolFlag{
		Name:    "yes",
		Usage:   "Don't ask questions, answer yes",
		Aliases: []string{"y"},
	})

	app.Action = func(c *cli.Context) error {
		if c.NArg() < 2 {
			cli.ShowAppHelpAndExit(c, 1)
		}
		jobPath := c.Args().Get(1)
		configPath := c.Args().Get(0)

		jobFile, err := os.Open(jobPath)

		if os.IsNotExist(err) {
			return cli.Exit(jobPath+": No such file or directory", 1)
		}

		job, err := jobspec2.Parse(jobPath, jobFile)
		if err != nil {
			log.Fatal(err)
		}
		var config configuration.Config
		err = hclsimple.DecodeFile(configPath, nil, &config)
		if err != nil {
			return cli.Exit("Failed to load configuration: "+err.Error(), 1)
		}

		err = allocation.Allocate(config, job)
		if err != nil {
			return cli.Exit("Failed create allocation: "+err.Error(), 1)
		}

		// Print
		for _, group := range job.TaskGroups {
			c := color.New(color.Bold).Add(color.FgBlack)
			c.Printf("%s", *group.Name)

			if group.Count != nil {
				c.Printf(" (%d)\n", *group.Count)
			} else {
				c.Printf("\n")
			}
			memoryOfGroup := 0
			cpuOfGroup := 0
			for _, task := range group.Tasks {
				memoryOfGroup += *task.Resources.MemoryMB
				cpuOfGroup += *task.Resources.CPU
			}
			drawUnitBar(memoryOfGroup, config.Infrastructure.Memory, "MB", "Memory")
			drawUnitBar(cpuOfGroup, config.Infrastructure.CPU, "MHz", "CPU")
			fmt.Println("\u2502")

			for _, task := range group.Tasks {
				fmt.Print("\u251c\u2500\u2500 ")
				c := color.New(color.FgBlack).Add(color.Bold)
				c.Println(task.Name)
				drawUnitBar(*task.Resources.MemoryMB, config.Infrastructure.Memory, "MB", "\u2502\tMemory")
				drawUnitBar(*task.Resources.CPU, config.Infrastructure.CPU, "MHz", "\u2514\tCPU")
			}
		}

		clientConfig := api.DefaultConfig()
		clientConfig.Address = *address
		client, err := api.NewClient(clientConfig)
		if c.Bool("yes") || askForDeployment() {
			client.Jobs().Plan(job, false, &api.WriteOptions{})
			_, err = client.Status().Leader()
		}

		if err != nil {
			return cli.Exit("Nomad returned an error: "+err.Error(), 1)
		}
		return nil
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
