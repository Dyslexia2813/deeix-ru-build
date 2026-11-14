package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base32"
	"regexp"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Label types. Labels are ASCII so that they survive JSON escaping and tokenisation intact:
// [FIO_k3m9q], [INN_ab2cd] ... The suffix is an HMAC of the normalised value, so the same value
// gets the same label in every request of a chat without the proxy keeping any state.
const (
	typeFIO     = "FIO"
	typeTel     = "TEL"
	typeEmail   = "EMAIL"
	typeINN     = "INN"
	typeSNILS   = "SNILS"
	typeOGRN    = "OGRN"
	typePasport = "PASPORT"
	typeCard    = "KARTA"
	typeAccount = "SCHET"
	typeBirth   = "DR"
	typeAddress = "ADRES"
	typeOther   = "PD" // values the user asked to hide in the preview
)

var labelPattern = regexp.MustCompile(`\[(FIO|TEL|EMAIL|INN|SNILS|OGRN|PASPORT|KARTA|SCHET|DR|ADRES|PD)_[a-z2-7]{5}\]`)

// labelPrefixPattern matches an unfinished label at the end of a streamed chunk.
var labelPrefixPattern = regexp.MustCompile(`\[[A-Z]{0,7}(?:_[a-z2-7]{0,5})?$`)

const maxLabelLen = len("[PASPORT_abcde]")

type span struct {
	start, end int
	kind       string
	key        string // normalised value: same key, same label
}

// Masker replaces personal data with labels and remembers label -> original for one request.
type Masker struct {
	secret      []byte
	enabled     map[string]bool
	values      map[string]string // label -> original text (longest form seen)
	counts      map[string]int    // type -> replacements
	labelCounts map[string]int    // label -> replacements
	labelKinds  map[string]string // label -> type
	rules       *Rules
	kept        map[string]string // label -> value left as is because of rules.Keep
	keptCounts  map[string]int
	labelManual map[string]string // label -> user's Hide word
}

func NewMasker(secret []byte, enabled map[string]bool) *Masker {
	return &Masker{secret: secret, enabled: enabled, values: map[string]string{}, counts: map[string]int{},
		labelCounts: map[string]int{}, labelKinds: map[string]string{}, kept: map[string]string{}, keptCounts: map[string]int{}, labelManual: map[string]string{}}
}

func (m *Masker) Total() int {
	total := 0
	for _, n := range m.counts {
		total += n
	}
	return total
}

func (m *Masker) label(kind, key string) string {
	mac := hmac.New(sha256.New, m.secret)
	mac.Write([]byte(kind + "|" + key))
	sum := base32.StdEncoding.EncodeToString(mac.Sum(nil))
	return "[" + kind + "_" + strings.ToLower(sum[:5]) + "]"
}

// Mask returns text with personal data replaced by labels.
func (m *Masker) Mask(text string) string {
	if text == "" {
		return text
	}
	spans := m.spans(text)
	if len(spans) == 0 {
		return text
	}
	var out strings.Builder
	out.Grow(len(text))
	last := 0
	for _, s := range spans {
		if m.enabled != nil && !m.enabled[s.kind] && s.kind != typeOther {
			continue
		}
		original := text[s.start:s.end]
		lbl := m.label(s.kind, s.key)
		if m.rules.keeps(s, original) {
			if prev, ok := m.kept[lbl]; !ok || displayScore(original) > displayScore(prev) {
				m.kept[lbl] = original
			}
			m.keptCounts[lbl]++
			m.labelKinds[lbl] = s.kind
			continue
		}
		m.labelCounts[lbl]++
		m.labelKinds[lbl] = s.kind
		if word := m.rules.hideWord(s); word != "" {
			m.labelManual[lbl] = word
		}
		if prev, ok := m.values[lbl]; !ok || displayScore(original) > displayScore(prev) {
			m.values[lbl] = original
		}
		m.counts[s.kind]++
		out.WriteString(text[last:s.start])
		out.WriteString(lbl)
		last = s.end
	}
	out.WriteString(text[last:])
	return out.String()
}

// Unmask puts the original values back. jsonString escapes them for the inside of a JSON string.
func (m *Masker) Unmask(text string, jsonString bool) string {
	if len(m.values) == 0 || !strings.Contains(text, "[") {
		return text
	}
	return labelPattern.ReplaceAllStringFunc(text, func(lbl string) string {
		value, ok := m.values[lbl]
		if !ok {
			return lbl
		}
		if jsonString {
			return jsonEscapeInner(value)
		}
		return value
	})
}

