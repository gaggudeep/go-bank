package mail

import (
	"github.com/gaggudeep/bank_go/util"
	"github.com/stretchr/testify/require"
	"testing"
)

// todo: fix this test which is breaking due to gmail auth
func TestSendEmailWithGmail(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	config, err := util.LoadConfig("..")
	require.NoError(t, err)

	sender := NewGmailEmailSender(config.EMAIL_SENDER_NAME,
		config.EMAIL_SENDER_ADDRESS, config.EMAIL_SENDER_PASSWORD)
	subject := "test email"
	content := `
	<h1>Hello world</h1>
	<p>This is a test message from <a href="https://github.com/gaggudeep">Gagan</a></p>
	`
	to := []string{"gagan121099@gmail.com"}
	attachFiles := []string{"../README.md"}

	err = sender.SendEmail(subject, content, to, nil, nil, attachFiles)
	require.NoError(t, err)
}
