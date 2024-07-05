package main

import (
	"github.com/magefile/mage/mg"
)

type FunctionalTests mg.Namespace

//// Basic runs a basic set of functional tests.
//func (FunctionalTests) Basic() error {
//	mg.Deps(ComposeUpLocalEnvironment)
//	testUser := helper.TestUser{
//		AppURL:     appEnv["APP_URL"],
//		MockApiURL: "http://localhost:8081",
//		Email:      "test-user-" + random.MakeRandomURLSafeB64(5) + "@example.com",
//		Password:   "hunter12",
//	}
//	log.Print("Test user email: " + testUser.Email)
//	log.Print("Testing signup...")
//	if err := functionaltests.TestSignUp(&testUser); err != nil {
//		return errs.Wrap("sign-up test failed", err)
//	}
//	log.Print("Testing password reset...")
//	if err := functionaltests.TestPasswordReset(&testUser); err != nil {
//		return errs.Wrap("password reset test failed", err)
//	}
//	log.Print("Testing CSRF...")
//	if err := functionaltests.TestCallWithMismatchingCsrfTokens(&testUser); err != nil {
//		return errs.Wrap("csrf test failed", err)
//	}
//	log.Print("Testing simple login...")
//	if err := functionaltests.TestSimpleLogin(&testUser); err != nil {
//		return errs.Wrap("simple login test failed", err)
//	}
//	log.Print("Testing change password...")
//	if err := functionaltests.TestChangePassword(&testUser); err != nil {
//		return errs.Wrap("change password test failed", err)
//	}
//	log.Print("Success!")
//	return nil
//}
//
//// All runs all functional tests.
//func (FunctionalTests) All() {
//	mg.Deps(FunctionalTests.Basic)
//}
