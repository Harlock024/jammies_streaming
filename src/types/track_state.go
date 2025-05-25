package types

type PlayTrackState struct {
	Event       string  `json:"event"`
	TrackID     string  `json:"track_id"`
	CurrentTime float64 `json:"current_time"`
}

type PlayTrackGetState struct {
	Event       string  `json:"event"`
	TrackID     string  `json:"track_id"`
	AudioURL    string  `json:"audio_url"`
	CurrentTime float64 `json:"current_time"`
}
