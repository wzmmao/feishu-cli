package cmd

import (
	"strings"
	"testing"

	larkdocx "github.com/larksuite/oapi-sdk-go/v3/service/docx/v1"
	"github.com/riba2534/feishu-cli/internal/converter"
	"github.com/spf13/cobra"
)

type fakeExportUserResolver struct {
	users map[string]converter.MentionUserInfo
	calls int
	ids   []string
}

func (r *fakeExportUserResolver) BatchResolve(userIDs []string) map[string]converter.MentionUserInfo {
	r.calls++
	r.ids = append([]string(nil), userIDs...)

	result := make(map[string]converter.MentionUserInfo)
	for _, id := range userIDs {
		if info, ok := r.users[id]; ok {
			result[id] = info
		}
	}
	return result
}

func TestReadExpandMentionsFlagDefaultsToTrue(t *testing.T) {
	if !readExpandMentionsFlag(nil) {
		t.Fatal("nil command should default expand-mentions to true")
	}

	cmdWithoutFlag := &cobra.Command{}
	if !readExpandMentionsFlag(cmdWithoutFlag) {
		t.Fatal("missing expand-mentions flag should default to true")
	}

	cmdWithFlag := &cobra.Command{}
	cmdWithFlag.Flags().Bool("expand-mentions", true, "")
	if !readExpandMentionsFlag(cmdWithFlag) {
		t.Fatal("default expand-mentions flag value should be true")
	}

	if err := cmdWithFlag.Flags().Set("expand-mentions", "false"); err != nil {
		t.Fatalf("set flag: %v", err)
	}
	if readExpandMentionsFlag(cmdWithFlag) {
		t.Fatal("explicit expand-mentions=false should be respected")
	}
}

func TestWikiExportCommandsRegisterExpandMentions(t *testing.T) {
	for _, tc := range []struct {
		name string
		cmd  *cobra.Command
	}{
		{name: "wiki export", cmd: exportWikiCmd},
		{name: "wiki export-tree", cmd: exportWikiTreeCmd},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.cmd.Flags().Lookup("expand-mentions") == nil {
				t.Fatal("expand-mentions flag is not registered")
			}
			got, err := tc.cmd.Flags().GetBool("expand-mentions")
			if err != nil {
				t.Fatalf("get expand-mentions flag: %v", err)
			}
			if !got {
				t.Fatal("expand-mentions should default to true")
			}
		})
	}
}

func TestWikiExportCommandsRegisterDomainURL(t *testing.T) {
	for _, tc := range []struct {
		name string
		cmd  *cobra.Command
	}{
		{name: "wiki export", cmd: exportWikiCmd},
		{name: "wiki export-tree", cmd: exportWikiTreeCmd},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.cmd.Flags().Lookup("domainUrl") == nil {
				t.Fatal("domainUrl flag is not registered")
			}
		})
	}
}

func TestNewExportBlockToMarkdownConverterExpandMentions(t *testing.T) {
	resolver := &fakeExportUserResolver{
		users: map[string]converter.MentionUserInfo{
			"ou_123": {Name: "张三", Email: "user@example.com"},
		},
	}

	conv := newExportBlockToMarkdownConverter(
		createExportMentionUserBlocks("ou_123"),
		converter.ConvertOptions{ExpandMentions: true},
		resolver,
	)
	got, err := conv.Convert()
	if err != nil {
		t.Fatalf("convert: %v", err)
	}

	if !strings.Contains(got, "[@张三](mailto:user@example.com)") {
		t.Fatalf("mention should be expanded, got: %s", got)
	}
	if resolver.calls != 1 {
		t.Fatalf("resolver calls = %d, want 1", resolver.calls)
	}
	if len(resolver.ids) != 1 || resolver.ids[0] != "ou_123" {
		t.Fatalf("resolver ids = %v, want [ou_123]", resolver.ids)
	}
}

