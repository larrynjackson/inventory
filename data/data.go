package data

type StorageReader interface {
	SelectUser(userId string) map[string]string
	SelectTDkeyList() []string
	SelectItems(tKey string) []map[string]string
	SelectUserdKeyList(tKey string) []string
	SelectAllUsers() []string
}

type StorageWriter interface {
	Open()

	InsertUser(userId, tKey, password string)
	UpdateUser(userId, password string)
	InsertItem(tKey, dKey, name, description, value, purchasedate, serialnum string)
	UpdateItem(dKey, name, description, value, purchasedate, serialnum string)
	DeleteUser(userId string)
	DeleteItems(tKey string)
	DeleteItem(dKey string)
}

type StorageReadWrite interface {
	StorageReader
	StorageWriter
}
