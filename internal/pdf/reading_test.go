package pdf

import "testing"

func TestUnwrapExtractedText(t *testing.T) {
	t.Parallel()

	input := "SUMMARY\n\n" +
		"Gut microbiota may influence antidepressant treatment outcomes, yet whether targeted modulation can\n" +
		"enhance efficacy remains unclear. We conducted a randomized, double-\n" +
		"blind, placebo-controlled trial.\n\n" +
		"A short first line\n" +
		"continues as the same paragraph.\n\n" +
		"A paragraph with enough words to establish the usual physical line width in this block\n" +
		"and a deliberately short ending.\n" +
		"Another paragraph begins after the short completed line.\n\n" +
		"INTRODUCTION\n" +
		"Major depressive disorder is influenced by genetic predisposition, gender disparities,\n\n" +
		"childhood trauma, substance misuse, and situational stressors.\n\n" +
		"A complete paragraph.\n\n" +
		"lowercase text that starts a distinct paragraph.\n\n" +
		"A lead-in:\n\n" +
		"lowercase item that must remain separate.\n\n" +
		"Highlights\n" +
		"• First finding\n" +
		"• Second finding\n"

	want := "SUMMARY\n\n" +
		"Gut microbiota may influence antidepressant treatment outcomes, yet whether targeted modulation can enhance efficacy remains unclear. We conducted a randomized, double-blind, placebo-controlled trial.\n\n" +
		"A short first line continues as the same paragraph.\n\n" +
		"A paragraph with enough words to establish the usual physical line width in this block and a deliberately short ending.\n" +
		"Another paragraph begins after the short completed line.\n\n" +
		"INTRODUCTION\n" +
		"Major depressive disorder is influenced by genetic predisposition, gender disparities, childhood trauma, substance misuse, and situational stressors.\n\n" +
		"A complete paragraph.\n\n" +
		"lowercase text that starts a distinct paragraph.\n\n" +
		"A lead-in:\n\n" +
		"lowercase item that must remain separate.\n\n" +
		"Highlights\n" +
		"• First finding\n" +
		"• Second finding\n"

	if got := unwrapExtractedText(input); got != want {
		t.Fatalf("unwrapExtractedText() mismatch\nwant: %q\n got: %q", want, got)
	}
}

func TestUnwrapExtractedTextPreservesPageBreakOnItsOwnLine(t *testing.T) {
	t.Parallel()

	got := unwrapExtractedText("First page line\ncontinues.\fSecond page line\ncontinues.\n")
	want := "First page line continues.\n\f\nSecond page line continues.\n"
	if got != want {
		t.Fatalf("unwrapExtractedText() page break mismatch\nwant: %q\n got: %q", want, got)
	}
}
