package main

import (
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Spreadsheet arguments of make_xlsx.
type xlsxArgs struct {
	FileName string      `json:"file_name"`
	Sheets   []sheetSpec `json:"sheets"`
}

type sheetSpec struct {
	Name         string              `json:"name"`
	Columns      []columnSpec        `json:"columns"`
	Rows         [][]json.RawMessage `json:"rows"`
	HeaderFill   string              `json:"header_fill"`
	FreezeHeader *bool               `json:"freeze_header"`
	Filter       *bool               `json:"filter"`
}

type columnSpec struct {
	Title  string  `json:"title"`
	Width  float64 `json:"width"`
	Format string  `json:"format"`
}

// cellSpec is the object form of a cell: {"value": ..., "fill": "#FFF2CC", "bold": true, "format": "money"}.
type cellSpec struct {
	Value  json.RawMessage `json:"value"`
	Fill   string          `json:"fill"`
	Bold   bool            `json:"bold"`
	Format string          `json:"format"`
}

const (
	xlsxMinWidth      = 6
	xlsxMaxWidth      = 60
	xlsxWrapFrom      = 50
	xlsxMaxPlainDigit = 9
	xlsxMaxDigits     = 15
	xlsxMaxRows       = 20000
	xlsxMaxColumns    = 200
)

var (
	numGrouped    = regexp.MustCompile(`^[+-]?\d{1,3}(?:[ \x{00a0}\x{202f}]\d{3})+(?:[.,]\d+)?$`)
	numPlain      = regexp.MustCompile(`^[+-]?\d+(?:[.,]\d+)?$`)
	numEnThousand = regexp.MustCompile(`^[+-]?\d{1,3}(?:,\d{3})+$`)
	dateISO       = regexp.MustCompile(`^(\d{4})-(\d{2})-(\d{2})$`)
	dateRU        = regexp.MustCompile(`^(\d{1,2})\.(\d{1,2})\.(\d{4})$`)
	hexColor      = regexp.MustCompile(`^#?([0-9A-Fa-f]{6})$`)
)

// parseNumber reads "1 234,56", "1234.5", "-7" as numbers. Anything unclear stays text:
// "1,234", "007", "12%", long digit strings (INN, accounts, phones).
func parseNumber(raw string) (float64, bool) {
	text := strings.TrimSpace(raw)
	if text == "" || len(text) > 32 || numEnThousand.MatchString(text) {
		return 0, false
	}
	grouped := numGrouped.MatchString(text)
	if !grouped && !numPlain.MatchString(text) {
		return 0, false
	}
	unsigned := strings.TrimLeft(text, "+-")
	if len(unsigned) > 1 && unsigned[0] == '0' && unsigned[1] >= '0' && unsigned[1] <= '9' {
		return 0, false
	}
	digits := 0
	for _, c := range unsigned {
		if c >= '0' && c <= '9' {
			digits++
		}
	}
	hasFraction := strings.ContainsAny(unsigned, ".,")
	if (!grouped && !hasFraction && digits > xlsxMaxPlainDigit) || digits > xlsxMaxDigits {
		return 0, false
	}
	clean := strings.NewReplacer(" ", "", " ", "", " ", "", ",", ".").Replace(text)
	value, err := strconv.ParseFloat(clean, 64)
	if err != nil || math.IsInf(value, 0) || math.IsNaN(value) {
		return 0, false
	}
	return value, true
}

// parseDate reads 2026-11-20 and 20.11.2026 into an Excel serial date.
func parseDate(raw string) (float64, bool) {
	text := strings.TrimSpace(raw)
	var y, m, d int
	if g := dateISO.FindStringSubmatch(text); g != nil {
		y, _ = strconv.Atoi(g[1])
		m, _ = strconv.Atoi(g[2])
		d, _ = strconv.Atoi(g[3])
	} else if g := dateRU.FindStringSubmatch(text); g != nil {
		d, _ = strconv.Atoi(g[1])
		m, _ = strconv.Atoi(g[2])
		y, _ = strconv.Atoi(g[3])
	} else {
		return 0, false
	}
	t := time.Date(y, time.Month(m), d, 0, 0, 0, 0, time.UTC)
	if t.Year() != y || int(t.Month()) != m || t.Day() != d || y < 1900 || y > 9999 {
		return 0, false
	}
	base := time.Date(1899, 12, 30, 0, 0, 0, 0, time.UTC)
	return t.Sub(base).Hours() / 24, true
}

func normalizeColor(value string) (string, bool) {
	g := hexColor.FindStringSubmatch(strings.TrimSpace(value))
	if g == nil {
		return "", false
	}
	return "FF" + strings.ToUpper(g[1]), true
}

func columnName(index int) string {
	name := ""
	for n := index + 1; n > 0; n = (n - 1) / 26 {
		name = string(rune('A'+(n-1)%26)) + name
	}
	return name
}

// Cell styles are collected while the sheets are written and become styles.xml.
type styleKey struct {
	Bold   bool
	Fill   string
	Wrap   bool
	NumFmt int
	Border bool
}

type styleBook struct {
	keys  []styleKey
	index map[styleKey]int
	fills []string
}

func newStyleBook() *styleBook {
	b := &styleBook{index: map[styleKey]int{}}
	b.get(styleKey{}) // xf 0: default
	return b
}

func (b *styleBook) get(k styleKey) int {
	if i, ok := b.index[k]; ok {
		return i
	}
	if k.Fill != "" {
		found := false
		for _, f := range b.fills {
			if f == k.Fill {
				found = true
				break
			}
		}
		if !found {
			b.fills = append(b.fills, k.Fill)
		}
	}
	b.index[k] = len(b.keys)
	b.keys = append(b.keys, k)
	return b.index[k]
}

const (
	fmtGeneral   = 0
	fmtThousands = 3
	fmtDecimal   = 4
	fmtPercent   = 10
	fmtDate      = 164
	fmtMoney     = 165
)

func (b *styleBook) xml() string {
	var s strings.Builder
	s.WriteString(xmlHeader + `<styleSheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">`)
	s.WriteString(`<numFmts count="2"><numFmt numFmtId="164" formatCode="dd.mm.yyyy"/><numFmt numFmtId="165" formatCode="#,##0.00"/></numFmts>`)
	s.WriteString(`<fonts count="2"><font><sz val="11"/><name val="Calibri"/><family val="2"/></font><font><b/><sz val="11"/><name val="Calibri"/><family val="2"/></font></fonts>`)
	fmt.Fprintf(&s, `<fills count="%d"><fill><patternFill patternType="none"/></fill><fill><patternFill patternType="gray125"/></fill>`, 2+len(b.fills))
	for _, f := range b.fills {
		fmt.Fprintf(&s, `<fill><patternFill patternType="solid"><fgColor rgb="%s"/><bgColor indexed="64"/></patternFill></fill>`, f)
	}
	s.WriteString(`</fills>`)
	s.WriteString(`<borders count="2"><border><left/><right/><top/><bottom/><diagonal/></border><border><left/><right/><top/><bottom style="thin"><color rgb="FF9AA9B4"/></bottom><diagonal/></border></borders>`)
	s.WriteString(`<cellStyleXfs count="1"><xf numFmtId="0" fontId="0" fillId="0" borderId="0"/></cellStyleXfs>`)
	fmt.Fprintf(&s, `<cellXfs count="%d">`, len(b.keys))
	for i, k := range b.keys {
		if i == 0 {
			s.WriteString(`<xf numFmtId="0" fontId="0" fillId="0" borderId="0" xfId="0"/>`)
			continue
		}
		font, fill, border := 0, 0, 0
		if k.Bold {
			font = 1
		}
		if k.Fill != "" {
			for j, f := range b.fills {
				if f == k.Fill {
					fill = 2 + j
				}
			}
		}
		if k.Border {
			border = 1
		}
		wrap := ""
		if k.Wrap {
			wrap = ` wrapText="1"`
		}
		fmt.Fprintf(&s, `<xf numFmtId="%d" fontId="%d" fillId="%d" borderId="%d" xfId="0" applyNumberFormat="1" applyFont="1" applyFill="1" applyBorder="1" applyAlignment="1"><alignment vertical="top"%s/></xf>`, k.NumFmt, font, fill, border, wrap)
	}
	s.WriteString(`</cellXfs><cellStyles count="1"><cellStyle name="Normal" xfId="0" builtinId="0"/></cellStyles></styleSheet>`)
	return s.String()
}

// sheetCell is one cell after parsing the model's JSON.
type sheetCell struct {
	kind    byte // 's' text, 'n' number, 'b' boolean, 'f' formula, 0 empty
	text    string
	number  float64
	boolean bool
	fill    string
	bold    bool
	numFmt  int
}

func formatID(name string) (int, bool) {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "money", "деньги", "сумма":
		return fmtMoney, true
	case "percent", "процент":
		return fmtPercent, true
	case "date", "дата":
		return fmtDate, true
	case "integer", "целое":
		return fmtThousands, true
	case "number", "число":
		return fmtDecimal, true
	case "text", "текст":
		return -1, true
	}
	return 0, false
}

