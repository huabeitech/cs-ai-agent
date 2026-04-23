package migration

import (
	"cs-agent/internal/pkg/constants"

	"github.com/mlogclub/simple/sqls"
)

func init() {
	register(4, "sync ai agent permissions", func() error {
		return sqls.WithTransaction(func(ctx *sqls.TxContext) error {
			permissions, err := ensurePermissions(ctx.Tx)
			if err != nil {
				return err
			}

			roles, err := ensureRoles(ctx.Tx)
			if err != nil {
				return err
			}

			if err = ensureRolePermissions(ctx.Tx, roles, permissions); err != nil {
				return err
			}

			return ensureBootstrapAdmin(ctx.Tx, roles[constants.RoleCodeSuperAdmin])
		})
	})
}
