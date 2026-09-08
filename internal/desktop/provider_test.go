package desktop

import (
	"context"
	"github.com/fu9zhou/finish-bit/pkg/operation"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestScreenArgumentsWithoutRecording(t *testing.T) {
	args, seconds, e := recordArguments(operation.Request{Options: map[string]any{"seconds": 5, "width": 640, "height": 480, "x": -100, "cursor": false}})
	if e != nil || seconds != 5 || !strings.Contains(strings.Join(args, " "), "-offset_x -100 -offset_y 0 -video_size 640x480 -i desktop") {
		t.Fatalf("%v %d %v", args, seconds, e)
	}
	if _, _, e := recordArguments(operation.Request{Options: map[string]any{"width": 640}}); e == nil {
		t.Fatal("unpaired dimensions accepted")
	}
}
func TestInstalledSpeechWithoutPlayback(t *testing.T) {
	if runtime.GOOS != "windows" || os.Getenv("FINISHBIT_TEST_LOCAL_HOME") == "" {
		t.Skip("opt-in Windows speech test")
	}
	result, e := speak(context.Background(), "speech.voices", operation.Request{})
	if e != nil {
		t.Fatal(e)
	}
	voices := result.Data["voices"].([]map[string]any)
	if len(voices) == 0 {
		t.Skip("no installed voices")
	}
	out := filepath.Join(t.TempDir(), "speech.wav")
	_, e = speak(context.Background(), "speech.synthesize", operation.Request{Inputs: []string{"Finish bit local speech."}, Options: map[string]any{"output": out}})
	if e != nil {
		t.Fatal(e)
	}
	data, e := os.ReadFile(out)
	if e != nil || len(data) < 44 || string(data[:4]) != "RIFF" {
		t.Fatal("invalid WAV", e)
	}
}