func parseCell(raw json.RawMessage, columnFormat string) sheetCell {
	var cell sheetCell
	var value json.RawMessage = raw
	format := columnFormat
	trimmed := strings.TrimSpace(string(raw))
	if strings.HasPrefix(trimmed, "{") {
		var spec cellSpec
		if json.Unmarshal(raw, &spec) == nil {
			value = spec.Value
			if c, ok := normalizeColor(spec.Fill); ok {
				cell.fill = c
			}
			cell.bold = spec.Bold
			if spec.Format != "" {
				format = spec.Format
			}
		}
	}
	forced, hasFormat := formatID(format)
	trimmed = strings.TrimSpace(string(value))
	switch {
	case trimmed == "" || trimmed == "null":
		return cell
	case trimmed == "true" || trimmed == "false":
		cell.kind, cell.boolean = 'b', trimmed == "true"
		return cell
	case strings.HasPrefix(trimmed, `"`):
		var text string
		if json.Unmarshal(value, &text) != nil {
			text = trimmed
		}
		cell.text = text
	default:
		number, err := strconv.ParseFloat(trimmed, 64)
		if err != nil || math.IsInf(number, 0) || math.IsNaN(number) {
			cell.kind, cell.text = 's', trimmed
			return cell
		}
		cell.kind, cell.number = 'n', number
		cell.numFmt = numberFormat(number, forced, hasFormat)
		return cell
	}
	text := cell.text
	if hasFormat && forced == -1 {
		cell.kind = 's'
		return cell
	}
	if strings.HasPrefix(text, "=") && len(text) > 1 {
		cell.kind, cell.text = 'f', strings.TrimPrefix(text, "=")
		if hasFormat && forced > 0 {
			cell.numFmt = forced
		}
		return cell
	}
	if serial, ok := parseDate(text); ok {
		cell.kind, cell.number, cell.numFmt = 'n', serial, fmtDate
		return cell
	}
	if strings.HasSuffix(strings.TrimSpace(text), "%") {
		if n, ok := parseNumber(strings.TrimSuffix(strings.TrimSpace(text), "%")); ok {
			cell.kind, cell.number, cell.numFmt = 'n', n/100, fmtPercent
			return cell
		}
	}
	if n, ok := parseNumber(text); ok {
		cell.kind, cell.number = 'n', n
		cell.numFmt = numberFormat(n, forced, hasFormat)
		return cell
	}
	cell.kind = 's'
	return cell
}

