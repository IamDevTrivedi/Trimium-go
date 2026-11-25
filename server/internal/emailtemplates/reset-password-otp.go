package emailtemplates

import (
	"bytes"
	"html/template"
	"time"
)

type ResetPasswordOTPData struct {
	Email    string
	OTP      string
	TTL      time.Duration
	IP       string
	Location string
	UA       string
}

func (d ResetPasswordOTPData) ExpiresInMinutes() int {
	return int(d.TTL.Minutes())
}

var resetPasswordOTPTemplate = template.Must(template.New("reset-password-otp").Parse(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Reset your password</title>
</head>
<body style="margin:0;padding:0;background:#f4f5f7;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,Helvetica,Arial,sans-serif;color:#111827;">
  <table role="presentation" width="100%" cellpadding="0" cellspacing="0" style="padding:40px 16px;">
    <tr>
      <td align="center">
        <table role="presentation" width="100%" cellpadding="0" cellspacing="0" style="max-width:500px;background:#ffffff;border:1px solid #e5e7eb;border-radius:10px;padding:36px;box-shadow:0 1px 3px rgba(0,0,0,0.05);">
          <tr>
            <td>
              <div style="font-size:20px;font-weight:700;color:#111827;letter-spacing:-0.5px;margin-bottom:24px;">Trimium</div>
              <h1 style="margin:0 0 12px;font-size:22px;font-weight:600;color:#111827;">Reset your password</h1>
              <p style="margin:0 0 20px;font-size:15px;line-height:1.6;color:#4b5563;">
                Hi <strong style="color:#111827;">{{.Email}}</strong>, we received a request to reset your password. Use the verification code below:
              </p>
              
              <div style="margin:0 0 24px;padding:18px;background:#f9fafb;border:1px solid #e5e7eb;border-radius:8px;text-align:center;">
                <span style="font-size:32px;font-weight:700;letter-spacing:10px;color:#111827;font-family:Courier,monospace;">{{.OTP}}</span>
              </div>
              
              <p style="margin:0 0 24px;font-size:14px;color:#6b7280;line-height:1.5;">
                This code will expire in <strong>{{.ExpiresInMinutes}} minutes</strong>. Never share this code with anyone.
              </p>

              {{if or .IP .Location .UA}}
              <div style="margin:24px 0;padding:16px;background:#f9fafb;border-radius:8px;border:1px solid #f3f4f6;">
                <p style="margin:0 0 10px;font-size:12px;font-weight:600;color:#6b7280;text-transform:uppercase;letter-spacing:0.5px;">Security Details</p>
                <table role="presentation" width="100%" cellpadding="0" cellspacing="0" style="font-size:13px;color:#374151;line-height:1.6;">
                  {{if .IP}}
                  <tr>
                    <td style="padding:2px 0;color:#6b7280;width:90px;">IP Address:</td>
                    <td style="padding:2px 0;font-weight:500;">{{.IP}}</td>
                  </tr>
                  {{end}}
                  {{if .Location}}
                  <tr>
                    <td style="padding:2px 0;color:#6b7280;">Location:</td>
                    <td style="padding:2px 0;font-weight:500;">{{.Location}}</td>
                  </tr>
                  {{end}}
                  {{if .UA}}
                  <tr>
                    <td style="padding:2px 0;color:#6b7280;">Device:</td>
                    <td style="padding:2px 0;font-weight:500;">{{.UA}}</td>
                  </tr>
                  {{end}}
                </table>
              </div>
              {{end}}

              <p style="margin:24px 0 0;font-size:13px;line-height:1.5;color:#9ca3af;border-top:1px solid #f3f4f6;padding-top:16px;">
                If you did not request a password reset, please secure your account immediately or ignore this email.
              </p>
            </td>
          </tr>
        </table>
      </td>
    </tr>
  </table>
</body>
</html>`))

func ResetPasswordOTPEmail(data ResetPasswordOTPData) (subject, htmlBody string, err error) {
	var buf bytes.Buffer
	if err := resetPasswordOTPTemplate.Execute(&buf, data); err != nil {
		return "", "", err
	}

	return "Your Trimium password reset code", buf.String(), nil
}
