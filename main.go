package main

import (
	"errors"
	"net/rpc"
	"time"

	"github.com/AlertFlow/runner/pkg/executions"
	"github.com/AlertFlow/runner/pkg/plugins"

	"github.com/v1Flows/alertFlow/services/backend/pkg/models"

	"github.com/hashicorp/go-plugin"
)

type Receiver struct {
	Receiver string `json:"receiver"`
}

// CollectDataActionPlugin is an implementation of the Plugin interface
type CollectDataActionPlugin struct{}

func (p *CollectDataActionPlugin) ExecuteTask(request plugins.ExecuteTaskRequest) (plugins.Response, error) {
	err := executions.UpdateStep(request.Config, request.Execution.ID.String(), models.ExecutionSteps{
		ID:        request.Step.ID,
		Messages:  []string{"Checking for flow actions"},
		Status:    "running",
		StartedAt: time.Now(),
	})
	if err != nil {
		return plugins.Response{
			Success: false,
		}, err
	}

	// check if flow got any action
	if len(request.Flow.Actions) > 0 {
		count := 0
		for _, action := range request.Flow.Actions {
			if action.Active {
				count++
			}
		}

		if count == 0 {
			err := executions.UpdateStep(request.Config, request.Execution.ID.String(), models.ExecutionSteps{
				ID:         request.Step.ID,
				Messages:   []string{"Flow has no active Actions defined. Cancel execution"},
				Status:     "canceled",
				CanceledBy: "Flow Action Check",
				CanceledAt: time.Now(),
				FinishedAt: time.Now(),
			})
			if err != nil {
				return plugins.Response{
					Success: false,
				}, err
			}
			return plugins.Response{
				Data: map[string]interface{}{
					"status": "canceled",
				},
				Success: false,
			}, nil
		} else {
			err := executions.UpdateStep(request.Config, request.Execution.ID.String(), models.ExecutionSteps{
				ID:         request.Step.ID,
				Messages:   []string{"Flow has Actions defined"},
				Status:     "finished",
				FinishedAt: time.Now(),
			})
			if err != nil {
				return plugins.Response{
					Success: false,
				}, err
			}
			return plugins.Response{
				Success: true,
			}, nil
		}
	} else {
		err := executions.UpdateStep(request.Config, request.Execution.ID.String(), models.ExecutionSteps{
			ID:         request.Step.ID,
			Messages:   []string{"Flow has no Actions defined. Cancel execution"},
			Status:     "canceled",
			CanceledBy: "Flow Action Check",
			CanceledAt: time.Now(),
			FinishedAt: time.Now(),
		})
		if err != nil {
			return plugins.Response{
				Success: false,
			}, err
		}
		return plugins.Response{
			Data: map[string]interface{}{
				"status": "canceled",
			},
			Success: false,
		}, nil
	}
}

func (p *CollectDataActionPlugin) HandlePayload(request plugins.PayloadHandlerRequest) (plugins.Response, error) {
	return plugins.Response{
		Success: false,
	}, errors.New("not implemented")
}

func (p *CollectDataActionPlugin) Info() (models.Plugins, error) {
	var plugin = models.Plugins{
		Name:    "Actions Check",
		Type:    "action",
		Version: "1.1.0",
		Author:  "JustNZ",
		Actions: models.Actions{
			Name:        "Actions Check",
			Description: "Check for actions in flow",
			Plugin:      "actions_check",
			Icon:        "solar:bolt-linear",
			Category:    "Flow",
			Params:      nil,
		},
		Endpoints: models.PayloadEndpoints{},
	}

	return plugin, nil
}

// PluginRPCServer is the RPC server for Plugin
type PluginRPCServer struct {
	Impl plugins.Plugin
}

func (s *PluginRPCServer) ExecuteTask(request plugins.ExecuteTaskRequest, resp *plugins.Response) error {
	result, err := s.Impl.ExecuteTask(request)
	*resp = result
	return err
}

func (s *PluginRPCServer) HandlePayload(request plugins.PayloadHandlerRequest, resp *plugins.Response) error {
	result, err := s.Impl.HandlePayload(request)
	*resp = result
	return err
}

func (s *PluginRPCServer) Info(args interface{}, resp *models.Plugins) error {
	result, err := s.Impl.Info()
	*resp = result
	return err
}

// PluginServer is the implementation of plugin.Plugin interface
type PluginServer struct {
	Impl plugins.Plugin
}

func (p *PluginServer) Server(*plugin.MuxBroker) (interface{}, error) {
	return &PluginRPCServer{Impl: p.Impl}, nil
}

func (p *PluginServer) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &plugins.PluginRPC{Client: c}, nil
}

func main() {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: plugin.HandshakeConfig{
			ProtocolVersion:  1,
			MagicCookieKey:   "PLUGIN_MAGIC_COOKIE",
			MagicCookieValue: "hello",
		},
		Plugins: map[string]plugin.Plugin{
			"plugin": &PluginServer{Impl: &CollectDataActionPlugin{}},
		},
		GRPCServer: plugin.DefaultGRPCServer,
	})
}
