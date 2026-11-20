package main

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
)

// Document arguments of make_docx.
type docxArgs struct {
	FileName    string      `json:"file_name"`
	Font        string      `json:"font"`
	FontSize    float64     `json:"font_size"`
	MarginsMM   *marginSpec `json:"margins_mm"`
	PageNumbers *bool       `json:"page_numbers"`
	Blocks      []blockSpec `json:"blocks"`
}

type marginSpec struct {
	Top    float64 `json:"top"`
	Bottom float64 `json:"bottom"`
	Left   float64 `json:"left"`
	Right  float64 `json:"right"`
}

type blockSpec struct {
	Type       string              `json:"type"`
	Text       string              `json:"text"`
	Level      int                 `json:"level"`
	Align      string              `json:"align"`
	Ordered    bool                `json:"ordered"`
	Items      []string            `json:"items"`
	Rows       [][]json.RawMessage `json:"rows"`
	Header     *bool               `json:"header"`
	HeaderFill string              `json:"header_fill"`
	Widths     []float64           `json:"widths"`
}

type tableCellSpec struct {
	Text  string `json:"text"`
	Fill  string `json:"fill"`
	Bold  bool   `json:"bold"`
	Align string `json:"align"`
}

const (
	a4Width        = 11906 // twips
	a4Height       = 16838
	twipsPerMM     = 56.6929
	docxMaxBlocks  = 5000
	docxMaxCells   = 50000
	defaultFont    = "Times New Roman"
	defaultFontPt  = 12
	defaultHeadHex = "E7EEF3"
)

// Fonts the model may ask for; anything else falls back to the default.
var allowedFonts = map[string]string{
	"times new roman": "Times New Roman", "arial": "Arial", "calibri": "Calibri", "cambria": "Cambria",
	"georgia": "Georgia", "verdana": "Verdana", "tahoma": "Tahoma", "segoe ui": "Segoe UI",
	"courier new": "Courier New", "pt serif": "PT Serif", "pt sans": "PT Sans", "liberation serif": "Liberation Serif",
}

func resolveFont(name string) string {
	if f, ok := allowedFonts[strings.ToLower(strings.TrimSpace(name))]; ok {
		return f
	}
	return defaultFont
}

func jcValue(align string) string {
	switch strings.ToLower(strings.TrimSpace(align)) {
	case "center", "по центру", "центр":
		return "center"
	case "right", "справа", "вправо":
		return "right"
	case "justify", "both", "по ширине":
		return "both"
	case "left", "слева", "влево":
		return "left"
	}
	return ""
}

func runsXML(text string, extraBold bool) string {
	var b strings.Builder
	for _, r := range parseInline(text) {
		props := ""
		if r.Bold || extraBold {
			props += "<w:b/><w:bCs/>"
		}
		if r.Italic {
			props += "<w:i/><w:iCs/>"
		}
		if r.Underline {
			props += `<w:u w:val="single"/>`
		}
		lines := strings.Split(strings.ReplaceAll(r.Text, "\r\n", "\n"), "\n")
		b.WriteString("<w:r>")
		if props != "" {
			b.WriteString("<w:rPr>" + props + "</w:rPr>")
		}
		for i, line := range lines {
			if i > 0 {
				b.WriteString("<w:br/>")
			}
			fmt.Fprintf(&b, `<w:t xml:space="preserve">%s</w:t>`, xmlText(line))
		}
		b.WriteString("</w:r>")
	}
	return b.String()
}

func paragraphXML(style, align, numXML, text string, bold bool) string {
	props := ""
	if style != "" {
		props += fmt.Sprintf(`<w:pStyle w:val="%s"/>`, style)
	}
	props += numXML
	if jc := jcValue(align); jc != "" {
		props += fmt.Sprintf(`<w:jc w:val="%s"/>`, jc)
	}
	if props != "" {
		props = "<w:pPr>" + props + "</w:pPr>"
	}
	return "<w:p>" + props + runsXML(text, bold) + "</w:p>"
}

func parseTableCell(raw json.RawMessage) tableCellSpec {
	trimmed := strings.TrimSpace(string(raw))
	var cell tableCellSpec
	switch {
	case strings.HasPrefix(trimmed, "{"):
		if json.Unmarshal(raw, &cell) != nil {
			cell.Text = trimmed
		}
	case strings.HasPrefix(trimmed, `"`):
		if json.Unmarshal(raw, &cell.Text) != nil {
			cell.Text = trimmed
		}
	case trimmed == "null" || trimmed == "":
	default:
		cell.Text = trimmed
	}
	return cell
}

