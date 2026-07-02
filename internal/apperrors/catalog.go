package apperrors

type MessageCatalog func(code Code) string

var catalog MessageCatalog = messageEN

func SetCatalog(c MessageCatalog) {
	catalog = c
}

func Message(code Code) string {
	return catalog(code)
}
