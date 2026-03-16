package cmd

import (
	"github.com/spf13/cobra"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "用户授权管理",
	Long: `管理 OAuth 2.0 用户授权，用于获取 User Access Token。

子命令:
  login    登录授权（Authorization Code Flow，需配置重定向 URL）
  device   设备流授权（Device Flow，无需重定向 URL，适合无头环境）
  status   查看当前授权状态
  logout   退出登录（清除本地 token）

搜索功能（search messages/docs/apps）需要 User Access Token。
通过 auth login 或 auth device 可以完成 OAuth 授权。

auth login 前置条件:
  在飞书开放平台 → 应用详情 → 安全设置 → 重定向 URL 中添加:
  http://127.0.0.1:9768/callback

示例:
  # 标准登录（需重定向 URL）
  feishu-cli auth login

  # 设备流登录（无需重定向 URL，适合 CI/CD 和无头环境）
  feishu-cli auth device

  # 查看授权状态
  feishu-cli auth status

  # 退出登录
  feishu-cli auth logout`,
}

func init() {
	rootCmd.AddCommand(authCmd)
}
