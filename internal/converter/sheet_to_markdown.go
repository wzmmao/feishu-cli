package converter

import (
	"fmt"
	"strings"
)

// SheetData 单个工作表的数据
type SheetData struct {
	Title  string  // 工作表标题
	Values [][]any // 二维数组，V2 API 返回
}

// SheetToMarkdown 将电子表格数据转换为 Markdown
func SheetToMarkdown(sheets []*SheetData) string {
	var sb strings.Builder

	for i, sheet := range sheets {
		if i > 0 {
			sb.WriteString("\n---\n\n")
		}

		if sheet.Title != "" {
			sb.WriteString("## ")
			sb.WriteString(sheet.Title)
			sb.WriteString("\n\n")
		}

		rows := trimEmptyRows(sheet.Values)
		if len(rows) == 0 {
			sb.WriteString("（空工作表）\n")
			continue
		}

		// 确定最大列数
		maxCols := 0
		for _, row := range rows {
			if len(row) > maxCols {
				maxCols = len(row)
			}
		}

		// 裁剪尾部全空列
		maxCols = trimEmptyCols(rows, maxCols)
		if maxCols == 0 {
			sb.WriteString("（空工作表）\n")
			continue
		}

		// 第一行作为表头
		writeRow(&sb, rows[0], maxCols)
		// 分隔行
		sb.WriteString("|")
		for c := 0; c < maxCols; c++ {
			sb.WriteString(" --- |")
		}
		sb.WriteString("\n")

		// 数据行
		for _, row := range rows[1:] {
			writeRow(&sb, row, maxCols)
		}

		sb.WriteString("\n")
	}

	return strings.TrimRight(sb.String(), "\n") + "\n"
}

// writeRow 写入一行 Markdown 表格
func writeRow(sb *strings.Builder, row []any, maxCols int) {
	sb.WriteString("|")
	for c := 0; c < maxCols; c++ {
		sb.WriteString(" ")
		if c < len(row) {
			sb.WriteString(cellToMarkdown(row[c]))
		}
		sb.WriteString(" |")
	}
	sb.WriteString("\n")
}

// cellToMarkdown 将单元格值转为 Markdown 文本
func cellToMarkdown(cell any) string {
	if cell == nil {
		return ""
	}

	switch v := cell.(type) {
	case string:
		return escapeMDTableCell(v)
	case float64:
		// 整数不带小数点
		if v == float64(int64(v)) {
			return fmt.Sprintf("%d", int64(v))
		}
		return fmt.Sprintf("%g", v)
	case bool:
		if v {
			return "TRUE"
		}
		return "FALSE"
	case []any:
		// 富文本数组（mention / attachment / text 混合）
		return escapeMDTableCell(richTextArrayToMarkdown(v))
	case map[string]any:
		// 单个富文本元素
		return escapeMDTableCell(richTextElementToMarkdown(v))
	default:
		return escapeMDTableCell(fmt.Sprintf("%v", v))
	}
}

// richTextArrayToMarkdown 将富文本数组转为 Markdown
func richTextArrayToMarkdown(elements []any) string {
	var parts []string
	for _, elem := range elements {
		m, ok := elem.(map[string]any)
		if !ok {
			parts = append(parts, fmt.Sprintf("%v", elem))
			continue
		}
		text := richTextElementToMarkdown(m)
		if text != "" {
			parts = append(parts, text)
		}
	}
	return strings.Join(parts, "")
}

// richTextElementToMarkdown 将单个富文本元素转为 Markdown
func richTextElementToMarkdown(m map[string]any) string {
	elemType, _ := m["type"].(string)

	switch elemType {
	case "text":
		text, _ := m["text"].(string)
		return text
	case "mention":
		// @文档/幻灯片等引用
		text, _ := m["text"].(string)
		link, _ := m["link"].(string)
		if link != "" && text != "" {
			return fmt.Sprintf("[%s](%s)", text, link)
		}
		return text
	case "attachment":
		text, _ := m["text"].(string)
		if text == "" {
			text = "附件"
		}
		return fmt.Sprintf("📎 %s", text)
	default:
		// 尝试通用处理：有 link+text 就做链接，否则取 text
		text, _ := m["text"].(string)
		link, _ := m["link"].(string)
		if link != "" && text != "" {
			return fmt.Sprintf("[%s](%s)", text, link)
		}
		if text != "" {
			return text
		}
		return ""
	}
}

// escapeMDTableCell 转义 Markdown 表格中的特殊字符
func escapeMDTableCell(s string) string {
	s = strings.ReplaceAll(s, "|", "\\|")
	s = strings.ReplaceAll(s, "\n", "<br>")
	return strings.TrimSpace(s)
}

// trimEmptyRows 去除尾部的全空行
func trimEmptyRows(rows [][]any) [][]any {
	last := len(rows)
	for last > 0 {
		if !isRowEmpty(rows[last-1]) {
			break
		}
		last--
	}
	return rows[:last]
}

// trimEmptyCols 返回裁剪尾部全空列后的列数
func trimEmptyCols(rows [][]any, maxCols int) int {
	for maxCols > 0 {
		allEmpty := true
		for _, row := range rows {
			if maxCols-1 < len(row) && row[maxCols-1] != nil {
				if s, ok := row[maxCols-1].(string); ok && s == "" {
					continue
				}
				allEmpty = false
				break
			}
		}
		if !allEmpty {
			break
		}
		maxCols--
	}
	return maxCols
}

// isRowEmpty 判断行是否全空
func isRowEmpty(row []any) bool {
	for _, cell := range row {
		if cell == nil {
			continue
		}
		if s, ok := cell.(string); ok && s == "" {
			continue
		}
		return false
	}
	return true
}
