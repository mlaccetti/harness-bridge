package session

import "testing"

func TestApplySlice_sinceCompaction(t *testing.T) {
	sess := Session{
		Turns: []Turn{
			{Role: RoleUser, Text: "a"},
			{Role: RoleAssistant, Text: "b"},
			{Role: RoleUser, Text: "c"},
		},
		Compactions: []CompactionEvent{{TurnIndex: 2, Trigger: "auto"}},
	}
	out, rep := sess.ApplySlice(SliceOptions{SinceCompaction: true})
	if len(out.Turns) != 1 || out.Turns[0].Text != "c" {
		t.Fatalf("got %v", out.Turns)
	}
	if rep.StartIndex != 2 {
		t.Fatalf("start %d", rep.StartIndex)
	}
}

func TestApplySlice_lastTurns(t *testing.T) {
	sess := Session{
		Turns: []Turn{
			{Role: RoleUser, Text: "1"},
			{Role: RoleUser, Text: "2"},
			{Role: RoleUser, Text: "3"},
		},
	}
	out, _ := sess.ApplySlice(SliceOptions{LastTurns: 2})
	if len(out.Turns) != 2 || out.Turns[0].Text != "2" {
		t.Fatalf("got %v", out.Turns)
	}
}