func TestNewExportBlockToMarkdownConverterPreservesMentionTagsWhenDisabled(t *testing.T) {
	resolver := &fakeExportUserResolver{
		users: map[string]converter.MentionUserInfo{
			"ou_123": {Name: "张三", Email: "user@example.com"},
		},
	}

	conv := newExportBlockToMarkdownConverter(
		createExportMentionUserBlocks("ou_123"),
		converter.ConvertOptions{ExpandMentions: false},
		resolver,
	)
	got, err := conv.Convert()
	if err != nil {
		t.Fatalf("convert: %v", err)
	}

	if !strings.Contains(got, `<mention-user id="ou_123"/>`) {
		t.Fatalf("mention tag should be preserved for roundtrip, got: %s", got)
	}
	if strings.Contains(got, "mailto:") || strings.Contains(got, "@张三") {
		t.Fatalf("mention should not be expanded when disabled, got: %s", got)
	}
	if resolver.calls != 0 {
		t.Fatalf("resolver calls = %d, want 0", resolver.calls)
	}
}

func TestBuildWikiReferenceURLFromInputURL(t *testing.T) {
	got := buildWikiReferenceURL("https://xxx.feishu.cn/wiki/Ad8Iw0oz3iSp4kkIi7QctVhin3e?foo=1#bar", "Ad8Iw0oz3iSp4kkIi7QctVhin3e", "")
	want := "https://xxx.feishu.cn/wiki/Ad8Iw0oz3iSp4kkIi7QctVhin3e"
	if got != want {
		t.Fatalf("buildWikiReferenceURL() = %q, want %q", got, want)
	}
}

func TestBuildWikiReferenceURLFromToken(t *testing.T) {
	got := buildWikiReferenceURL("Ad8Iw0oz3iSp4kkIi7QctVhin3e", "Ad8Iw0oz3iSp4kkIi7QctVhin3e", "")
	want := "https://feishu.cn/wiki/Ad8Iw0oz3iSp4kkIi7QctVhin3e"
	if got != want {
		t.Fatalf("buildWikiReferenceURL() = %q, want %q", got, want)
	}
}

func TestPrependReferenceQuote(t *testing.T) {
	markdown := "# 标题\n\n正文"
	got := prependReferenceQuote(markdown, "https://xxx.feishu.cn/wiki/Ad8Iw0oz3iSp4kkIi7QctVhin3e")
	want := "> https://xxx.feishu.cn/wiki/Ad8Iw0oz3iSp4kkIi7QctVhin3e\n\n# 标题\n\n正文"
	if got != want {
		t.Fatalf("prependReferenceQuote() = %q, want %q", got, want)
	}
}

func TestBuildExportConvertOptionsIncludesOutputMarkdownPath(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().Bool("download-images", true, "")
	cmd.Flags().String("assets-dir", "./assets", "")
	cmd.Flags().Bool("expand-sheets", true, "")
	cmd.Flags().Bool("expand-mentions", true, "")
	cmd.Flags().String("domainUrl", "https://example.feishu.cn", "")

	if err := cmd.Flags().Set("download-images", "true"); err != nil {
		t.Fatalf("set download-images: %v", err)
	}

	options := buildExportConvertOptions(
		cmd,
		"docxToken",
		"userToken",
		"./tree-assets/Team/Plan",
		"./backup/Team/Plan.md",
	)

	if !options.DownloadImages {
		t.Fatal("DownloadImages should be true")
	}
	if options.AssetsDir != "./tree-assets/Team/Plan" {
		t.Fatalf("AssetsDir = %q, want %q", options.AssetsDir, "./tree-assets/Team/Plan")
	}
	if options.OutputMarkdownPath != "./backup/Team/Plan.md" {
		t.Fatalf("OutputMarkdownPath = %q, want %q", options.OutputMarkdownPath, "./backup/Team/Plan.md")
	}
	if options.DocumentID != "docxToken" {
		t.Fatalf("DocumentID = %q, want %q", options.DocumentID, "docxToken")
	}
	if options.UserAccessToken != "userToken" {
		t.Fatalf("UserAccessToken = %q, want %q", options.UserAccessToken, "userToken")
	}
}

func createExportMentionUserBlocks(userID string) []*larkdocx.Block {
	blockType := int(converter.BlockTypeText)
	return []*larkdocx.Block{
		{
			BlockId:   exportStringPtr("block1"),
			BlockType: &blockType,
			Text: &larkdocx.Text{
				Elements: []*larkdocx.TextElement{
					{TextRun: &larkdocx.TextRun{Content: exportStringPtr("你好 ")}},
					{MentionUser: &larkdocx.MentionUser{UserId: exportStringPtr(userID)}},
					{TextRun: &larkdocx.TextRun{Content: exportStringPtr(" 请查看")}},
				},
			},
		},
	}
}

func exportStringPtr(s string) *string {
	return &s
}
