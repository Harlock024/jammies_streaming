package models

type Track struct {
	ID       string  `json:"id" gorm:"primaryKey"`
	AudioUrl string  `json:"audio_url"`
	Duration float64 `json:"duration"`
}
