package httpapi

import (
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/MaaXYZ/MaaDebugger/internal/response"
)

// supportedLocaleBases 将 Accept-Language 中的基础语言映射到界面 locale。
// 新增界面语言时在此注册即可，例如 "ja": "ja"。
var supportedLocaleBases = map[string]string{
	"zh": "zh-CN",
	"en": "en",
}

// defaultLocale 是检测失败或语言不受支持时的回退语言。
const defaultLocale = "en"

type localeResponse struct {
	Locale string `json:"locale"` // 生效的界面语言，如 "zh-CN" / "en"
	Source string `json:"source"` // "user"（用户在配置中显式指定）或 "detected"（按请求头检测）
}

// handleLocale 返回当前生效的界面语言：
//  1. 若配置中保存了用户的显式选择（config["locale"]，可为 "zh-CN"/"en" 或 {"choice": "..."}），优先返回该值；
//  2. 否则根据请求的 Accept-Language 头检测用户语言；
//  3. 检测失败、语言不受支持或值为 "auto" 时回退到英语。
func (r *router) handleLocale(w http.ResponseWriter, req *http.Request) {
	cfgValue, _ := r.deps.ConfigStore.Get("locale")
	if userLocale, ok := readUserLocaleChoice(cfgValue); ok {
		response.OK(w, localeResponse{Locale: userLocale, Source: "user"})
		return
	}

	detected := detectLocaleFromHeader(req.Header.Get("Accept-Language"))
	response.OK(w, localeResponse{Locale: detected, Source: "detected"})
}

// readUserLocaleChoice 从配置值中解析用户显式选择的界面语言。
// 兼容两种存储形态：纯字符串（"zh-CN"）与 Pinia 持久化的对象（{"choice": "zh-CN"}）。
// "auto"、空值、不受支持的语言均视为未显式选择。
func readUserLocaleChoice(value any) (string, bool) {
	switch v := value.(type) {
	case string:
		return normalizeUserChoice(v)
	case map[string]any:
		if raw, ok := v["choice"]; ok {
			if s, ok := raw.(string); ok {
				return normalizeUserChoice(s)
			}
		}
	}
	return "", false
}

func normalizeUserChoice(choice string) (string, bool) {
	choice = strings.TrimSpace(choice)
	if choice == "" || strings.EqualFold(choice, "auto") {
		return "", false
	}
	// 显式选择需是受支持的完整 locale（大小写不敏感）。
	lower := strings.ToLower(choice)
	for _, supported := range supportedLocaleBases {
		if lower == strings.ToLower(supported) {
			return supported, true
		}
	}
	return "", false
}

// detectLocaleFromHeader 解析 Accept-Language 头并返回第一个受支持语言的 locale，
// 找不到受支持语言时回退到 defaultLocale。
func detectLocaleFromHeader(header string) string {
	for _, tag := range parseAcceptLanguage(header) {
		if locale, ok := supportedLocaleBases[tag.base]; ok {
			return locale
		}
	}
	return defaultLocale
}

type languageQuality struct {
	base  string
	q     float64
	order int
}

// parseAcceptLanguage 按 RFC 7231 解析 Accept-Language 头：
// 逗号分隔的 language-range，每个可带 ";q=" 权重；结果按权重降序稳定排序。
// 解析失败的片段会被忽略（相当于 q=0）。
func parseAcceptLanguage(header string) []languageQuality {
	header = strings.TrimSpace(header)
	if header == "" {
		return nil
	}

	var out []languageQuality
	for order, part := range strings.Split(header, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		tag := part
		q := 1.0
		if idx := strings.Index(part, ";"); idx >= 0 {
			tag = strings.TrimSpace(part[:idx])
			q = parseQuality(strings.TrimSpace(part[idx+1:]))
		}
		if q <= 0 {
			continue
		}

		base := strings.ToLower(strings.SplitN(tag, "-", 2)[0])
		if base == "" {
			continue
		}
		out = append(out, languageQuality{base: base, q: q, order: order})
	}

	sort.SliceStable(out, func(i, j int) bool { return out[i].q > out[j].q })
	return out
}

func parseQuality(raw string) float64 {
	if !strings.HasPrefix(raw, "q=") {
		return 0
	}
	q, err := strconv.ParseFloat(strings.TrimSpace(raw[2:]), 64)
	if err != nil {
		return 0
	}
	if q < 0 || q > 1 {
		return 0
	}
	return q
}
