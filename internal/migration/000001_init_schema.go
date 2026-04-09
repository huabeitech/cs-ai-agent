package migration

func init() {
	register(1, "init schema migration", func() error {
		return nil
	})
}
