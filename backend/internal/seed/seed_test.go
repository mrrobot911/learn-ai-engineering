package seed

import "testing"

func TestReportCreated(t *testing.T) {
	r := Report{TopicsInserted: 1, LessonsSkipped: 2}
	if got := r.Created(); got != 1 {
		t.Fatalf("Created() = %d, want 1", got)
	}
	empty := Report{TopicsSkipped: 1, QuestionsSkipped: 6}
	if got := empty.Created(); got != 0 {
		t.Fatalf("Created() = %d, want 0", got)
	}
}
