package main

import (
	"fmt"
	"log"
	"os"

	"golang.org/x/term"

	"github.com/fatih/color"
	"github.com/felkr/roamer/internal/allocation"
	"github.com/felkr/roamer/internal/configuration"
	"github.com/hashicorp/hcl/v2/hclsimple"
	"github.com/hashicorp/nomad/api"
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
func printGroup(group *api.TaskGroup, config configuration.Config) {
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
	drawUnitBar(memoryOfGroup, *config.ClusterConfig.Memory, "MB", "Memory")
	drawUnitBar(cpuOfGroup, *config.ClusterConfig.CPU, "MHz", "CPU")
	fmt.Println("\u2502")

	for _, task := range group.Tasks {
		fmt.Print("\u251c\u2500\u2500 ")
		c := color.New(color.Underline)
		c.Println(task.Name)
		drawUnitBar(*task.Resources.MemoryMB, *config.ClusterConfig.Memory, "MB", "\u2502\tMemory")
		drawUnitBar(*task.Resources.CPU, *config.ClusterConfig.CPU, "MHz", "\u2514\tCPU")
	}
}
func printJob(job *api.Job, config configuration.Config) {
	for _, group := range job.TaskGroups {
		printGroup(group, config)
	}
}
func getResourceUsage(client api.Client) (memory int64, cpu int64) {
	totalMemory := int64(0)
	totalCPU := int64(0)
	allocations, _, _ := client.Allocations().List(&api.QueryOptions{Params: map[string]string{"resources": "true"}})
	for _, allocation := range allocations {
		if allocation.AllocatedResources != nil {
			if allocation.ClientStatus == "running" {
				for _, task := range allocation.AllocatedResources.Tasks {
					totalMemory += task.Memory.MemoryMB
					totalCPU += task.Cpu.CpuShares
				}
			}
		}
	}
	return totalMemory, totalCPU
}
func getClusterResources(client api.Client) (memory int64, cpu int64) {
	totalMemory := int64(0)
	totalCPU := int64(0)
	nodes, _, _ := client.Nodes().List(&api.QueryOptions{
		Params: map[string]string{"resources": "true"},
	})
	for _, node := range nodes {
		if node.NodeResources != nil {
			totalCPU += node.NodeResources.Cpu.CpuShares
			totalMemory += int64(node.NodeResources.Memory.MemoryMB)
		}
	}
	return totalMemory, totalCPU
}

func main() {
	address := new(string)
	app := &cli.App{
		Name:  "roamer",
		Usage: "streamlined nomad deployment",
	}

	app.Flags = append(app.Flags, &cli.BoolFlag{
		Name:    "yes",
		Usage:   "Don't ask questions, answer yes",
		Aliases: []string{"y"},
	})

	app.Flags = append(app.Flags, &cli.StringFlag{
		Name:        "address",
		Usage:       "The address of the Nomad server",
		Value:       "http://localhost:4646",
		Destination: address,
	})

	cli.AppHelpTemplate = fmt.Sprintf(`%s
EXAMPLES:
   roamer deploy --config deployment.hcl my_project.nomad
   roamer overview
	`, cli.AppHelpTemplate)

	var config configuration.Config

	app.Commands = []*cli.Command{
		{
			Name: "overview",
			Action: func(c *cli.Context) error {
				clientConfig := api.DefaultConfig()
				clientConfig.Address = *address
				client, _ := api.NewClient(clientConfig)
				clusterMemory, clusterCPU := getClusterResources(*client)

				jobs, _, err := client.Jobs().List(&api.QueryOptions{})
				for _, job := range jobs {
					c := color.New(color.Underline)
					fmt.Print(" ")
					c.Println(job.Name)
					fmt.Print("┌")
					width, _, _ := term.GetSize(0)
					for i := 0; i < width-1; i++ {
						fmt.Print("\u2500")
					}
					fmt.Println()
					allocations, _, _ := client.Jobs().Allocations(job.ID, true, &api.QueryOptions{})
					for _, allocation := range allocations {
						fmt.Println("│ " + allocation.Name)
						res, _, _ := client.Allocations().Info(allocation.ID, &api.QueryOptions{})
						drawUnitBar(*res.Resources.MemoryMB, int(clusterMemory), "MB", "│\tMemory")
						drawUnitBar(*res.Resources.CPU, int(clusterCPU), "MHz", "│\tCPU")

						fmt.Print("├")
						width, _, _ := term.GetSize(0)
						for i := 0; i < width-1; i++ {
							fmt.Print("\u2500")
						}

					}
				}

				if err != nil {
					return cli.Exit("Nomad returned an error: "+err.Error(), 1)
				}
				return nil
			},
		},
		{
			Name:      "deploy",
			UsageText: "roamer [global flags] deploy --config <config hcl file> <job hcl file>",
			Aliases:   []string{"d"},
			Flags: []cli.Flag{&cli.StringFlag{
				Name:     "config",
				Usage:    "Path to the configuration file",
				Required: true,
			}},
			Usage: "Allocate resources and deploy to a nomad server",
			Action: func(c *cli.Context) error {
				err := hclsimple.DecodeFile(c.String("config"), nil, &config)
				if err != nil {
					return cli.Exit("Failed to load configuration: "+err.Error(), 1)
				}
				if c.NArg() != 1 {
					cli.ShowAppHelpAndExit(c, 1)
				}
				jobPath := c.Args().Get(0)

				jobFile, err := os.Open(jobPath)

				if os.IsNotExist(err) {
					return cli.Exit(jobPath+": No such file or directory", 1)
				}

				job, err := jobspec2.Parse(jobPath, jobFile)
				if err != nil {
					log.Fatal(err)
				}
				clientConfig := api.DefaultConfig()
				clientConfig.Address = *address
				client, _ := api.NewClient(clientConfig)
				_, err = client.Status().Leader() // Try a basic command to find out if we succeeded in connecting to the Nomad agent
				if err != nil {
					log.Fatal(err)
				}
				clusterMemoryTotal, clusterCPUTotal := getClusterResources(*client)
				usedMemory, usedCPU := getResourceUsage(*client)
				clusterMemory := clusterMemoryTotal - usedMemory
				clusterCPU := clusterCPUTotal - usedCPU

				config.ClusterConfig.CPU = new(int)
				*config.ClusterConfig.CPU = int(clusterCPU)
				config.ClusterConfig.Memory = new(int)
				*config.ClusterConfig.Memory = int(clusterMemory)

				// Don't swap those two conditions around (short-circuit evaluation)
				if config.ClusterConfig == nil || config.ClusterConfig.SafetyMargin == nil {
					safetyMargin := new(int)
					*safetyMargin = 3
					config.ClusterConfig.SafetyMargin = safetyMargin
				}
				if config.ClusterConfig == nil || config.ClusterConfig.CPU == nil {
					infrastructureCPU := new(int)
					*infrastructureCPU = int(clusterCPU)
					config.ClusterConfig.CPU = infrastructureCPU
				}
				if config.ClusterConfig == nil || config.ClusterConfig.Memory == nil {
					infrastructureMemory := new(int)
					*infrastructureMemory = int(clusterMemory)
					config.ClusterConfig.Memory = infrastructureMemory
				}

				err = allocation.Allocate(config, job)
				if err != nil {
					return cli.Exit("Failed create allocation: "+err.Error(), 1)
				}

				// Print
				printJob(job, config)

				if c.Bool("yes") || askForDeployment() {
					_, _, err = client.Jobs().Register(job, &api.WriteOptions{})
					if err != nil {
						return cli.Exit("Nomad returned an error: "+err.Error(), 1)
					}
				}

				if err != nil {
					return cli.Exit("Nomad returned an error: "+err.Error(), 1)
				}
				return nil
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
