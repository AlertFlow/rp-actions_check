package main

import (
	"time"

	"github.com/gin-gonic/gin"
	"gitlab.justlab.xyz/alertflow-public/runner/pkg/executions"
	"gitlab.justlab.xyz/alertflow-public/runner/pkg/models"
)

type ActionsCheckPlugin struct{}

func (p *ActionsCheckPlugin) Init() models.Plugin {
	return models.Plugin{
		Name:    "Actions Check",
		Type:    "action",
		Version: "1.0.1",
		Creator: "JustNZ",
	}
}

func (p *ActionsCheckPlugin) Details() models.PluginDetails {
	return models.PluginDetails{
		Action: models.ActionDetails{
			Name:        "Actions Check",
			Description: "Check if there are any actions defined in the flow",
			Icon:        "solar:bolt-linear",
			Type:        "actions_check",
			Category:    "Flow",
			Function:    p.Execute,
			IsHidden:    true,
			Params:      nil,
		},
	}
}

func (p *ActionsCheckPlugin) Execute(execution models.Execution, flow models.Flows, payload models.Payload, steps []models.ExecutionSteps, step models.ExecutionSteps, action models.Actions) (data map[string]interface{}, finished bool, canceled bool, no_pattern_match bool, failed bool) {
	err := executions.UpdateStep(execution.ID.String(), models.ExecutionSteps{
		ID:             step.ID,
		ActionID:       action.ID.String(),
		ActionMessages: []string{"Checking for flow actions"},
		Pending:        false,
		Running:        true,
		StartedAt:      time.Now(),
	})
	if err != nil {
		return nil, false, false, false, true
	}

	// check if flow got any action
	if len(flow.Actions) > 0 {
		count := 0
		for _, action := range flow.Actions {
			if action.Active {
				count++
			}
		}

		if count == 0 {
			err := executions.UpdateStep(execution.ID.String(), models.ExecutionSteps{
				ID:             step.ID,
				ActionMessages: []string{"Flow has no active Actions defined. Cancel execution"},
				Running:        false,
				Canceled:       true,
				CanceledBy:     "Flow Action Check",
				CanceledAt:     time.Now(),
				Finished:       true,
				FinishedAt:     time.Now(),
			})
			if err != nil {
				return nil, false, false, false, true
			}
			return nil, false, true, false, false
		} else {
			err := executions.UpdateStep(execution.ID.String(), models.ExecutionSteps{
				ID:             step.ID,
				ActionMessages: []string{"Flow has Actions defined"},
				Running:        false,
				Finished:       true,
				FinishedAt:     time.Now(),
			})
			if err != nil {
				return nil, false, false, false, true
			}
			return nil, true, false, false, false
		}
	} else {
		err := executions.UpdateStep(execution.ID.String(), models.ExecutionSteps{
			ID:             step.ID,
			ActionMessages: []string{"Flow has no Actions defined. Cancel execution"},
			Canceled:       true,
			CanceledBy:     "Flow Action Check",
			CanceledAt:     time.Now(),
			Finished:       true,
			FinishedAt:     time.Now(),
		})
		if err != nil {
			return nil, false, false, false, true
		}
		return nil, false, true, false, false
	}
}

func (p *ActionsCheckPlugin) Handle(context *gin.Context) {}

var Plugin ActionsCheckPlugin
