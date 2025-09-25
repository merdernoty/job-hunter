package telegram

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/merdernoty/job-hunter/config"
)

type TelegramUser struct {
	ID           int64  `json:"id"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name,omitempty"`
	Username     string `json:"username,omitempty"`
	PhotoURL     string `json:"photo_url,omitempty"`
	LanguageCode string `json:"language_code,omitempty"`
}

type WebAppInitData struct {
	QueryID      string       `json:"query_id,omitempty"`
	User         TelegramUser `json:"user,omitempty"`
	Receiver     TelegramUser `json:"receiver,omitempty"`
	Chat         interface{}  `json:"chat,omitempty"`
	ChatType     string       `json:"chat_type,omitempty"`
	ChatInstance string       `json:"chat_instance,omitempty"`
	StartParam   string       `json:"start_param,omitempty"`
	CanSendAfter int          `json:"can_send_after,omitempty"`
	AuthDate     int64        `json:"auth_date"`
	Hash         string       `json:"hash"`
}

type TelegramAuth struct {
	botToken string
}

func NewTelegramAuth(cfg *config.Config) *TelegramAuth {
	return &TelegramAuth{
		botToken: cfg.Bot.Token,
	}
}

func (ta *TelegramAuth) ValidateWebAppData(initData string) (*WebAppInitData, error) {
	if initData == "" {
		return nil, fmt.Errorf("initData is empty")
	}

	values, err := url.ParseQuery(initData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse init data: %w", err)
	}

	hash := values.Get("hash")
	if hash == "" {
		return nil, fmt.Errorf("hash parameter is missing")
	}

	values.Del("hash")

	var pairs []string
	for key, vals := range values {
		if len(vals) > 0 && vals[0] != "" {
			pairs = append(pairs, fmt.Sprintf("%s=%s", key, vals[0]))
		}
	}
	sort.Strings(pairs)
	dataCheckString := strings.Join(pairs, "\n")

	secretKey := hmac.New(sha256.New, []byte("WebAppData"))
	secretKey.Write([]byte(ta.botToken))

	signature := hmac.New(sha256.New, secretKey.Sum(nil))
	signature.Write([]byte(dataCheckString))
	expectedHash := hex.EncodeToString(signature.Sum(nil))

	if hash != expectedHash {
		return nil, fmt.Errorf("invalid hash signature")
	}

	authDateStr := values.Get("auth_date")
	if authDateStr == "" {
		return nil, fmt.Errorf("auth_date is missing")
	}

	authDate, err := strconv.ParseInt(authDateStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid auth_date format")
	}

	if time.Now().Unix()-authDate > 86400 { 
		return nil, fmt.Errorf("auth_date is too old")
	}
	var webAppData WebAppInitData
	webAppData.AuthDate = authDate
	webAppData.Hash = hash

	if userStr := values.Get("user"); userStr != "" {
		if err := json.Unmarshal([]byte(userStr), &webAppData.User); err != nil {
			return nil, fmt.Errorf("failed to parse user data: %w", err)
		}
	}

	if queryID := values.Get("query_id"); queryID != "" {
		webAppData.QueryID = queryID
	}

	if startParam := values.Get("start_param"); startParam != "" {
		webAppData.StartParam = startParam
	}

	return &webAppData, nil
}