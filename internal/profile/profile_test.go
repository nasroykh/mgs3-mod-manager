package profile

import "testing"

func TestPathPolicy(t *testing.T) {
	valid := []string{
		"textures/flatlist/_win/sna_def_olive.bmp.ctxr",
		"hqtex/flatlist/_win/sna_def_olive.bmp.ctxr",
	}
	for _, path := range valid {
		if err := Target(path); err != nil {
			t.Fatalf("Target(%q): %v", path, err)
		}
	}
	invalid := []string{
		"textures\\flatlist\\_win\\x.ctxr",
		"textures/flatlist/_win/../x.ctxr",
		"textures/flatlist/_win/x.ctxr:evil",
		"textures/flatlist/_win/CON.ctxr",
		"textures/other/x.ctxr",
		"textures/flatlist/_win/x.dds",
	}
	for _, path := range invalid {
		if err := Target(path); err == nil {
			t.Errorf("Target(%q) unexpectedly succeeded", path)
		}
	}
	if got := Key(`HQTEX\FLATLIST\_WIN\X.CTX`); got != "hqtex/flatlist/_win/x.ctx" {
		t.Fatalf("Key returned %q", got)
	}
}

func TestPluginTargetPolicy(t *testing.T) {
	for _, path := range []string{"dinput8.asi", "plugin-1.asi"} {
		if err := PluginTarget(path); err != nil {
			t.Errorf("PluginTarget(%q): %v", path, err)
		}
	}
	for _, path := range []string{".asi", "DINPUT8.asi", "dinput8.ASI", "plugins/dinput8.asi", "../dinput8.asi", "dinput8.asi/"} {
		if err := PluginTarget(path); err == nil {
			t.Errorf("PluginTarget(%q) unexpectedly succeeded", path)
		}
	}
}
