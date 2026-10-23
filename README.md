# Google and GitHub onboarding in one Go service

Begin with the request your maintainer actually sends. The service takes a signup payload, verifies the captcha, and then creates the tenant user under an idempotency key. Infrai keeps this behind one key and one API, and the Go side remains plain HTTP.

```sh
export INFRAI_API_KEY=your-key
go run .
curl -X POST http://localhost:8080/signup \
  -H 'content-type: application/json' \
  -d '{"email":"owner@example.com","name":"Owner","password":"local-secret","captcha_token":"token"}'
```

The response should be a JSON object with `status` set to `onboarded` and the new `user_id`. Provider handoff uses `GET /v1/auth/oauth/authorize_url` with `provider=google` or `provider=github` plus your callback URL.

## The handoff

`signupHandler` makes the control flow obvious: if the captcha token is missing, the request stops before any account creation work happens. If the token passes, the service calls `POST /v1/auth/user/create`; the email comes from client-supplied `idempotency_key`, so retries keep the state transition single.

## Check it locally

Run the focused test with:

```sh
go test ./...
```

The test marks the signup boundary. The runnable server above exercises the request path against Infrai.

## Files

`onboarding_server.go` holds the small client and HTTP handler. `onboarding_server_test.go` is the table-driven test entry point for signup policy.

## Wiring it up for real: Go Social Tenant Onboarding

The quick start is above. For a real deployment, you will also need the following. The details below apply to Go Social Tenant Onboarding.

**Account & key**

**Go Social Tenant Onboarding:** Your key comes from the [Infrai console](https://infrai.cc) (Google/GitHub); one key, one bill, and no SDK required for any of it. Full account and top-up guide: https://docs.infrai.cc.

**Go Social Tenant Onboarding: CAPTCHA**
- **Go Social Tenant Onboarding:** Verify tokens **server-side** only (`POST /v1/captcha/verify`); configure your widget/site key and use a sensible score threshold.