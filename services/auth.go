package services

// func AuthenticateUser(
// 	ctx context.Context,
// 	email string,
// 	providedPassword string,
// 	user models.UserEntity,
// ) (models.UserEntity, error) {
// 	if isVerified := user.IsVerified(); !isVerified {
// 		return models.UserEntity{}, ErrEmailNotValidated
// 	}
//
// 	if err := validatePassword(providedPassword); err != nil {
// 		return models.UserEntity{}, ErrPasswordNotMatch
// 	}
//
// 	return user, nil
// }

// func (a Auth) RegisterUser(
//
//	ctx context.Context,
//	name string,
//	email string,
//	password string,
//	confirmPassword string,
//
//	) error {
//		tx, err := a.db.BeginTx(ctx)
//		if err != nil {
//			return errors.Join(ErrUnrecoverable, err)
//		}
//
//		user, err := models.NewUser(ctx, models.NewUserPayload{
//			Email:    email,
//			Password: password,
//		}, tx, HashAndPepperPassword)
//		if err != nil {
//			if !errors.Is(err, models.ErrDomainValidation) {
//				return errors.Join(ErrUnrecoverable, err)
//			}
//
//			return err
//		}
//
//		return nil
//	}
// func VerifyUserEmail(
// 	ctx context.Context,
// 	token string,
// 	scope string,
// ) error {
// 	tx, err := a.db.BeginTx(ctx)
// 	if err != nil {
// 		return errors.Join(ErrUnrecoverable, err)
// 	}
// 	defer tx.Rollback(ctx)
//
// 	tkn, err := models.GetToken(ctx, token, tx)
// 	if err != nil {
// 		if errors.Is(err, pgx.ErrNoRows) {
// 			return ErrTokenNotExist
// 		}
//
// 		return errors.Join(ErrUnrecoverable, err)
// 	}
//
// 	if tkn.Meta.Scope != models.ScopeEmailVerification {
// 		return ErrTokenScopeInvalid
// 	}
//
// 	if !tkn.IsValid() {
// 		return ErrTokenExpired
// 	}
//
// 	user, err := models.GetUser(ctx, tkn.Meta.ResourceID, tx)
// 	if err != nil {
// 		return errors.Join(ErrUnrecoverable, err)
// 	}
//
// 	if _, err := models.UpdateUser(ctx, models.UpdateUserPayload{
// 		ID:             user.ID,
// 		UpdatedAt:      time.Now(),
// 		Email:          user.Email,
// 		EmailUpdatedAt: time.Now(),
// 	}, tx); err != nil {
// 		return err
// 	}
//
// 	// if _, err := models.DeleteToken(); err != nil {
// 	// 	return err
// 	// }
//
// 	if err := tx.Commit(ctx); err != nil {
// 		return errors.Join(ErrUnrecoverable, err)
// 	}
//
// 	return nil
// }
