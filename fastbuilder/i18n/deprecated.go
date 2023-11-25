package I18n

// 此表标记了已弃用的语言
var DeprecatedLanguages []string = []string{
	LanguageSimplifiedChinese,
	LanguageTraditionalChinese,
	LanguageTaiwanChinese,
}

// 确定代号为 langeuage_name 的语言是否被弃用。
// 若弃用，返回真，否则返回假
func IsDeprecated(language_name string) (has bool) {
	for _, value := range DeprecatedLanguages {
		if value == language_name {
			return true
		}
	}
	return false
}
