package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/riba2534/feishu-cli/internal/auth"
	"github.com/riba2534/feishu-cli/internal/config"
	"github.com/spf13/cobra"
)

var authDeviceCmd = &cobra.Command{
	Use:   "device",
	Short: "设备流授权（无需浏览器回调，适合无头环境）",
	Long: `通过 OAuth 2.0 Device Authorization Grant（RFC 8628）完成用户授权。

适用场景:
  · 无法配置重定向 URL 的应用
  · CI/CD 流水线、无头服务器、容器化环境
  · 不方便在本机运行回调服务器的情况

与 auth login 的区别:
  · auth login  需要配置重定向 URL（http://127.0.0.1:9768/callback）
  · auth device 无需重定向 URL，显示用户码后轮询等待用户在浏览器完成授权

流程:
  1. 命令自动向飞书请求设备码和用户码
  2. 终端显示用户码和验证链接
  3. 用户在任意设备的浏览器打开链接并输入用户码完成授权
  4. 命令自动轮询检测授权结果，成功后保存 token

Token 保存位置: ~/.feishu-cli/token.json

示例:
  # 默认授权（自动包含 offline_access）
  feishu-cli auth device

  # 指定 scope
  feishu-cli auth device --scopes "search:docs:read search:message offline_access"

  # 仅获取设备码信息（不轮询，供脚本使用）
  feishu-cli auth device --print-code`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := config.Validate(); err != nil {
			return err
		}

		cfg := config.Get()
		scopes, _ := cmd.Flags().GetString("scopes")
		printCode, _ := cmd.Flags().GetBool("print-code")

		// 步骤一：请求设备码
		deviceResp, err := auth.RequestDeviceAuthorization(cfg.AppID, cfg.AppSecret, cfg.BaseURL, scopes)
		if err != nil {
			return err
		}

		// --print-code 模式：仅输出 JSON，不轮询
		if printCode {
			return printJSON(deviceResp)
		}

		// 展示用户码和验证链接
		fmt.Fprintln(os.Stderr, "\n请在浏览器中完成以下操作:")
		fmt.Fprintln(os.Stderr, "─────────────────────────────────────────────")
		fmt.Fprintf(os.Stderr, "  1. 打开链接: %s\n", deviceResp.VerificationURI)
		fmt.Fprintf(os.Stderr, "  2. 输入用户码: %s\n", formatUserCode(deviceResp.UserCode))
		fmt.Fprintln(os.Stderr, "─────────────────────────────────────────────")
		if deviceResp.VerificationURIComplete != "" && deviceResp.VerificationURIComplete != deviceResp.VerificationURI {
			fmt.Fprintf(os.Stderr, "\n或直接访问完整链接（含用户码）:\n  %s\n", deviceResp.VerificationURIComplete)
		}
		fmt.Fprintf(os.Stderr, "\n等待授权（%d 秒后过期）...\n", deviceResp.ExpiresIn)

		// 尝试打开浏览器（本地桌面环境）
		openURL := deviceResp.VerificationURIComplete
		if openURL == "" {
			openURL = deviceResp.VerificationURI
		}
		_ = auth.TryOpenBrowser(openURL)

		// 步骤二：轮询 token
		lastLine := ""
		token, err := auth.PollDeviceToken(
			cfg.AppID, cfg.AppSecret, cfg.BaseURL,
			deviceResp.DeviceCode, deviceResp.Interval, deviceResp.ExpiresIn,
			func(elapsed, total int) {
				line := fmt.Sprintf("\r  轮询中... 已等待 %ds / %ds", elapsed, total)
				// 清除上次输出，避免行末残留
				if len(line) < len(lastLine) {
					line += strings.Repeat(" ", len(lastLine)-len(line))
				}
				lastLine = line
				fmt.Fprint(os.Stderr, line)
			},
		)

		// 换行，结束进度显示
		if lastLine != "" {
			fmt.Fprintln(os.Stderr)
		}

		if err != nil {
			return err
		}

		// 保存 token
		if err := auth.SaveToken(token); err != nil {
			return err
		}

		path, _ := auth.TokenPath()
		fmt.Fprintln(os.Stderr, "\n✓ 授权成功！")
		fmt.Fprintf(os.Stderr, "  Token 已保存到 %s\n", path)
		fmt.Fprintf(os.Stderr, "  Access Token 有效期至: %s\n", token.ExpiresAt.Format("2006-01-02 15:04:05"))
		if !token.RefreshExpiresAt.IsZero() {
			fmt.Fprintf(os.Stderr, "  Refresh Token 有效期至: %s\n", token.RefreshExpiresAt.Format("2006-01-02 15:04:05"))
		}
		if token.Scope != "" {
			fmt.Fprintf(os.Stderr, "  授权范围: %s\n", token.Scope)
		}

		return nil
	},
}

// formatUserCode 将用户码格式化为易读形式（如 ABCD-EFGH）
func formatUserCode(code string) string {
	// 如果已包含分隔符则直接返回
	if strings.ContainsAny(code, "-_ ") {
		return code
	}
	// 8 位以上的码按 4+4 格式展示
	if len(code) == 8 {
		return code[:4] + "-" + code[4:]
	}
	return code
}

func init() {
	authCmd.AddCommand(authDeviceCmd)

	authDeviceCmd.Flags().String("scopes", "", "请求的 OAuth scope（空格分隔，offline_access 自动包含）")
	authDeviceCmd.Flags().Bool("print-code", false, "仅输出设备码信息 JSON，不轮询（供脚本使用）")
}
