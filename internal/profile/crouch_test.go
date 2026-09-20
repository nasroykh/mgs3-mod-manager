package profile

import "testing"

func TestCrouchMetadataAndTargetClassification(t *testing.T) {
	files := CrouchFiles()
	if len(files) != 119 {
		t.Fatalf("CrouchFiles count = %d, want 119", len(files))
	}
	if len(files) == 0 {
		t.Fatal("empty metadata")
	}
	copyFiles := CrouchFiles()
	copyFiles[0].Target = "tampered"
	if got, ok := CrouchFileFor(files[0].Target); !ok || got.Target != files[0].Target {
		t.Fatal("CrouchFiles did not return a read-only copy")
	}
	for _, file := range files {
		if _, ok := CrouchFileFor(file.Target); !ok {
			t.Fatalf("lookup failed for %q", file.Target)
		}
		if file.OriginalAbsent {
			if err := AddedTarget(file.Target); err != nil {
				t.Errorf("absent target %q rejected: %v", file.Target, err)
			}
			if err := ReplacementTarget(file.Target); err == nil {
				t.Errorf("absent target %q accepted as replacement", file.Target)
			}
		} else {
			if err := ReplacementTarget(file.Target); err != nil {
				t.Errorf("present target %q rejected: %v", file.Target, err)
			}
			if err := AddedTarget(file.Target); err == nil {
				t.Errorf("present target %q accepted as added", file.Target)
			}
		}
	}
	if err := Target(files[0].Target); err == nil {
		t.Fatal("Target accepted crouch asset target; texture policy changed")
	}
}