func numberFormat(value float64, forced int, hasFormat bool) int {
	if hasFormat && forced > 0 {
		return forced
	}
	if value != math.Trunc(value) {
		return fmtDecimal
	}
	if math.Abs(value) >= 10000 {
		return fmtThousands
	}
	return fmtGeneral
}

func longestLine(text string) int {
	longest := 0
	for _, line := range strings.Split(text, "\n") {
		if n := len([]rune(line)); n > longest {
			longest = n
		}
	}
	return longest
}

// sanitizeSheetName keeps Excel's rules: no []:*?/\, at most 31 characters, unique.
func sanitizeSheetName(name string, index int, used map[string]bool) string {
	cleaned := strings.Map(func(r rune) rune {
		if strings.ContainsRune(`[]:*?/\`, r) || r < 32 {
			return ' '
		}
		return r
	}, name)
	cleaned = strings.Trim(strings.Join(strings.Fields(cleaned), " "), "'")
	if cleaned == "" {
		cleaned = fmt.Sprintf("Лист%d", index+1)
	}
	runes := []rune(cleaned)
	if len(runes) > 31 {
		cleaned = string(runes[:31])
	}
	base := cleaned
	for n := 2; used[strings.ToLower(cleaned)]; n++ {
		suffix := fmt.Sprintf(" (%d)", n)
		r := []rune(base)
		if len(r)+len(suffix) > 31 {
			r = r[:31-len(suffix)]
		}
		cleaned = string(r) + suffix
	}
	used[strings.ToLower(cleaned)] = true
	return cleaned
}

func sheetXML(sheet sheetSpec, styles *styleBook) (string, string, error) {
	if len(sheet.Rows) > xlsxMaxRows {
		return "", "", fmt.Errorf("на листе больше %d строк", xlsxMaxRows)
	}
	columnCount := len(sheet.Columns)
	for _, row := range sheet.Rows {
		if len(row) > columnCount {
			columnCount = len(row)
		}
	}
	if columnCount == 0 {
		columnCount = 1
	}
	if columnCount > xlsxMaxColumns {
		return "", "", fmt.Errorf("на листе больше %d колонок", xlsxMaxColumns)
	}
	hasHeader := false
	for _, c := range sheet.Columns {
		if strings.TrimSpace(c.Title) != "" {
			hasHeader = true
		}
	}

	cells := make([][]sheetCell, len(sheet.Rows))
	for r, row := range sheet.Rows {
		cells[r] = make([]sheetCell, len(row))
		for c, raw := range row {
			format := ""
			if c < len(sheet.Columns) {
				format = sheet.Columns[c].Format
			}
			cells[r][c] = parseCell(raw, format)
		}
	}

	widths := make([]float64, columnCount)
	wraps := make([]bool, columnCount)
	for c := 0; c < columnCount; c++ {
		longest := 0
		if c < len(sheet.Columns) {
			longest = longestLine(sheet.Columns[c].Title)
		}
		multiline := false
		for _, row := range cells {
			if c >= len(row) {
				continue
			}
			cell := row[c]
			length := 0
			switch cell.kind {
			case 's':
				length = longestLine(cell.text)
				multiline = multiline || strings.Contains(cell.text, "\n")
			case 'n':
				length = len(strconv.FormatFloat(cell.number, 'f', -1, 64)) + 3
				if cell.numFmt == fmtDate {
					length = 10
				}
			case 'b':
				length = 5
			case 'f':
				length = 12
			}
			if length > longest {
				longest = length
			}
		}
		width := math.Min(xlsxMaxWidth, math.Max(xlsxMinWidth, float64(longest+2)))
		if c < len(sheet.Columns) && sheet.Columns[c].Width > 0 {
			width = math.Min(150, math.Max(2, sheet.Columns[c].Width))
		}
		widths[c] = width
		wraps[c] = multiline || longest > xlsxWrapFrom || float64(longest) > width
	}

	headerFill := "FFE7EEF3"
	if c, ok := normalizeColor(sheet.HeaderFill); ok {
		headerFill = c
	}

	var rows strings.Builder
	rowNumber := 0
	if hasHeader {
		rowNumber++
		fmt.Fprintf(&rows, `<row r="%d">`, rowNumber)
		for c := 0; c < columnCount; c++ {
			title := ""
			if c < len(sheet.Columns) {
				title = sheet.Columns[c].Title
			}
			style := styles.get(styleKey{Bold: true, Fill: headerFill, Wrap: wraps[c], Border: true})
			fmt.Fprintf(&rows, `<c r="%s%d" s="%d" t="inlineStr"><is><t xml:space="preserve">%s</t></is></c>`, columnName(c), rowNumber, style, xmlText(title))
		}
		rows.WriteString(`</row>`)
	}
	for _, row := range cells {
		rowNumber++
		fmt.Fprintf(&rows, `<row r="%d">`, rowNumber)
		for c, cell := range row {
			if cell.kind == 0 && cell.fill == "" {
				continue
			}
			ref := fmt.Sprintf("%s%d", columnName(c), rowNumber)
			style := styles.get(styleKey{Bold: cell.bold, Fill: cell.fill, Wrap: wraps[c] && cell.kind == 's', NumFmt: cell.numFmt})
			switch cell.kind {
			case 'n':
				fmt.Fprintf(&rows, `<c r="%s" s="%d"><v>%s</v></c>`, ref, style, strconv.FormatFloat(cell.number, 'f', -1, 64))
			case 'b':
				v := 0
				if cell.boolean {
					v = 1
				}
				fmt.Fprintf(&rows, `<c r="%s" s="%d" t="b"><v>%d</v></c>`, ref, style, v)
			case 'f':
				fmt.Fprintf(&rows, `<c r="%s" s="%d"><f>%s</f></c>`, ref, style, xmlText(cell.text))
			case 's':
				fmt.Fprintf(&rows, `<c r="%s" s="%d" t="inlineStr"><is><t xml:space="preserve">%s</t></is></c>`, ref, style, xmlText(cell.text))
			default:
				fmt.Fprintf(&rows, `<c r="%s" s="%d"/>`, ref, style)
			}
		}
		rows.WriteString(`</row>`)
	}

	last := rowNumber
	if last == 0 {
		last = 1
	}
	lastRef := fmt.Sprintf("%s%d", columnName(columnCount-1), last)
	var s strings.Builder
	s.WriteString(xmlHeader + `<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">`)
	fmt.Fprintf(&s, `<dimension ref="A1:%s"/>`, lastRef)
	freeze := hasHeader && (sheet.FreezeHeader == nil || *sheet.FreezeHeader)
	if freeze {
		s.WriteString(`<sheetViews><sheetView workbookViewId="0"><pane ySplit="1" topLeftCell="A2" activePane="bottomLeft" state="frozen"/><selection pane="bottomLeft" activeCell="A2" sqref="A2"/></sheetView></sheetViews>`)
	} else {
		s.WriteString(`<sheetViews><sheetView workbookViewId="0"/></sheetViews>`)
	}
	s.WriteString(`<sheetFormatPr defaultRowHeight="15"/><cols>`)
	for c, w := range widths {
		fmt.Fprintf(&s, `<col min="%d" max="%d" width="%s" customWidth="1"/>`, c+1, c+1, strconv.FormatFloat(w, 'f', 1, 64))
	}
	s.WriteString(`</cols><sheetData>`)
	s.WriteString(rows.String())
	s.WriteString(`</sheetData>`)
	filterRef := ""
	if hasHeader && len(cells) > 0 && (sheet.Filter == nil || *sheet.Filter) {
		filterRef = "A1:" + lastRef
		fmt.Fprintf(&s, `<autoFilter ref="%s"/>`, filterRef)
	}
	s.WriteString(`<pageMargins left="0.7" right="0.7" top="0.75" bottom="0.75" header="0.3" footer="0.3"/></worksheet>`)
	return s.String(), filterRef, nil
}

func buildXlsx(args xlsxArgs) ([]byte, error) {
	if len(args.Sheets) == 0 {
		return nil, fmt.Errorf("нет ни одного листа (sheets)")
	}
	if len(args.Sheets) > 50 {
		return nil, fmt.Errorf("больше 50 листов")
	}
	styles := newStyleBook()
	used := map[string]bool{}
	var sheetsXML, relsXML, typesXML, definedNames strings.Builder
	files := []zipFile{}
	for i, sheet := range args.Sheets {
		name := sanitizeSheetName(sheet.Name, i, used)
		body, filterRef, err := sheetXML(sheet, styles)
		if err != nil {
			return nil, fmt.Errorf("лист «%s»: %w", name, err)
		}
		files = append(files, zipFile{fmt.Sprintf("xl/worksheets/sheet%d.xml", i+1), body})
		fmt.Fprintf(&sheetsXML, `<sheet name="%s" sheetId="%d" r:id="rId%d"/>`, xmlText(name), i+1, i+1)
		fmt.Fprintf(&relsXML, `<Relationship Id="rId%d" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet%d.xml"/>`, i+1, i+1)
		fmt.Fprintf(&typesXML, `<Override PartName="/xl/worksheets/sheet%d.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/>`, i+1)
		if filterRef != "" {
			parts := strings.SplitN(filterRef, ":", 2)
			fmt.Fprintf(&definedNames, `<definedName name="_xlnm._FilterDatabase" localSheetId="%d" hidden="1">'%s'!%s:%s</definedName>`, i, xmlText(strings.ReplaceAll(name, "'", "''")), absRef(parts[0]), absRef(parts[1]))
		}
	}
	styleRel := len(args.Sheets) + 1
	fmt.Fprintf(&relsXML, `<Relationship Id="rId%d" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles" Target="styles.xml"/>`, styleRel)
	workbook := xmlHeader + `<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"><sheets>` + sheetsXML.String() + `</sheets>`
	if definedNames.Len() > 0 {
		workbook += `<definedNames>` + definedNames.String() + `</definedNames>`
	}
	// Formulas have no cached values: Excel and LibreOffice calculate them on opening.
	workbook += `<calcPr calcId="0" fullCalcOnLoad="1"/></workbook>`
	files = append([]zipFile{
		{"[Content_Types].xml", xmlHeader + `<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"><Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/><Default Extension="xml" ContentType="application/xml"/><Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/><Override PartName="/xl/styles.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.styles+xml"/>` + typesXML.String() + `</Types>`},
		{"_rels/.rels", xmlHeader + `<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="xl/workbook.xml"/></Relationships>`},
		{"xl/workbook.xml", workbook},
		{"xl/_rels/workbook.xml.rels", xmlHeader + `<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">` + relsXML.String() + `</Relationships>`},
	}, files...)
	// Styles are complete only after every sheet was written.
	files = append(files, zipFile{"xl/styles.xml", styles.xml()})
	return writeZip(files)
}

var refParts = regexp.MustCompile(`^([A-Z]+)(\d+)$`)

func absRef(ref string) string {
	g := refParts.FindStringSubmatch(ref)
	if g == nil {
		return ref
	}
	return "$" + g[1] + "$" + g[2]
}
