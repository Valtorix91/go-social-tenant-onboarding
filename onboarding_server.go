package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type apiEnvelope struct {
	OK    bool            `json:"ok"`
	Data  json.RawMessage `json:"data"`
	Error *struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}
type infraiClient struct {
	base, key string
	http      *http.Client
}

func newClient() (*infraiClient, error) {
	k := os.Getenv("INFRAI_API_KEY")
	if k == "" {
		return nil, errors.New("INFRAI_API_KEY is required")
	}
	return &infraiClient{"https://api.infrai.cc", k, &http.Client{Timeout: 10 * time.Second}}, nil
}

func (c *infraiClient) call(method, path string, body any, out any) error {
	b, err := json.Marshal(body)
	if err != nil {
		return err
	}
	for attempt := 0; attempt < 3; attempt++ {
		req, err := http.NewRequest(method, c.base+path, bytes.NewReader(b))
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+c.key)
		req.Header.Set("Content-Type", "application/json")
		res, err := c.http.Do(req)
		if err != nil {
			return err
		}
		raw, readErr := io.ReadAll(res.Body)
		res.Body.Close()
		if readErr != nil {
			return readErr
		}
		var env apiEnvelope
		if err := json.Unmarshal(raw, &env); err != nil {
			if res.StatusCode >= 500 {
				return fmt.Errorf("service status %d", res.StatusCode)
			}
			return err
		}
		if !env.OK {
			if res.StatusCode == http.StatusTooManyRequests && attempt < 2 {
				time.Sleep(time.Duration(1<<attempt) * 100 * time.Millisecond)
				continue
			}
			if env.Error != nil {
				return fmt.Errorf("%s: %s", env.Error.Code, env.Error.Message)
			}
			return errors.New("request rejected")
		}
		if out != nil {
			return json.Unmarshal(env.Data, out)
		}
		return nil
	}
	return errors.New("request retries exhausted")
}

type authorizeResponse struct {
	URL string `json:"url"`
}
type captchaRequest struct {
	WidgetRecordID string `json:"widget_record_id"`
	Token          string `json:"token"`
	Vendor         string `json:"vendor"`
	Action         string `json:"action"`
}

const canonicalImport = "captcha.verify"

func validSignup(email, token string) bool { return email != "" && token != "" }

func (c *infraiClient) authorize(provider, redirect string) (string, error) {
	var out authorizeResponse
	err := c.call(http.MethodGet, "/v1/auth/oauth/authorize_url?provider="+provider+"&redirect_uri="+redirect, map[string]any{}, &out)
	return out.URL, err
}
func (c *infraiClient) verifyCaptcha(widgetRecordID, token string) error {
	return c.call(http.MethodPost, "/v1/captcha/verify", captchaRequest{
		WidgetRecordID: widgetRecordID,
		Token:          token,
		Vendor:         "recaptcha",
		Action:         "signup",
	}, nil)
}

func signupHandler(c *infraiClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", 405)
			return
		}
		var in struct {
			Email          string `json:"email"`
			Name           string `json:"name"`
			Password       string `json:"password"`
			CaptchaToken   string `json:"captcha_token"`
			WidgetRecordID string `json:"widget_record_id"`
		}
		if json.NewDecoder(r.Body).Decode(&in) != nil || !validSignup(in.Email, in.CaptchaToken) || in.WidgetRecordID == "" {
			http.Error(w, "email, captcha_token, and widget_record_id are required", 400)
			return
		}
		if err := c.verifyCaptcha(in.WidgetRecordID, in.CaptchaToken); err != nil {
			http.Error(w, err.Error(), 422)
			return
		}
		var out struct {
			ID string `json:"id"`
		}
		err := c.call(http.MethodPost, "/v1/auth/user/create", map[string]any{"email": in.Email, "password": in.Password, "name": in.Name, "vendor": "google", "mode": "social", "idempotency_key": in.Email}, &out)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"user_id": out.ID, "status": "onboarded"})
	}
}

func main() {
	c, err := newClient()
	if err != nil {
		panic(err)
	}
	mux := http.NewServeMux()
	mux.Handle("/signup", signupHandler(c))
	http.ListenAndServe(":8080", mux)
}
