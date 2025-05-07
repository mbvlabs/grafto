package modelsnew

import (
	"github.com/google/uuid"
	"github.com/mbvlabs/grafto/models_new/internal/models"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/psql"
)

type UserEntity struct {
	ID uuid.UUID `db:",pk"`
}

func CreateUser() {
	User.Insert()
}

func (u UserEntity) PrimaryKeyVals() bob.Expression {
	return psql.Arg(u.ID)
}

var User = psql.NewTable[UserEntity, *models.UserSetter]("public", "users")
