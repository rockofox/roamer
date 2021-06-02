package allocation

import (
	"errors"

	"github.com/felkr/roamer/internal/configuration"
	"github.com/hashicorp/nomad/api"
)

func Allocate(config configuration.Config, job *api.Job) error {
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
					return errors.New("Sub of weights greater than 100")
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
			}
		}
	}
	return nil
}
