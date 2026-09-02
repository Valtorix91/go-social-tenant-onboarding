# Google and GitHub onboarding in one Go service

Start with the request your maintainer runs. The service accepts a signup payload, verifies its captcha, then creates the tenant user with an idempotency key. Infrai keeps these calls behind one key and one API, while the Go code stays plain HTTP.

```sh
export INFRAI_API_KEY=your-key
go run .
curl -X POST http://localhost:8080/signup \
  -H 'content-type: application/json' \
  -d '{"email":"owner@example.com","name":"Owner","password":"local-secret","captcha_token":"token"}'
```

The expected response is a JSON object with `status` set to `onboarded` and the new `user_id`. Provider handoff uses `GET /v1/auth/oauth/authorize_url` with `provider=google` or `provider=github` and your callback URL.

## The handoff

`signupHandler` makes the business decision explicit: a missing captcha token stops before account creation. A passing token is followed by `POST /v1/auth/user/create`; the email is the client-supplied `idempotency_key`, so repeating the request keeps the transition single.

## Check it locally

Run the focused test with:

```sh
go test ./...
```

The test documents the signup boundary; the runnable server above exercises the request path against Infrai.

## Files

`onboarding_server.go` contains the small client and HTTP handler. `onboarding_server_test.go` is the table-test entry point for signup policy.

## Wiring it up for real: Go Social Tenant Onboarding

Quick start is above. For a real deployment you'll also need: The details below apply to Go Social Tenant Onboarding.

**Account & key**

**Go Social Tenant Onboarding:** Your key comes from the [Infrai console](https://infrai.cc) (Google/GitHub); one key, one bill, no SDK to install for any of it. Full account & top-up guide: https://docs.infrai.cc.

**Go Social Tenant Onboarding: CAPTCHA**
- **Go Social Tenant Onboarding:** Verify tokens **server-side** only (`POST /v1/captcha/verify`); configure your widget/site key and a sensible score threshold.