// displayScore prefers the fullest, normally written form of a value for putting it back:
// «Иванов Иван Иванович» over «Иванов И.И.» and over «ИВАНОВУ ИВАНУ ИВАНОВИЧУ».
func displayScore(value string) int {
	score := len(strings.Fields(value)) * 10
	if strings.ToUpper(value) != value {
		score += 5
	}
	return score
}

func jsonEscapeInner(value string) string {
	var b strings.Builder
	for _, r := range value {
		switch r {
		case '"':
			b.WriteString(`\"`)
		case '\\':
			b.WriteString(`\\`)
		case '\n':
			b.WriteString(`\n`)
		case '\r':
			b.WriteString(`\r`)
		case '\t':
			b.WriteString(`\t`)
		default:
			if r < 0x20 {
				b.WriteString(`\u00`)
				b.WriteByte("0123456789abcdef"[r>>4])
				b.WriteByte("0123456789abcdef"[r&0xf])
			} else {
				b.WriteRune(r)
			}
		}
	}
	return b.String()
}

// ---------------------------------------------------------------------------
// Detectors
// ---------------------------------------------------------------------------

const (
	capWord = `[А-ЯЁ][а-яё]+(?:-[А-ЯЁ][а-яё]+)?`
	patrM   = `[А-ЯЁ][а-яё]+(?:ович|евич|ьич|ич)(?:а|у|ем|е)?`
	patrF   = `[А-ЯЁ][а-яё]+(?:овн|евн|ичн|инишн)(?:а|ы|е|у|ой|ою)`
	patr    = `(?:` + patrM + `|` + patrF + `)`
	initial = `[А-ЯЁ]\.\s?`
)

var (
	reFIOSurnameFirst  = regexp.MustCompile(`(` + capWord + `)\s+(` + capWord + `)\s+(` + patr + `)`)
	reFIONameFirst     = regexp.MustCompile(`(` + capWord + `)\s+(` + patr + `)\s+(` + capWord + `)`)
	reFIONamePatr      = regexp.MustCompile(`(` + capWord + `)\s+(` + patr + `)`)
	reFIOInitialsLast  = regexp.MustCompile(`(` + capWord + `)\s+([А-ЯЁ])\.\s?([А-ЯЁ])\.`)
	reFIOInitialsFirst = regexp.MustCompile(`([А-ЯЁ])\.\s?([А-ЯЁ])\.\s?(` + capWord + `)`)

	reEmail       = regexp.MustCompile(`[A-Za-z0-9._%+\-]+@[A-Za-z0-9\-]+(?:\.[A-Za-z0-9\-]+)*\.[A-Za-z]{2,}`)
	reTel         = regexp.MustCompile(`(?:\+7|8)[\s\-]?\(?\d{3}\)?[\s\-]?\d{3}[\s\-]?\d{2}[\s\-]?\d{2}`)
	reCard        = regexp.MustCompile(`\d{4}[ \-]?\d{4}[ \-]?\d{4}[ \-]?\d{4}(?:[ \-]?\d{1,3})?`)
	reAcct        = regexp.MustCompile(`\d{5}[ .]?\d{3}[ .]?\d[ .]?\d{4}[ .]?\d{7}`)
	reSNILS       = regexp.MustCompile(`\d{3}[\- ]\d{3}[\- ]\d{3}[\- ]\d{2}`)
	reSNILSCtx    = regexp.MustCompile(`(?i)снилс[^\d\n]{0,20}(\d{11})`)
	reDigits      = regexp.MustCompile(`\d{10,15}`)
	rePasport     = regexp.MustCompile(`(?i)(?:паспорт[а-яё]*|сери[яиюей])[^\d\n]{0,40}?(\d{2}\s?\d{2}\s*(?:№|N|номер)?\s*\d{6})`)
	reDeptCode    = regexp.MustCompile(`(?i)код[а-яё]*\s+подразделени[яе][^\d\n]{0,5}(\d{3}-\d{3})`)
	reBirthCtx    = regexp.MustCompile(`(?i)(?:дата\s+рождения|д\.\s?р\.|родил(?:ся|ась)|рожд[её]н(?:ный|ная|ия)?)[^\d\n]{0,15}(\d{1,2}[./\-]\d{1,2}[./\-]\d{2,4}|\d{1,2}\s+[а-яё]+\s+\d{4})`)
	reBirthSuffix = regexp.MustCompile(`(\d{1,2}\.\d{1,2}\.\d{4})\s*(?:г\.\s*)?(?:г\.\s?р\.|года\s+рождения)`)
	reAddress     = regexp.MustCompile(`(?i)(?:ул\.|улица|пр-т|пр\.|проспект|пер\.|переулок|ш\.|шоссе|б-р|бульвар|наб\.|набережная|пл\.|площадь|мкр\.?|микрорайон|проезд)\s*[^,\n]{2,40},\s*(?:д\.|дом)\s*\d+[а-яА-Я]?(?:\s*/\s*\d+)?(?:,?\s*(?:корп\.|корпус|к\.|стр\.|строение)\s*\d+[а-яА-Я]?)?(?:,?\s*(?:кв\.|квартира|оф\.|офис|пом\.)\s*\d+)?`)
)

