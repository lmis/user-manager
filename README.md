# user-manager
A currently pretty useless backend written with gin and go-jet.

## Build and run
1. Make sure you have docker daemon running
2. Run `check`
3. Run `local-dev-startup`

### Run endpoint tests
To run the endpoint tests, run `run-tests` while a local instance is running.

## API

#### Current Endpoints
##### POST   /api/auth/sign-up
Accepts sign-up data and creates a user account
```typescript
{
  userName: string,
  language: string,
  email:    string,
  password: string /* base64 encoded byte array */
}
```
##### POST   /api/auth/login
Logs in users who do not have 2FA enabled. If 2FA is enabled and the correct username password is submitted, informs user that 2FA is required.
```typescript
{
  email:    string,
  password: string /* base64 encoded byte array */
}

{ 
  status:  "logged-in" | "second-factor-required" | "invalid-credentials" 
}
```
##### POST   /api/auth/login-with-second-factor
Logs in users who have 2FA enabled
```typescript
{
  email:    string,
  password: string /* base64 encoded byte array */
  secondFactor:   string,
  rememberDevice: bool,
}

{
  loggedIn:     bool,
  timeoutUntil: number, /* unix timestamp */
}
```
##### POST   /api/auth/logout
Logs out users and deletes their sessions
```typescript
{
  forgetDevice: bool,
}
```
##### POST   /api/auth/request-password-reset
Triggers a password reset email
```typescript
{
  email: string,
}
```
##### POST   /api/auth/reset-password
Resets password when given the token from the reset email
```typescript
{
  email:       string,
  token:       string,
  password:    string /* base64 encoded byte array */
}


{
  status:  "success" | "invalid-token"
}
```

##### GET    /api/user-info
Returns information on the current user and their roles
```typescript
{
  roles:         string[],
  emailVerified: bool,
  language:      "DE" | "EN",
}
```
##### POST   /api/user/confirm-email
Accepts an email confirmation token
```typescript
{
  token: string
}

{
  status:  "already-confirmed" | "newly-confirmed" | "invalid-token"
}
```
##### POST   /api/user/re-trigger-confirmation-email
Re-sends email confirmation token
```typescript
{
  sent: bool
}
```

##### POST   /api/user/settings/language
Allows users to set their language choice
```typescript
{
  language: "DE" | "EN"
}
```
##### POST   /api/user/settings/sudo
Creates a short-lived 'sudo' session which allows adjusting sensitive settings
```typescript
{
  password:    string /* base64 encoded byte array */
}

{
  success: bool
}
```
##### POST   /api/user/settings/confirm-email-change
Completes an email change flow, by accepting the token sent to the new email
```typescript
{
  token: string
}

{
  status: "no-change-in-progress" | "invalid-token" | "new-email-confirmed"
  email:  string
}
```
##### POST   /api/user/settings/sensitive/initiate-email-change
Allows user to set a new email for login and communcation
```typescript
{
  newEmail: string
}
```

#### Future endpoints
##### POST   /api/user/settings/generate-temporary-second-factor-token
Generates and stores a secret to be used when users enable 2FA
##### POST   /api/user/settings/sensitive/change-password
Allows users to change their password
##### POST   /api/user/settings/sensitive/second-factor
Enable / Disable 2FA login

### Session cookies
The following HTTP-only cookies may be set by the application `LOGIN_TOKEN`, `DEVICE_TOKEN`, `SUDO_TOKEN`
### CSRF Protection
An application must send the same value both in the cookie `CSRF-Token` (local environment) or `__Host-CSRF-Token` as well as on the `X-CSRF-Token` header

### Roles
Access to the endpoint `GET /api/user` requires no authentication but may return an empty struct.
Access to endpoints under `/api/user/` requires a valid `LOGIN_TOKEN` (obtained via login endpoint with correct credentials).
Access to endpoints under `/api/user/settings` require the user's email adress to be verified.
Access to endpoints under `/api/user/settings/sensitive` require a valid `SUDO_TOKEN` obtained via the sudo enpoint.
