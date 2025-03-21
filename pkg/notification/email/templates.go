package email

// otpEmailTemplate is the HTML template for OTP emails
const otpEmailTemplate = `<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Verification Code</title>
  <style>
    body {
      font-family: Arial, sans-serif;
      line-height: 1.6;
      color: #333;
      margin: 0;
      padding: 0;
    }
    .container {
      max-width: 600px;
      margin: 0 auto;
      padding: 20px;
    }
    .header {
      text-align: center;
      padding: 20px 0;
      border-bottom: 1px solid #eee;
    }
    .content {
      padding: 20px 0;
    }
    .otp-code {
      font-size: 28px;
      font-weight: bold;
      text-align: center;
      margin: 30px 0;
      letter-spacing: 3px;
      color: #4a154b;
    }
    .footer {
      text-align: center;
      font-size: 12px;
      color: #888;
      border-top: 1px solid #eee;
      padding-top: 20px;
    }
  </style>
</head>
<body>
  <div class="container">
    <div class="header">
      <h1>Qubool Kallyaanam</h1>
    </div>
    <div class="content">
      <p>Hello {{.Name}},</p>
      <p>Your verification code for Qubool Kallyaanam is:</p>
      <div class="otp-code">{{.OTP}}</div>
      <p>This code will expire in 24 hours. Please do not share this code with anyone.</p>
      <p>If you did not request this code, please ignore this email.</p>
      <p>Thank you,<br>Qubool Kallyaanam Team</p>
    </div>
    <div class="footer">
      <p>&copy; 2023 Qubool Kallyaanam. All rights reserved.</p>
      <p>This is an automated message, please do not reply to this email.</p>
    </div>
  </div>
</body>
</html>`
