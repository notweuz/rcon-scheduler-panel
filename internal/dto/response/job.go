package response

import "rcon-scheduler-panel/internal/model"

type Job struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	ServerID   int64  `json:"server_id"`
	Cron       string `json:"cron"`
	LogResults bool   `json:"log_results"`
	Steps      []Step `json:"steps"`
}

type Step struct {
	Command string `json:"command,omitempty"`
	Wait    string `json:"wait,omitempty"`
}

func JobFromModel(job model.Job) Job {
	j := Job{
		ID:         job.ID,
		Name:       job.Name,
		ServerID:   job.ServerID,
		Cron:       job.Cron,
		LogResults: job.LogResults,
		Steps:      make([]Step, len(job.Steps)),
	}
	for i, s := range job.Steps {
		j.Steps[i] = Step{
			Command: s.Command,
			Wait:    s.Wait,
		}
	}
	return j
}