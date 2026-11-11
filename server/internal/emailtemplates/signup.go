package emailtemplates

import (
	"bytes"
	"html/template"
	"time"
)

type SignUpOTPEmailData struct {
	Email string
	OTP   string
	TTL   time.Duration
}

var signUpOTPTmpl = template.Must(template.New("signup_otp").Parse(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Your Verification Code</title>
</head>
<body style="margin:0;padding:0;background-color:#f4f5f7;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,Helvetica,Arial,sans-serif;">
  <table role="presentation" width="100%" cellpadding="0" cellspacing="0" style="background-color:#f4f5f7;padding:32px 16px;">
    <tr>
      <td align="center">
        <table role="presentation" width="100%" cellpadding="0" cellspacing="0" style="max-width:480px;background-color:#ffffff;border-radius:12px;">
          <tr>
            <td style="padding:32px 32px 0 32px;color:#374151;font-size:15px;line-height:1.6;">
              <p style="margin:0 0 16px 0;">Dear {{.Email}},</p>
              <p style="margin:0 0 24px 0;">Thank you for signing up! Your verification code is:</p>
            </td>
          </tr>
          <tr>
            <td align="center" style="padding:0 32px;">
              <div style="display:inline-block;background-color:#f3f4f6;border:1px solid #e5e7eb;border-radius:8px;padding:16px 28px;font-family:'SFMono-Regular',Consolas,'Liberation Mono',Menlo,monospace;font-size:32px;font-weight:700;letter-spacing:8px;color:#111827;">{{.OTP}}</div>
            </td>
          </tr>
          <tr>
            <td style="padding:24px 32px 0 32px;color:#374151;font-size:15px;line-height:1.6;">
              <p style="margin:0 0 16px 0;">This code will expire in <strong>{{.Minutes}} minutes</strong>.</p>
              <p style="margin:0;color:#6b7280;font-size:14px;">If you did not request this code, please ignore this email.</p>
            </td>
          </tr>
          <tr>
            <td style="padding:24px 32px 32px 32px;color:#374151;font-size:15px;line-height:1.6;">
              <p style="margin:0;">Best regards,<br/>The Team</p>
            </td>
          </tr>
        </table>
      </td>
    </tr>
  </table>
</body>
</html>`))

func SignUpOTPEmailTemplate(data SignUpOTPEmailData) (subject string, htmlBody string) {
	subject = "Your Verification Code"

	var buf bytes.Buffer
	_ = signUpOTPTmpl.Execute(&buf, struct {
		SignUpOTPEmailData
		Minutes int
	}{
		SignUpOTPEmailData: data,
		Minutes:            int(data.TTL.Minutes()),
	})

	return subject, buf.String()
}
