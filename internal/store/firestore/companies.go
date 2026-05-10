package firestore

import (
	"context"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/kusakari/itsumo/internal/domain"
)

const companiesBatchSize = 500

// UpsertCompanies は企業マスタを code をキーに upsert します。
func (r *Repo) UpsertCompanies(ctx context.Context, companies []*domain.Company) error {
	now := time.Now()
	for i := 0; i < len(companies); i += companiesBatchSize {
		end := i + companiesBatchSize
		if end > len(companies) {
			end = len(companies)
		}
		batch := r.client.Batch()
		for _, c := range companies[i:end] {
			if c.Code == "" {
				continue
			}
			c.UpdatedAt = now
			ref := r.client.Collection("companies").Doc(c.Code)
			batch.Set(ref, c)
		}
		if _, err := batch.Commit(ctx); err != nil {
			return err
		}
	}
	return nil
}

// ListCompanyIndustries は companies から業界一覧を返します。
func (r *Repo) ListCompanyIndustries(ctx context.Context) ([]string, error) {
	docs, err := r.client.Collection("companies").Documents(ctx).GetAll()
	if err != nil {
		return nil, err
	}
	set := map[string]struct{}{}
	for _, doc := range docs {
		c := &domain.Company{}
		if err := doc.DataTo(c); err != nil {
			return nil, err
		}
		if c.Industry == "" {
			continue
		}
		set[c.Industry] = struct{}{}
	}
	industries := make([]string, 0, len(set))
	for s := range set {
		industries = append(industries, s)
	}
	sort.Strings(industries)
	return industries, nil
}

// ListCompanies は業界とキーワードで企業を絞り込みます。
func (r *Repo) ListCompanies(ctx context.Context, industry, keyword string, limit int) ([]*domain.Company, error) {
	industry = strings.TrimSpace(industry)
	keyword = strings.TrimSpace(keyword)
	if industry == "" && keyword == "" {
		return []*domain.Company{}, nil
	}

	q := r.client.Collection("companies").Query
	if industry != "" {
		q = q.Where("industry", "==", industry)
	}
	docs, err := q.Documents(ctx).GetAll()
	if err != nil {
		return nil, err
	}

	keywordNorm := normalizeSearchText(keyword)
	keywordRoman := romanizedKeyword(keyword)
	out := make([]*domain.Company, 0, len(docs))
	for _, doc := range docs {
		c := &domain.Company{}
		if err := doc.DataTo(c); err != nil {
			return nil, err
		}
		if keyword != "" {
			nameNorm := normalizeSearchText(c.Name)
			codeNorm := normalizeSearchText(c.Code)
			nameRoman := romanizeKana(c.Name)

			matchNorm := keywordNorm != "" && (strings.Contains(nameNorm, keywordNorm) || strings.Contains(codeNorm, keywordNorm))
			matchRoman := keywordRoman != "" && strings.Contains(nameRoman, keywordRoman)
			if !matchNorm && !matchRoman {
				continue
			}
		}
		out = append(out, c)
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].Name == out[j].Name {
			return out[i].Code < out[j].Code
		}
		return out[i].Name < out[j].Name
	})

	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

func normalizeSearchText(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	b := strings.Builder{}
	for _, r := range s {
		r = katakanaToHiragana(r)
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func katakanaToHiragana(r rune) rune {
	if r >= 'ァ' && r <= 'ヶ' {
		return r - 0x60
	}
	return r
}

func romanizedKeyword(s string) string {
	if x := romanizeKana(s); x != "" {
		return x
	}
	return normalizeASCII(s)
}

func normalizeASCII(s string) string {
	s = strings.ToLower(s)
	b := strings.Builder{}
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func romanizeKana(s string) string {
	h := []rune(normalizeSearchText(s))
	if len(h) == 0 {
		return ""
	}

	mono := map[rune]string{
		'あ': "a", 'い': "i", 'う': "u", 'え': "e", 'お': "o",
		'か': "ka", 'き': "ki", 'く': "ku", 'け': "ke", 'こ': "ko",
		'さ': "sa", 'し': "shi", 'す': "su", 'せ': "se", 'そ': "so",
		'た': "ta", 'ち': "chi", 'つ': "tsu", 'て': "te", 'と': "to",
		'な': "na", 'に': "ni", 'ぬ': "nu", 'ね': "ne", 'の': "no",
		'は': "ha", 'ひ': "hi", 'ふ': "fu", 'へ': "he", 'ほ': "ho",
		'ま': "ma", 'み': "mi", 'む': "mu", 'め': "me", 'も': "mo",
		'や': "ya", 'ゆ': "yu", 'よ': "yo",
		'ら': "ra", 'り': "ri", 'る': "ru", 'れ': "re", 'ろ': "ro",
		'わ': "wa", 'を': "wo", 'ん': "n",
		'が': "ga", 'ぎ': "gi", 'ぐ': "gu", 'げ': "ge", 'ご': "go",
		'ざ': "za", 'じ': "ji", 'ず': "zu", 'ぜ': "ze", 'ぞ': "zo",
		'だ': "da", 'ぢ': "ji", 'づ': "zu", 'で': "de", 'ど': "do",
		'ば': "ba", 'び': "bi", 'ぶ': "bu", 'べ': "be", 'ぼ': "bo",
		'ぱ': "pa", 'ぴ': "pi", 'ぷ': "pu", 'ぺ': "pe", 'ぽ': "po",
		'ぁ': "a", 'ぃ': "i", 'ぅ': "u", 'ぇ': "e", 'ぉ': "o",
	}

	combo := map[string]string{
		"きゃ": "kya", "きゅ": "kyu", "きょ": "kyo",
		"しゃ": "sha", "しゅ": "shu", "しょ": "sho",
		"ちゃ": "cha", "ちゅ": "chu", "ちょ": "cho",
		"にゃ": "nya", "にゅ": "nyu", "にょ": "nyo",
		"ひゃ": "hya", "ひゅ": "hyu", "ひょ": "hyo",
		"みゃ": "mya", "みゅ": "myu", "みょ": "myo",
		"りゃ": "rya", "りゅ": "ryu", "りょ": "ryo",
		"ぎゃ": "gya", "ぎゅ": "gyu", "ぎょ": "gyo",
		"じゃ": "ja", "じゅ": "ju", "じょ": "jo",
		"びゃ": "bya", "びゅ": "byu", "びょ": "byo",
		"ぴゃ": "pya", "ぴゅ": "pyu", "ぴょ": "pyo",
	}

	isSmall := func(r rune) bool {
		switch r {
		case 'ゃ', 'ゅ', 'ょ', 'ぁ', 'ぃ', 'ぅ', 'ぇ', 'ぉ':
			return true
		default:
			return false
		}
	}

	b := strings.Builder{}
	geminate := false
	for i := 0; i < len(h); {
		ch := h[i]
		if ch == 'っ' {
			geminate = true
			i++
			continue
		}
		if ch == 'ー' {
			i++
			continue
		}

		syl := ""
		if i+1 < len(h) && isSmall(h[i+1]) {
			if v, ok := combo[string([]rune{ch, h[i+1]})]; ok {
				syl = v
				i += 2
			}
		}
		if syl == "" {
			if v, ok := mono[ch]; ok {
				syl = v
				i++
			} else {
				i++
				continue
			}
		}

		if geminate && len(syl) > 0 {
			syl = string(syl[0]) + syl
			geminate = false
		}
		b.WriteString(syl)
	}

	return b.String()
}
