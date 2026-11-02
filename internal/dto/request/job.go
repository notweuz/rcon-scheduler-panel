package request

type CreateJob struct {
	Name       string `json:"name" binding:"required,min=3,max=50"`
	Cron       string `json:"cron" binding:"required,max=75"`
	LogResults bool   `json:"log_results"`
	Steps      []Step `json:"steps" binding:"required,min=1,dive"`
}

type UpdateJob = CreateJob

type Step struct {
	Command string `json:"command" binding:"required_without=Wait,excluded_with=Wait"`
	Wait    string `json:"wait" binding:"required_without=Command,excluded_with=Command"`
}
