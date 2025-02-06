package portal

type GetTimelineRequest struct {
	Cursor string `validate:"omitempty,cursor"`
	Limit  int64  `validate:"omitempty,limit"`
}
