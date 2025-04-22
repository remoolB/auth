package email

import (
	"auth/biz/infrastructure/config"
	"crypto/tls"
	"fmt"
	"math/rand"
	"net/smtp"
	"strconv"
	"strings"
	"time"
)

// SendEmail 发送邮件
func SendEmail(to, subject, body string) error {
	// 获取邮箱配置
	emailConfig := config.GetConfig().Email
	fmt.Println("正在发送邮件，配置:", emailConfig.Host, emailConfig.Port, emailConfig.Username)

	// 设置发件人
	from := emailConfig.Username

	// 设置消息内容
	message := []byte(fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"Content-Type: text/html; charset=UTF-8\r\n"+
		"\r\n"+
		"%s", from, to, subject, body))

	// 根据端口决定使用TLS还是普通SMTP
	if emailConfig.Port == 465 {
		return SendEmailWithTLS(emailConfig, from, to, message)
	}

	// 普通SMTP (端口25或587)
	smtpServer := fmt.Sprintf("%s:%d", emailConfig.Host, emailConfig.Port)
	fmt.Println("使用普通SMTP服务器地址:", smtpServer)

	// 身份验证信息
	authHost := strings.Split(emailConfig.Host, ":")[0]
	fmt.Println("认证服务器:", authHost)
	auth := smtp.PlainAuth("", emailConfig.Username, emailConfig.Password, authHost)

	// 发送邮件
	err := smtp.SendMail(smtpServer, auth, from, []string{to}, message)
	if err != nil {
		fmt.Println("SMTP发送邮件失败:", err)
	}
	return err
}

// SendEmailWithTLS 通过TLS发送邮件
func SendEmailWithTLS(emailConfig config.EmailConfig, from, to string, message []byte) error {
	host := strings.Split(emailConfig.Host, ":")[0]
	fmt.Println("使用TLS发送邮件，服务器:", host)

	// TLS配置
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         host,
	}

	// 连接到SMTP服务器
	conn, err := tls.Dial("tcp", fmt.Sprintf("%s:%d", host, emailConfig.Port), tlsConfig)
	if err != nil {
		fmt.Println("TLS连接失败:", err)
		return err
	}

	// 创建SMTP客户端
	smtpClient, err := smtp.NewClient(conn, host)
	if err != nil {
		fmt.Println("创建SMTP客户端失败:", err)
		return err
	}
	defer smtpClient.Quit()

	// 身份验证
	auth := smtp.PlainAuth("", emailConfig.Username, emailConfig.Password, host)
	if err = smtpClient.Auth(auth); err != nil {
		fmt.Println("SMTP认证失败:", err)
		return err
	}

	// 设置发件人和收件人
	if err = smtpClient.Mail(from); err != nil {
		fmt.Println("设置发件人失败:", err)
		return err
	}

	if err = smtpClient.Rcpt(to); err != nil {
		fmt.Println("设置收件人失败:", err)
		return err
	}

	// 发送邮件正文
	writer, err := smtpClient.Data()
	if err != nil {
		fmt.Println("获取数据写入器失败:", err)
		return err
	}

	_, err = writer.Write(message)
	if err != nil {
		fmt.Println("写入邮件内容失败:", err)
		return err
	}

	err = writer.Close()
	if err != nil {
		fmt.Println("关闭写入器失败:", err)
		return err
	}

	fmt.Println("TLS邮件发送成功")
	return nil
}

// SendVerificationCode 发送验证码邮件
func SendVerificationCode(to, code string) error {
	subject := "验证码 - 登录验证"
	fmt.Println("准备发送验证码邮件至:", to, "验证码:", code)

	// 构建HTML邮件内容
	htmlBody := fmt.Sprintf(`
		<div style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto; padding: 20px; border: 1px solid #e0e0e0; border-radius: 5px;">
			<h2 style="color: #333;">您的验证码</h2>
			<p style="font-size: 16px; color: #666;">您好，</p>
			<p style="font-size: 16px; color: #666;">您的验证码是：</p>
			<div style="background-color: #f5f5f5; padding: 15px; text-align: center; font-size: 24px; font-weight: bold; letter-spacing: 5px; margin: 20px 0;">
				%s
			</div>
			<p style="font-size: 14px; color: #999;">验证码有效期为5分钟，请勿泄露给他人。</p>
			<p style="font-size: 14px; color: #999;">如果您没有请求此验证码，请忽略此邮件。</p>
			<div style="margin-top: 30px; padding-top: 20px; border-top: 1px solid #e0e0e0; text-align: center; color: #999; font-size: 12px;">
				此邮件由系统自动发送，请勿回复。
			</div>
		</div>
	`, code)

	return SendEmail(to, subject, htmlBody)
}

// GenerateVerificationCode 生成6位随机验证码
func GenerateVerificationCode() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	code := r.Intn(900000) + 100000 // 生成100000-999999之间的随机数
	return strconv.Itoa(code)
}

// buildEmail 构建邮件内容
func buildEmail(to, subject, body string) []byte {
	// 获取邮箱配置
	emailConfig := config.GetConfig().Email

	// 邮件头
	headers := make(map[string]string)
	headers["From"] = emailConfig.Username
	headers["To"] = to
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/plain; charset=\"utf-8\""
	headers["Content-Transfer-Encoding"] = "base64"

	// 构建邮件头
	message := ""
	for key, value := range headers {
		message += fmt.Sprintf("%s: %s\r\n", key, value)
	}
	message += "\r\n" + body

	return []byte(message)
}

// sendEmail 发送邮件
func sendEmail(to string, message []byte) error {
	// 获取邮箱配置
	emailConfig := config.GetConfig().Email

	// SMTP服务器
	smtpServer := fmt.Sprintf("%s:%d", emailConfig.Host, emailConfig.Port)

	// SMTP身份验证
	auth := smtp.PlainAuth(
		"",
		emailConfig.Username,
		emailConfig.Password,
		emailConfig.Host,
	)

	// 发送邮件
	err := smtp.SendMail(
		smtpServer,
		auth,
		emailConfig.Username,
		[]string{to},
		message,
	)

	return err
}
