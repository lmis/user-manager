package auth

// type LogoutTO struct {
// 	ForgetDevice bool `json:"forgetDevice"`
// }

// func PostLogout(requestContext *ginext.RequestContext, request LogoutTO, c *gin.Context) error {
// 	securityLog := requestContext.SecurityLog
// 	tx := requestContext.Tx
// 	authentication := requestContext.Authentication

// 	securityLog.Info("Logout")

// 	session_service.RemoveSessionCookie(c, models.UserSessionTypeLOGIN)
// 	if authentication != nil && authentication.UserSession.IsNotNil() {
// 		if err := db.ExecSingleMutation(func(ctx context.Context) (int64, error) { return authentication.UserSession.Val.Delete(ctx, tx) }); err != nil {
// 			return util.Wrap("issue while deleting login session", err)
// 		}
// 	}

// 	sudoSessionId, err := session_service.GetSessionCookie(c, models.UserSessionTypeSUDO)
// 	if err != nil {
// 		return util.Wrap("issue reading sudo session cookie", err)
// 	}
// 	if sudoSessionId.IsNotNil() {
// 		session_service.RemoveSessionCookie(c, models.UserSessionTypeSUDO)
// 		sudoSession, err := session_service.FetchSessionAndUser(requestContext, *sudoSessionId.Val, models.UserSessionTypeSUDO)
// 		if err != nil {
// 			return util.Wrap("issue getting sudo session", err)
// 		}
// 		if err := db.ExecSingleMutation(func(ctx context.Context) (int64, error) { return sudoSession.Delete(ctx, tx) }); err != nil {
// 			return util.Wrap("issue while deleting sudo session", err)
// 		}
// 	}
// 	if request.ForgetDevice {
// 		deviceSessionId, err := session_service.GetSessionCookie(c, models.UserSessionTypeREMEMBER_DEVICE)
// 		if err != nil {
// 			return util.Wrap("issue reading device session cookie", err)
// 		}
// 		if deviceSessionId.IsNotNil() {
// 			session_service.RemoveSessionCookie(c, models.UserSessionTypeREMEMBER_DEVICE)
// 			deviceSession, err := session_service.FetchSessionAndUser(requestContext, *deviceSessionId.Val, models.UserSessionTypeREMEMBER_DEVICE)
// 			if err != nil {
// 				return util.Wrap("issue getting device session", err)
// 			}
// 			if err := db.ExecSingleMutation(func(ctx context.Context) (int64, error) { return deviceSession.Delete(ctx, tx) }); err != nil {
// 				return util.Wrap("issue while deleting device session", err)
// 			}
// 		}
// 	}

// 	return nil
// }
