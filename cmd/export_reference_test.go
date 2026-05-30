package cmd

import "testing"

func TestBuildWikiReferenceURLUsesDomainOverride(t *testing.T) {
	got := buildWikiReferenceURL("Ad8Iw0oz3iSp4kkIi7QctVhin3e", "Ad8Iw0oz3iSp4kkIi7QctVhin3e", "https://oxl611w1w1.feishu.cn/")
	want := "https://oxl611w1w1.feishu.cn/wiki/Ad8Iw0oz3iSp4kkIi7QctVhin3e"
	if got != want {
		t.Fatalf("buildWikiReferenceURL() = %q, want %q", got, want)
	}
}

func TestBuildDocReferenceURLUsesDomainOverride(t *testing.T) {
	got := buildDocReferenceURL("ABC123def456", "ABC123def456", "oxl611w1w1.feishu.cn")
	want := "https://oxl611w1w1.feishu.cn/docx/ABC123def456"
	if got != want {
		t.Fatalf("buildDocReferenceURL() = %q, want %q", got, want)
	}
}

func TestBuildDocReferenceURLPreservesInputURL(t *testing.T) {
	got := buildDocReferenceURL("https://xxx.feishu.cn/docx/ABC123def456?foo=1#bar", "ABC123def456", "https://oxl611w1w1.feishu.cn/")
	want := "https://xxx.feishu.cn/docx/ABC123def456"
	if got != want {
		t.Fatalf("buildDocReferenceURL() = %q, want %q", got, want)
	}
}

func TestNormalizeDomainURLFallsBackToDefault(t *testing.T) {
	got := normalizeDomainURL("://bad-url")
	want := "https://feishu.cn"
	if got != want {
		t.Fatalf("normalizeDomainURL() = %q, want %q", got, want)
	}
}

func TestDocExportRegistersDomainURLFlag(t *testing.T) {
	if exportMarkdownCmd.Flags().Lookup("domainUrl") == nil {
		t.Fatal("domainUrl flag is not registered on doc export")
	}
}
