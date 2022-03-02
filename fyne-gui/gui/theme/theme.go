package theme

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/theme"
	"image/color"
)

type MyTheme struct {
	defaultTheme                                 fyne.Theme
	IsLight                                      binding.Bool
	Regular, Bold, Italic, BoldItalic, Monospace fyne.Resource
	SizeScale float32
}

func NewTheme() *MyTheme {
	isLight := false
	t := &MyTheme{
		IsLight: binding.BindBool(&isLight),
		SizeScale: 1.0,
	}
	t.SetDefaultFont()
	t.SetDark()
	return t
}

func (t *MyTheme) SetDark() {
	t.IsLight.Set(false)
	t.defaultTheme = theme.DarkTheme()
}
func (t *MyTheme) SetLight() {
	t.IsLight.Set(true)
	t.defaultTheme = theme.LightTheme()
}

func (t *MyTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	return t.defaultTheme.Color(name, variant)
}

func (t *MyTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return t.defaultTheme.Icon(name)
}

func (t *MyTheme) Font(style fyne.TextStyle) fyne.Resource {
	if style.Monospace {
		return t.Monospace
	}
	if style.Bold {
		if style.Italic {
			return t.BoldItalic
		}
		return t.Bold
	}
	if style.Italic {
		return t.Italic
	}
	return t.Regular
}

func (t *MyTheme) Size(name fyne.ThemeSizeName) float32 {
	return t.defaultTheme.Size(name)*t.SizeScale
}

func (t *MyTheme) SetDefaultFont() {
	t.Regular = theme.TextFont()
	t.Bold = theme.TextBoldFont()
	t.Italic = theme.TextItalicFont()
	t.BoldItalic = theme.TextBoldItalicFont()
	t.Monospace = theme.TextMonospaceFont()
}

//
//func (t *MyTheme) SetFontsFromAssets(regularFontPath string, monoFontPath string, onError func(err error)) {
//	t.Regular = theme.TextFont()
//	t.Bold = theme.TextBoldFont()
//	t.Italic = theme.TextItalicFont()
//	t.BoldItalic = theme.TextBoldItalicFont()
//	t.Monospace = theme.TextMonospaceFont()
//
//	if regularFontPath != "" {
//		t.Regular = loadCustomFontFromAssets(regularFontPath, "Regular", t.Regular, onError)
//		t.Bold = loadCustomFontFromAssets(regularFontPath, "Bold", t.Bold, onError)
//		t.Italic = loadCustomFontFromAssets(regularFontPath, "Italic", t.Italic, onError)
//		t.BoldItalic = loadCustomFontFromAssets(regularFontPath, "BoldItalic", t.BoldItalic, onError)
//	}
//	if monoFontPath != "" {
//		t.Monospace = loadCustomFontFromAssets(monoFontPath, "Regular", t.Monospace, onError)
//	} else {
//		t.Monospace = t.Regular
//	}
//}
//
//func loadCustomFontFromAssets(env, variant string, fallback fyne.Resource, onError func(err error)) fyne.Resource {
//
//	variantPath := strings.Replace(env, "Regular", variant, -1)
//	assets, err := utils.LoadFromAssets(variantPath, variantPath)
//	if err != nil {
//		onError(errRead)
//		return fallback
//	}
//
//	f, err := asset.Open(variantPath)
//	if err != nil {
//		onError(err)
//		return fallback
//	}
//
//	buf, errRead := ioutil.ReadAll(f)
//	f.Close()
//	if errRead != nil {
//		onError(errRead)
//		return fallback
//	}
//
//	return &fyne.StaticResource{
//		StaticName:    variantPath,
//		StaticContent: buf,
//	}
//}
