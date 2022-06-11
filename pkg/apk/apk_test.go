package apk

import "testing"

func TestAPKOpening(t *testing.T) {
	apk, err := NewAPK("testdata/helloworld.apk")
	if err != nil {
		t.Fatal(err)
	}

	defer apk.Close()

	if apk.Identifier() != "com.example.helloworld" {
		t.Errorf("PackageName: expected %s, got %s", "com.example.helloworld", apk.Identifier())
	}

	if apk.Label() != "HelloWorld" {
		t.Errorf("Label: expected %s, got %s", "Hello World", apk.Label())
	}

	if apk.VersionCode() != 1 {
		t.Errorf("VersionCode: expected %d, got %d", 1, apk.VersionCode())
	}

	if apk.VersionName() != "1.0" {
		t.Errorf("VersionName: expected %s, got %s", "1.0", apk.VersionName())
	}

	for dpi := ScreenLDPI; dpi <= ScreenXXXHDPI; dpi++ {
		icon, err := apk.Icon(dpi)
		if err != nil {
			t.Errorf("Icon: %s", err)
		}

		if icon == nil {
			t.Errorf("Icon: expected non-nil %s image, got nil", dpi)
		} else {
			var iconSizes = map[ScreenDPI]int{
				ScreenLDPI:    36,
				ScreenMDPI:    48,
				ScreenHDPI:    72,
				ScreenXHDPI:   96,
				ScreenXXHDPI:  144,
				ScreenXXXHDPI: 192,
			}

			// check icon dimensions
			if icon.Bounds().Dx() != iconSizes[dpi] {
				t.Errorf("Icon: expected %s image size %dpx, got %dpx", dpi, iconSizes[dpi], icon.Bounds().Dx())
			}
		}
	}
}
