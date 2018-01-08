package distributor

type deploy struct {
	ID         int64
	Server     string
	CreateTime string `db_default:"now()"`
}