func tableXML(block blockSpec, textWidth int) string {
	columns := 0
	for _, row := range block.Rows {
		if len(row) > columns {
			columns = len(row)
		}
	}
	if columns == 0 {
		return ""
	}
	// Column widths: shares given by the model, or equal.
	shares := make([]float64, columns)
	total := 0.0
	for i := range shares {
		shares[i] = 1
		if i < len(block.Widths) && block.Widths[i] > 0 {
			shares[i] = block.Widths[i]
		}
		total += shares[i]
	}
	widths := make([]int, columns)
	for i, s := range shares {
		widths[i] = int(math.Round(float64(textWidth) * s / total))
	}
	header := block.Header == nil || *block.Header
	headerFill := defaultHeadHex
	if c, ok := normalizeColor(block.HeaderFill); ok {
		headerFill = c[2:]
	}

	var b strings.Builder
	b.WriteString(`<w:tbl><w:tblPr><w:tblStyle w:val="TableGrid"/>`)
	fmt.Fprintf(&b, `<w:tblW w:w="%d" w:type="dxa"/><w:tblLayout w:type="fixed"/>`, textWidth)
	b.WriteString(`<w:tblBorders><w:top w:val="single" w:sz="4" w:space="0" w:color="808080"/><w:left w:val="single" w:sz="4" w:space="0" w:color="808080"/><w:bottom w:val="single" w:sz="4" w:space="0" w:color="808080"/><w:right w:val="single" w:sz="4" w:space="0" w:color="808080"/><w:insideH w:val="single" w:sz="4" w:space="0" w:color="808080"/><w:insideV w:val="single" w:sz="4" w:space="0" w:color="808080"/></w:tblBorders>`)
	b.WriteString(`<w:tblCellMar><w:top w:w="40" w:type="dxa"/><w:left w:w="80" w:type="dxa"/><w:bottom w:w="40" w:type="dxa"/><w:right w:w="80" w:type="dxa"/></w:tblCellMar></w:tblPr><w:tblGrid>`)
	for _, w := range widths {
		fmt.Fprintf(&b, `<w:gridCol w:w="%d"/>`, w)
	}
	b.WriteString(`</w:tblGrid>`)
	for r, row := range block.Rows {
		isHeader := header && r == 0
		b.WriteString("<w:tr>")
		if isHeader {
			b.WriteString(`<w:trPr><w:tblHeader/><w:cantSplit/></w:trPr>`)
		}
		for c := 0; c < columns; c++ {
			var cell tableCellSpec
			if c < len(row) {
				cell = parseTableCell(row[c])
			}
			fill := ""
			if color, ok := normalizeColor(cell.Fill); ok {
				fill = color[2:]
			} else if isHeader {
				fill = headerFill
			}
			fmt.Fprintf(&b, `<w:tc><w:tcPr><w:tcW w:w="%d" w:type="dxa"/>`, widths[c])
			if fill != "" {
				fmt.Fprintf(&b, `<w:shd w:val="clear" w:color="auto" w:fill="%s"/>`, fill)
			}
			b.WriteString(`</w:tcPr>`)
			b.WriteString(paragraphXML("TableText", cell.Align, "", cell.Text, cell.Bold || isHeader))
			b.WriteString("</w:tc>")
		}
		b.WriteString("</w:tr>")
	}
	b.WriteString("</w:tbl>")
	// Word needs a paragraph after a table; it also separates two tables in a row.
	b.WriteString(`<w:p><w:pPr><w:pStyle w:val="TableText"/></w:pPr></w:p>`)
	return b.String()
}

