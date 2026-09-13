package models

import "time"

type ExerciseType string

const (
	ExerciseTypeDiscussion        ExerciseType = "discussion"
	ExerciseTypeDiscussionHandsOn ExerciseType = "discussion_and_hands_on"
)

type ExerciseCreate struct {
	Type               ExerciseType `json:"type"`
	ScheduledStartTime time.Time    `json:"scheduled_start_time"`
	ScheduledEndTime   time.Time    `json:"scheduled_end_time"`
	Title              string       `json:"title"`
	Scenario           string       `json:"scenario"`
}
