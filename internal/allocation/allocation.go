package allocation

import (
	"errors"

	"github.com/felkr/roamer/internal/configuration"
	"github.com/hashicorp/nomad/api"
)

func applySafetyMargin(resource int, config configuration.Config) int {
	return int(float32(resource) * (1.0 - (float32(*config.ClusterConfig.SafetyMargin) / 100.0)))
}
func minOf(vars ...int) int {
	min := vars[0]

	for _, i := range vars {
		if min > i {
			min = i
		}
	}

	return min
}

// Allocate resources to the given job
func Allocate(config configuration.Config, job *api.Job, smallestMemory int, smallestCPU int) error {
	absoluteTasks := 0
	for _, group := range job.TaskGroups {
		tasksInGroup := len(group.Tasks)

		if group.Count != nil {
			tasksInGroup *= *group.Count
		}
		absoluteTasks += tasksInGroup
	}
	availableCPU := *config.ClusterConfig.CPU
	availableMemory := *config.ClusterConfig.Memory
	println(availableMemory)
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
					return errors.New("sum of weights greater than 100")
				}
				if groupConfig.Name == *group.Name {
					assignedMemory := minOf(smallestMemory, availableMemory) * groupConfig.Weight / 100 / len(group.Tasks)
					assignedCPU := minOf(availableCPU, smallestCPU) * groupConfig.Weight / 100 / len(group.Tasks)
					if task.Resources == nil {
						task.Resources = new(api.Resources)
						task.Resources.MemoryMB = new(int)
						task.Resources.CPU = new(int)
					}
					*task.Resources.MemoryMB = applySafetyMargin(assignedMemory, config)
					*task.Resources.CPU = applySafetyMargin(assignedCPU, config)
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
				assignedMemory := minOf(availableMemory, smallestMemory) / weightlessTasks
				assignedCPU := minOf(availableCPU, smallestCPU) / weightlessTasks
				if task.Resources == nil {
					task.Resources = new(api.Resources)
					task.Resources.MemoryMB = new(int)
					task.Resources.CPU = new(int)
				}
				*task.Resources.MemoryMB = applySafetyMargin(assignedMemory, config)
				*task.Resources.CPU = applySafetyMargin(assignedCPU, config)
			}
		}
	}
	return nil
}
