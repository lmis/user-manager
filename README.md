# user-manager
A currently pretty useless backend written with gin and sqlboiler.

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
```golang
type SignUpTO struct {
	UserName string `json:"userName"`
	Language string `json:"language"`
	Email    string `json:"email"`
	Password []byte `json:"password"`
}
```
##### POST   /api/auth/login
Logs in users who do not have 2FA enabled. If 2FA is enabled and the correct username password is submitted, informs user that 2FA is required.
```golang
type LoginTO struct {
	Email    string `json:"email"`
	Password []byte `json:"password"`
}

type LoginResponseStatus string

const (
	LoggedIn             LoginResponseStatus = "logged-in"
	SecondFactorRequired LoginResponseStatus = "second-factor-required"
	InvalidCredentials   LoginResponseStatus = "invalid-credentials"
)
```
##### POST   /api/auth/login-with-second-factor
Logs in users who have 2FA enabled
```golang
type LoginWithSecondFactorTO struct {
	LoginTO
	SecondFactor   string `json:"secondFactor"`
	RememberDevice bool   `json:"rememberDevice"`
}

type LoginWithSecondFactorResponseTO struct {
	LoggedIn     bool      `json:"loggedIn"`
	TimeoutUntil time.Time `json:"timeoutUntil,omitempty"`
}
```
##### POST   /api/auth/logout
Logs out users and deletes their sessions
```golang
type LogoutTO struct {
	ForgetDevice bool `json:"forgetDevice"`
}
```
##### POST   /api/auth/request-password-reset
Triggers a password reset email
```golang
type PasswordResetRequestTO struct {
	Email string `json:"email"`
}
```
##### POST   /api/auth/reset-password
Resets password when given the token from the reset email
```golang
type ResetPasswordTO struct {
	Email       string `json:"email"`
	Token       string `json:"token"`
	NewPassword []byte `json:"newPassword"`
}

type ResetPasswordStatus string

const (
	Success      ResetPasswordStatus = "success"
	InvalidToken ResetPasswordStatus = "invalid-token"
)

type ResetPasswordResponseTO struct {
	Status ResetPasswordStatus `json:"status"`
}
```

##### GET    /api/user-info
Returns information on the current user and their roles
```golang
type UserInfoTO struct {
	Roles         []domain_model.UserRole   `json:"roles"`
	EmailVerified bool                `json:"emailVerified"`
	Language      domain_model.UserLanguage `json:"language"`
}
```
##### POST   /api/user/confirm-email
Accepts an email confirmation token
```golang
type EmailConfirmationTO struct {
	Token string `json:"token"`
}

type EmailConfirmationStatus string

const (
	AlreadyConfirmed EmailConfirmationStatus = "already-confirmed"
	NewlyConfirmed   EmailConfirmationStatus = "newly-confirmed"
	InvalidToken     EmailConfirmationStatus = "invalid-token"
)

type EmailConfirmationResponseTO struct {
	Status EmailConfirmationStatus `json:"status"`
}
```
##### POST   /api/user/re-trigger-confirmation-email
Re-sends email confirmation token
```golang
type RetriggerConfirmationEmailResponseTO struct {
	Sent bool `json:"sent"`
}
```

##### POST   /api/user/settings/language
Allows users to set their language choice
```golang
type LanguageTO struct {
	Language domain_model.UserLanguage `json:"language"`
}
```
##### POST   /api/user/settings/sudo
Creates a short-lived 'sudo' session which allows adjusting sensitive settings
```golang
type SudoTO struct {
	Password []byte `json:"password"`
}

type SudoResponseTO struct {
	Success bool `json:"success"`
}
```
##### POST   /api/user/settings/confirm-email-change
Completes an email change flow, by accepting the token sent to the new email
```golang
type EmailChangeConfirmationTO struct {
	Token string `json:"token"`
}

type EmailChangeStatus string

const (
	NoEmailChangeInProgress EmailChangeStatus = "no-change-in-progress"
	InvalidToken            EmailChangeStatus = "invalid-token"
	NewEmailConfirmed       EmailChangeStatus = "new-email-confirmed"
)

type EmailChangeConfirmationResponseTO struct {
	Status EmailChangeStatus `json:"status"`
	Email  string            `json:"email"`
}
```
##### POST   /api/user/settings/sensitive/change-email
Allows user to set a new email for login and communcation
```golang
type ChangeEmailTO struct {
	NewEmail string `json:"newEmail"`
}
```

#### Future endpoints
##### POST   /api/user/settings/generate-temporary-2fa
Generates and stores a secret to be used when users enable 2FA
##### POST   /api/user/settings/sensitive/change-password
Allows users to change their password
##### POST   /api/user/settings/sensitive/2fa
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


