package main

import (
	"fmt"
	"log"
	"os"

	"github.com/fatih/color"
	"github.com/felkr/roamer/internal/configuration"
	"github.com/hashicorp/hcl/v2/hclsimple"
	"github.com/hashicorp/nomad/api"
	"github.com/hashicorp/nomad/jobspec"
	"github.com/hashicorp/nomad/jobspec2"
	"github.com/urfave/cli"
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
func main() {
	jobspec.ParseFile("config.hcl")
	address := new(string)
	app := &cli.App{
		Name:      "roamer",
		Usage:     "streamlined nomad deployment",
		UsageText: "roamer <config hcl file> <job hcl file>",
	}
	app.Flags = append(app.Flags, &cli.StringFlag{
		Name:        "address",
		Usage:       "The address of the Nomad server",
		Value:       "http://localhost:4646",
		Destination: address,
	})

	app.Action = func(c *cli.Context) error {
		if c.NArg() < 2 {
			cli.ShowAppHelpAndExit(c, 1)
		}
		jobPath := c.Args().Get(1)
		configPath := c.Args().Get(0)

		jobFile, err := os.Open(jobPath)

		if os.IsNotExist(err) {
			return cli.NewExitError(jobPath+": No such file or directory", 1)
		}

		job, err := jobspec2.Parse(jobPath, jobFile)
		if err != nil {
			log.Fatal(err)
		}
		var config configuration.Config
		err = hclsimple.DecodeFile(configPath, nil, &config)
		if err != nil {
			log.Fatalf("Failed to load configuration: %s", err)
		}
		if config.Infrastructure.SafetyMargin == 0 {
			config.Infrastructure.SafetyMargin = 3
		}
		absoluteTasks := 0
		for _, group := range job.TaskGroups {
			tasksInGroup := len(group.Tasks)

			if group.Count != nil {
				tasksInGroup *= *group.Count
			}
			absoluteTasks += tasksInGroup
		}
		// cpu := distributeEvenly(config.Infrastructure.CPU, absoluteTasks, config)
		// mem := distributeEvenly(config.Infrastructure.Memory, absoluteTasks, config)
		// fmt.Printf("CPU per group %d\n", cpu)
		// fmt.Printf("Memory per group %d\n", mem)
		availableCPU := config.Infrastructure.CPU * (1.0 - ((config.Infrastructure.SafetyMargin) / 100))
		availableMemory := config.Infrastructure.Memory * (1.0 - ((config.Infrastructure.SafetyMargin) / 100))
		weightlessTasks := absoluteTasks
		// First assign resources to the groups that have a weight set in the config file
		for _, group := range job.TaskGroups {
			if group.Count == nil {
				group.Count = new(int)
				*group.Count = 1
			}
			for _, task := range group.Tasks {

				sumOfWeights := 0
				for _, groupConfig := range config.Groups {
					sumOfWeights += groupConfig.Weight
					if sumOfWeights > 100 {
						return cli.NewExitError("Sum of weights greater than 100", 1)
					}
					if groupConfig.Name == *group.Name {
						assignedMemory := config.Infrastructure.Memory * groupConfig.Weight / 100 / len(group.Tasks)
						assignedCPU := config.Infrastructure.CPU * groupConfig.Weight / 100 / len(group.Tasks)
						*task.Resources.MemoryMB = assignedMemory
						*task.Resources.CPU = assignedCPU
						availableCPU -= assignedCPU
						availableMemory -= assignedMemory
						weightlessTasks--
					}
				}
			}
		}
		// Then evenly split up the rest
		for _, group := range job.TaskGroups {
			if group.Count == nil {
				group.Count = new(int)
				*group.Count = 1
			}
			found := false
			for _, task := range group.Tasks {
				for _, groupConfig := range config.Groups {
					if groupConfig.Name == *group.Name {
						found = true
					}
				}

				if !found {
					assignedMemory := availableMemory / weightlessTasks
					assignedCPU := availableCPU / weightlessTasks
					*task.Resources.MemoryMB = assignedMemory
					*task.Resources.CPU = assignedCPU
					// availableCPU -= assignedCPU
					// availableMemory -= assignedMemory
				}
			}
		}

		// Print
		for _, group := range job.TaskGroups {
			c := color.New(color.Underline)
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
				c = color.New(color.FgBlack).Add(color.Italic)
				// fmt.Println()
				drawUnitBar(*task.Resources.MemoryMB, config.Infrastructure.Memory, "MB", "\u2502\tMemory")
				drawUnitBar(*task.Resources.CPU, config.Infrastructure.CPU, "MHz", "\u2514\tCPU")
			}
		}

		clientConfig := api.DefaultConfig()
		clientConfig.Address = *address
		// client, err := api.NewClient(clientConfig)
		// client.Jobs().Plan(job, false, &api.WriteOptions{})

		if err != nil {
			log.Fatalf("%s", err)
		}
		return nil
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
