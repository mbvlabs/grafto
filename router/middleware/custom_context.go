package middleware

// func RegisterAppContext(
// 	next echo.HandlerFunc,
// ) echo.HandlerFunc {
// 	return func(c echo.Context) error {
// 		if strings.HasPrefix(c.Request().URL.Path, "/static") ||
// 			strings.HasPrefix(c.Request().URL.Path, "/fragments") {
// 			return next(c)
// 		}
//
// 		sess, err := session.Get(reqmeta.AuthenticatedSessionName, c)
// 		if err != nil {
// 			return err
// 		}
//
// 		isAuth, _ := sess.Values[reqmeta.SessIsAuthenticated].(bool)
// 		userID, _ := sess.Values[reqmeta.SessUserID].(uuid.UUID)
// 		userEmail, _ := sess.Values[reqmeta.SessUserEmail].(string)
// 		isAdmin, _ := sess.Values[reqmeta.SessIsAdmin].(bool)
//
// 		ac := reqmeta.App{
// 			Context:         c,
// 			UserID:          userID,
// 			Email:           userEmail,
// 			IsAuthenticated: isAuth,
// 			IsAdmin:         isAdmin,
// 			CurrentPath:     c.Request().URL.Path,
// 		}
//
// 		c.Set(reqmeta.AppKey{}.String(), ac)
//
// 		return next(c)
// 	}
// }
//
// func RegisterFlashMessagesContext(
// 	next echo.HandlerFunc,
// ) echo.HandlerFunc {
// 	return func(c echo.Context) error {
// 		if strings.HasPrefix(c.Request().URL.Path, "/static") {
// 			return next(c)
// 		}
//
// 		sess, err := session.Get(reqmeta.FlashSessionKey, c)
// 		if err != nil {
// 			return err
// 		}
//
// 		flashMessages := []reqmeta.FlashMessage{}
// 		if flashes := sess.Flashes(reqmeta.FlashSessionKey); len(
// 			flashes,
// 		) > 0 {
// 			for _, flash := range flashes {
// 				if msg, ok := flash.(reqmeta.FlashMessage); ok {
// 					flashMessages = append(
// 						flashMessages,
// 						reqmeta.FlashMessage{
// 							Context:   c,
// 							ID:        msg.ID,
// 							Type:      msg.Type,
// 							CreatedAt: msg.CreatedAt,
// 							Message:   msg.Message,
// 						},
// 					)
// 				}
// 			}
//
// 			if err := sess.Save(c.Request(), c.Response()); err != nil {
// 				return err
// 			}
// 		}
//
// 		c.Set(reqmeta.FlashKey{}.String(), flashMessages)
//
// 		return next(c)
// 	}
// }