func buildDocx(args docxArgs) ([]byte, error) {
	if len(args.Blocks) == 0 {
		return nil, fmt.Errorf("документ пустой: нет блоков (blocks)")
	}
	if len(args.Blocks) > docxMaxBlocks {
		return nil, fmt.Errorf("больше %d блоков", docxMaxBlocks)
	}
	font := resolveFont(args.Font)
	size := args.FontSize
	if size < 6 || size > 72 {
		size = defaultFontPt
	}
	margins := marginSpec{Top: 20, Bottom: 20, Left: 30, Right: 15}
	if m := args.MarginsMM; m != nil {
		pick := func(v, def float64) float64 {
			if v >= 5 && v <= 80 {
				return v
			}
			return def
		}
		margins = marginSpec{pick(m.Top, 20), pick(m.Bottom, 20), pick(m.Left, 30), pick(m.Right, 15)}
	}
	tw := func(mm float64) int { return int(math.Round(mm * twipsPerMM)) }
	textWidth := a4Width - tw(margins.Left) - tw(margins.Right)
	pageNumbers := args.PageNumbers == nil || *args.PageNumbers

	var body strings.Builder
	cells := 0
	orderedLists := 0
	for _, block := range args.Blocks {
		switch strings.ToLower(strings.TrimSpace(block.Type)) {
		case "heading", "заголовок":
			level := block.Level
			if level < 1 {
				level = 1
			}
			if level > 3 {
				level = 3
			}
			body.WriteString(paragraphXML(fmt.Sprintf("Heading%d", level), block.Align, "", block.Text, false))
		case "list", "список":
			num := 1
			if block.Ordered {
				orderedLists++
				num = 1 + orderedLists
			}
			for _, item := range block.Items {
				body.WriteString(paragraphXML("ListParagraph", "", fmt.Sprintf(`<w:numPr><w:ilvl w:val="0"/><w:numId w:val="%d"/></w:numPr>`, num), item, false))
			}
		case "table", "таблица":
			for _, row := range block.Rows {
				cells += len(row)
			}
			if cells > docxMaxCells {
				return nil, fmt.Errorf("в таблицах больше %d ячеек", docxMaxCells)
			}
			body.WriteString(tableXML(block, textWidth))
		case "page_break", "разрыв страницы":
			body.WriteString(`<w:p><w:r><w:br w:type="page"/></w:r></w:p>`)
		default: // paragraph and anything unknown: the text is kept as a paragraph
			body.WriteString(paragraphXML("", block.Align, "", block.Text, false))
		}
	}

	footerRef := ""
	if pageNumbers {
		footerRef = `<w:footerReference w:type="default" r:id="rIdFooter1"/>`
	}
	fmt.Fprintf(&body, `<w:sectPr>%s<w:pgSz w:w="%d" w:h="%d"/><w:pgMar w:top="%d" w:right="%d" w:bottom="%d" w:left="%d" w:header="709" w:footer="709" w:gutter="0"/></w:sectPr>`,
		footerRef, a4Width, a4Height, tw(margins.Top), tw(margins.Right), tw(margins.Bottom), tw(margins.Left))

	ns := `xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"`
	document := xmlHeader + `<w:document ` + ns + `><w:body>` + body.String() + `</w:body></w:document>`

	halfPoints := int(math.Round(size * 2))
	fontXML := fmt.Sprintf(`<w:rFonts w:ascii="%[1]s" w:hAnsi="%[1]s" w:cs="%[1]s" w:eastAsia="%[1]s"/>`, xmlText(font))
	heading := func(level, sz int) string {
		return fmt.Sprintf(`<w:style w:type="paragraph" w:styleId="Heading%[1]d"><w:name w:val="heading %[1]d"/><w:basedOn w:val="Normal"/><w:next w:val="Normal"/><w:qFormat/><w:pPr><w:keepNext/><w:keepLines/><w:spacing w:before="240" w:after="120"/><w:outlineLvl w:val="%[3]d"/></w:pPr><w:rPr><w:b/><w:bCs/><w:sz w:val="%[2]d"/><w:szCs w:val="%[2]d"/></w:rPr></w:style>`, level, sz, level-1)
	}
	styles := xmlHeader + `<w:styles ` + ns + `>` +
		`<w:docDefaults><w:rPrDefault><w:rPr>` + fontXML + fmt.Sprintf(`<w:sz w:val="%[1]d"/><w:szCs w:val="%[1]d"/>`, halfPoints) + `<w:lang w:val="ru-RU" w:eastAsia="ru-RU" w:bidi="ar-SA"/></w:rPr></w:rPrDefault>` +
		`<w:pPrDefault><w:pPr><w:spacing w:after="120" w:line="264" w:lineRule="auto"/></w:pPr></w:pPrDefault></w:docDefaults>` +
		`<w:style w:type="paragraph" w:default="1" w:styleId="Normal"><w:name w:val="Normal"/><w:qFormat/></w:style>` +
		heading(1, halfPoints+8) + heading(2, halfPoints+4) + heading(3, halfPoints) +
		`<w:style w:type="paragraph" w:styleId="ListParagraph"><w:name w:val="List Paragraph"/><w:basedOn w:val="Normal"/><w:qFormat/><w:pPr><w:spacing w:after="60"/><w:ind w:left="720"/></w:pPr></w:style>` +
		`<w:style w:type="paragraph" w:styleId="TableText"><w:name w:val="Table Text"/><w:basedOn w:val="Normal"/><w:pPr><w:spacing w:after="0" w:line="240" w:lineRule="auto"/></w:pPr></w:style>` +
		`<w:style w:type="table" w:default="1" w:styleId="TableNormal"><w:name w:val="Normal Table"/><w:tblPr><w:tblInd w:w="0" w:type="dxa"/><w:tblCellMar><w:top w:w="0" w:type="dxa"/><w:left w:w="108" w:type="dxa"/><w:bottom w:w="0" w:type="dxa"/><w:right w:w="108" w:type="dxa"/></w:tblCellMar></w:tblPr></w:style>` +
		`<w:style w:type="table" w:styleId="TableGrid"><w:name w:val="Table Grid"/><w:basedOn w:val="TableNormal"/></w:style>` +
		`</w:styles>`

	// numId 1: bullets; numId 2..: one per numbered list, so each starts from 1.
	var numbering strings.Builder
	numbering.WriteString(xmlHeader + `<w:numbering ` + ns + `>`)
	numbering.WriteString(`<w:abstractNum w:abstractNumId="0"><w:multiLevelType w:val="singleLevel"/><w:lvl w:ilvl="0"><w:start w:val="1"/><w:numFmt w:val="bullet"/><w:lvlText w:val="•"/><w:lvlJc w:val="left"/><w:pPr><w:ind w:left="720" w:hanging="360"/></w:pPr></w:lvl></w:abstractNum>`)
	numbering.WriteString(`<w:abstractNum w:abstractNumId="1"><w:multiLevelType w:val="singleLevel"/><w:lvl w:ilvl="0"><w:start w:val="1"/><w:numFmt w:val="decimal"/><w:lvlText w:val="%1."/><w:lvlJc w:val="left"/><w:pPr><w:ind w:left="720" w:hanging="360"/></w:pPr></w:lvl></w:abstractNum>`)
	numbering.WriteString(`<w:num w:numId="1"><w:abstractNumId w:val="0"/></w:num>`)
	for i := 0; i < orderedLists; i++ {
		fmt.Fprintf(&numbering, `<w:num w:numId="%d"><w:abstractNumId w:val="1"/><w:lvlOverride w:ilvl="0"><w:startOverride w:val="1"/></w:lvlOverride></w:num>`, i+2)
	}
	numbering.WriteString(`</w:numbering>`)

	footer := xmlHeader + `<w:ftr ` + ns + `><w:p><w:pPr><w:pStyle w:val="TableText"/><w:jc w:val="center"/></w:pPr><w:r><w:fldChar w:fldCharType="begin"/></w:r><w:r><w:instrText xml:space="preserve"> PAGE </w:instrText></w:r><w:r><w:fldChar w:fldCharType="separate"/></w:r><w:r><w:t>1</w:t></w:r><w:r><w:fldChar w:fldCharType="end"/></w:r></w:p></w:ftr>`

	rels := xmlHeader + `<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">` +
		`<Relationship Id="rIdStyles" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles" Target="styles.xml"/>` +
		`<Relationship Id="rIdNumbering" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/numbering" Target="numbering.xml"/>`
	types := xmlHeader + `<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"><Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/><Default Extension="xml" ContentType="application/xml"/>` +
		`<Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>` +
		`<Override PartName="/word/styles.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.styles+xml"/>` +
		`<Override PartName="/word/numbering.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.numbering+xml"/>`
	files := []zipFile{}
	if pageNumbers {
		rels += `<Relationship Id="rIdFooter1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/footer" Target="footer1.xml"/>`
		types += `<Override PartName="/word/footer1.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.footer+xml"/>`
		files = append(files, zipFile{"word/footer1.xml", footer})
	}
	rels += `</Relationships>`
	types += `</Types>`
	files = append([]zipFile{
		{"[Content_Types].xml", types},
		{"_rels/.rels", xmlHeader + `<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/></Relationships>`},
		{"word/document.xml", document},
		{"word/_rels/document.xml.rels", rels},
		{"word/styles.xml", styles},
		{"word/numbering.xml", numbering.String()},
	}, files...)
	return writeZip(files)
}