func isWordRune(r rune) bool { return unicode.IsLetter(r) || unicode.IsDigit(r) }

func boundaryOK(text string, start, end int, digits bool) bool {
	if start > 0 {
		r, _ := utf8.DecodeLastRuneInString(text[:start])
		if digits && unicode.IsDigit(r) || !digits && isWordRune(r) {
			return false
		}
	}
	if end < len(text) {
		r, _ := utf8.DecodeRuneInString(text[end:])
		if digits && unicode.IsDigit(r) || !digits && isWordRune(r) {
			return false
		}
	}
	return true
}

func onlyDigits(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// titleShadow returns a copy of text where ALL-CAPS Cyrillic words are title-cased, so that
// «ИВАНОВ ИВАН ИВАНОВИЧ» is found by the same patterns. Cyrillic case mapping keeps byte length.
func titleShadow(text string) string {
	b := []rune(text)
	changed := false
	for i := 0; i < len(b); {
		if !isCyrUpper(b[i]) {
			i++
			continue
		}
		j := i
		for j < len(b) && (isCyrUpper(b[j]) || b[j] == '-') {
			j++
		}
		prevLetter := i > 0 && unicode.IsLetter(b[i-1])
		nextLetter := j < len(b) && unicode.IsLetter(b[j])
		if j-i >= 3 && !prevLetter && !nextLetter {
			for k := i + 1; k < j; k++ {
				if b[k] != '-' && b[k-1] != '-' {
					b[k] = unicode.ToLower(b[k])
				}
			}
			changed = true
		}
		i = j
	}
	if !changed {
		return text
	}
	return string(b)
}

func isCyrUpper(r rune) bool { return (r >= 'А' && r <= 'Я') || r == 'Ё' }

var nameEnding = regexp.MustCompile(`(?:ой|ей|ем|ом|ию|ии|ия|а|я|у|ю|е|ы|и|ь)$`)

var surnameEnding = regexp.MustCompile(`(?:ыми|ими|ого|его|ому|ему|ой|ей|ым|им|ом|ем|ою|ая|ую|ий|ый|а|у|е|ы|я|ю)$`)

func surnameStem(word string) string {
	w := strings.ToLower(strings.ReplaceAll(word, "ё", "е"))
	if utf8.RuneCountInString(w) <= 3 {
		return w
	}
	return surnameEnding.ReplaceAllString(w, "")
}

func firstLetter(word string) string {
	r, _ := utf8.DecodeRuneInString(strings.ToLower(word))
	if r == 'ё' {
		r = 'е'
	}
	return string(r)
}

func patrStem(word string) string {
	w := strings.ToLower(word)
	for _, suffix := range []string{"овичем", "евичем", "овичу", "евичу", "овича", "евича", "овиче", "евиче", "ович", "евич",
		"овной", "евной", "овною", "овна", "овны", "овне", "овну", "евна", "евны", "евне", "евну", "ичной", "ична", "ичны", "ичне", "ичну",
		"ичем", "ичу", "ича", "иче", "ич"} {
		if strings.HasSuffix(w, suffix) {
			return strings.TrimSuffix(w, suffix)
		}
	}
	return w
}

func nameStem(word string) string {
	w := strings.ToLower(strings.ReplaceAll(word, "ё", "е"))
	if utf8.RuneCountInString(w) <= 3 {
		return w
	}
	return nameEnding.ReplaceAllString(w, "")
}

// notNames are capitalised words that often stand before a name but are not part of it.
var notNames = map[string]bool{
	"гражданин": true, "гражданка": true, "господин": true, "госпожа": true, "уважаемый": true, "уважаемая": true,
	"директор": true, "генеральный": true, "заказчик": true, "исполнитель": true, "подрядчик": true, "поставщик": true,
	"покупатель": true, "продавец": true, "арендатор": true, "арендодатель": true, "истец": true, "ответчик": true,
	"представитель": true, "доверитель": true, "поверенный": true, "работник": true, "работодатель": true,
	"российской": true, "российская": true, "федерации": true, "федерация": true, "москва": true, "москвы": true,
}

func findFIO(text string) []span {
	shadow := titleShadow(text)
	var spans []span
	add := func(start, end int, key string) {
		if boundaryOK(text, start, end, false) {
			spans = append(spans, span{start: start, end: end, kind: typeFIO, key: key})
		}
	}
	for _, m := range reFIOSurnameFirst.FindAllStringSubmatchIndex(shadow, -1) {
		s, n, p := shadow[m[2]:m[3]], shadow[m[4]:m[5]], shadow[m[6]:m[7]]
		if notNames[strings.ToLower(s)] {
			// «Директор Иван Иванович»: name + patronymic only, handled below.
			continue
		}
		add(m[0], m[1], surnameStem(s)+"|"+firstLetter(n)+"|"+firstLetter(p))
	}
	for _, m := range reFIONameFirst.FindAllStringSubmatchIndex(shadow, -1) {
		n, p, s := shadow[m[2]:m[3]], shadow[m[4]:m[5]], shadow[m[6]:m[7]]
		if notNames[strings.ToLower(n)] || notNames[strings.ToLower(s)] {
			continue
		}
		add(m[0], m[1], surnameStem(s)+"|"+firstLetter(n)+"|"+firstLetter(p))
	}
	for _, m := range reFIONamePatr.FindAllStringSubmatchIndex(shadow, -1) {
		n, p := shadow[m[2]:m[3]], shadow[m[4]:m[5]]
		if notNames[strings.ToLower(n)] {
			continue
		}
		add(m[0], m[1], "~"+nameStem(n)+"|"+patrStem(p))
	}
	for _, m := range reFIOInitialsLast.FindAllStringSubmatchIndex(shadow, -1) {
		s := shadow[m[2]:m[3]]
		if notNames[strings.ToLower(s)] {
			continue
		}
		add(m[0], m[1], surnameStem(s)+"|"+firstLetter(shadow[m[4]:m[5]])+"|"+firstLetter(shadow[m[6]:m[7]]))
	}
	for _, m := range reFIOInitialsFirst.FindAllStringSubmatchIndex(shadow, -1) {
		s := shadow[m[6]:m[7]]
		if notNames[strings.ToLower(s)] {
			continue
		}
		add(m[0], m[1], surnameStem(s)+"|"+firstLetter(shadow[m[2]:m[3]])+"|"+firstLetter(shadow[m[4]:m[5]]))
	}
	return spans
}

func luhn(digits string) bool {
	sum := 0
	double := false
	for i := len(digits) - 1; i >= 0; i-- {
		d := int(digits[i] - '0')
		if double {
			d *= 2
			if d > 9 {
				d -= 9
			}
		}
		sum += d
		double = !double
	}
	return sum%10 == 0
}

func innValid(d string) bool {
	n := func(i int) int { return int(d[i] - '0') }
	check := func(weights []int) int {
		s := 0
		for i, w := range weights {
			s += w * n(i)
		}
		return s % 11 % 10
	}
	switch len(d) {
	case 10:
		return check([]int{2, 4, 10, 3, 5, 9, 4, 6, 8}) == n(9)
	case 12:
		return check([]int{7, 2, 4, 10, 3, 5, 9, 4, 6, 8}) == n(10) &&
			check([]int{3, 7, 2, 4, 10, 3, 5, 9, 4, 6, 8}) == n(11)
	}
	return false
}

func ogrnValid(d string) bool {
	var mod, divisor int64
	switch len(d) {
	case 13:
		divisor = 11
	case 15:
		divisor = 13
	default:
		return false
	}
	var num int64
	for _, r := range d[:len(d)-1] {
		num = num*10 + int64(r-'0')
	}
	mod = num % divisor % 10
	return mod == int64(d[len(d)-1]-'0')
}

func snilsValid(d string) bool {
	if len(d) != 11 {
		return false
	}
	sum := 0
	for i := 0; i < 9; i++ {
		sum += int(d[i]-'0') * (9 - i)
	}
	var check int
	switch {
	case sum < 100:
		check = sum
	case sum == 100 || sum == 101:
		check = 0
	default:
		check = sum % 101
		if check == 100 {
			check = 0
		}
	}
	return check == int(d[9]-'0')*10+int(d[10]-'0')
}

func findSpans(text string) []span {
	return resolveOverlaps(findRawSpans(text))
}

func findRawSpans(text string) []span {
	var spans []span
	addDigits := func(start, end int, kind, key string) {
		if boundaryOK(text, start, end, true) {
			spans = append(spans, span{start: start, end: end, kind: kind, key: key})
		}
	}
	for _, m := range reEmail.FindAllStringIndex(text, -1) {
		spans = append(spans, span{m[0], m[1], typeEmail, strings.ToLower(text[m[0]:m[1]])})
	}
	for _, m := range rePasport.FindAllStringSubmatchIndex(text, -1) {
		spans = append(spans, span{m[2], m[3], typePasport, onlyDigits(text[m[2]:m[3]])})
	}
	for _, m := range reDeptCode.FindAllStringSubmatchIndex(text, -1) {
		spans = append(spans, span{m[2], m[3], typePasport, "kod" + onlyDigits(text[m[2]:m[3]])})
	}
	for _, m := range reBirthCtx.FindAllStringSubmatchIndex(text, -1) {
		spans = append(spans, span{m[2], m[3], typeBirth, strings.ToLower(text[m[2]:m[3]])})
	}
	for _, m := range reBirthSuffix.FindAllStringSubmatchIndex(text, -1) {
		spans = append(spans, span{m[2], m[3], typeBirth, text[m[2]:m[3]]})
	}
	for _, m := range reCard.FindAllStringIndex(text, -1) {
		d := onlyDigits(text[m[0]:m[1]])
		if luhn(d) {
			addDigits(m[0], m[1], typeCard, d)
		}
	}
	for _, m := range reAcct.FindAllStringIndex(text, -1) {
		addDigits(m[0], m[1], typeAccount, onlyDigits(text[m[0]:m[1]]))
	}
	for _, m := range reSNILS.FindAllStringIndex(text, -1) {
		d := onlyDigits(text[m[0]:m[1]])
		if snilsValid(d) {
			addDigits(m[0], m[1], typeSNILS, d)
		}
	}
	for _, m := range reSNILSCtx.FindAllStringSubmatchIndex(text, -1) {
		addDigits(m[2], m[3], typeSNILS, text[m[2]:m[3]])
	}
	for _, m := range reTel.FindAllStringIndex(text, -1) {
		d := onlyDigits(text[m[0]:m[1]])
		addDigits(m[0], m[1], typeTel, d[len(d)-10:])
	}
	for _, m := range reDigits.FindAllStringIndex(text, -1) {
		d := text[m[0]:m[1]]
		switch {
		case (len(d) == 10 || len(d) == 12) && innValid(d):
			addDigits(m[0], m[1], typeINN, d)
		case (len(d) == 13 || len(d) == 15) && ogrnValid(d):
			addDigits(m[0], m[1], typeOGRN, d)
		}
	}
	for _, m := range reAddress.FindAllStringIndex(text, -1) {
		addr := text[m[0]:m[1]]
		spans = append(spans, span{m[0], m[1], typeAddress, strings.ToLower(strings.Join(strings.Fields(addr), " "))})
	}
	spans = append(spans, findFIO(text)...)
	return spans
}

// resolveOverlaps keeps the longest span among overlapping ones (earlier start on ties).
func resolveOverlaps(spans []span) []span {
	if len(spans) < 2 {
		return spans
	}
	sort.SliceStable(spans, func(i, j int) bool {
		li, lj := spans[i].end-spans[i].start, spans[j].end-spans[j].start
		if li != lj {
			return li > lj
		}
		return spans[i].start < spans[j].start
	})
	kept := make([]span, 0, len(spans))
	for _, s := range spans {
		overlap := false
		for _, k := range kept {
			if s.start < k.end && k.start < s.end {
				overlap = true
				break
			}
		}
		if !overlap {
			kept = append(kept, s)
		}
	}
	sort.Slice(kept, func(i, j int) bool { return kept[i].start < kept[j].start })
	return kept
}
